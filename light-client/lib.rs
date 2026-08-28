use sha2::{Digest, Sha256};
use std::ffi::{c_char, CStr, CString};
use std::ptr;
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
struct Header {
    chain_id: String,
    height: i64,
    time: String,
    prev_hash: String,
    app_hash: String,
    block_hash: String,
}

fn hex(bytes: &[u8]) -> String { bytes.iter().map(|b| format!("{:02x}", b)).collect() }
fn hash_header(h: &Header) -> String {
    let raw = format!("{}|{}|{}|{}|{}", h.chain_id, h.height, h.time, h.prev_hash, h.app_hash);
    hex(&Sha256::digest(raw.as_bytes()))
}

/// 验证区块头，不访问 RPC，也不接受节点口头声明；调用方必须提供完整头部 JSON。
pub fn verify_header_json(json: &str, trusted_height: i64, trusted_hash: &str) -> Result<bool, String> {
    let h: Header = serde_json::from_str(json).map_err(|e| format!("JSON 无效: {e}"))?;
    if h.chain_id != "bitcoin-local-1" { return Err("chain_id 不匹配".into()); }
    if h.height <= trusted_height { return Err("区块高度未前进".into()); }
    if !h.prev_hash.is_empty() && h.prev_hash != trusted_hash { return Err("前序哈希不匹配".into()); }
    let expected = hash_header(&h);
    if expected != h.block_hash { return Err("区块哈希校验失败".into()); }
    Ok(true)
}

/// 轻客户端测试接口：验证相邻区块链接和确定性区块哈希。
pub fn verify_chain_json(json_blocks: &str) -> Result<usize, String> {
    let blocks: Vec<Header> = serde_json::from_str(json_blocks).map_err(|e| format!("区块列表 JSON 无效: {e}"))?;
    let mut previous = String::new(); let mut height = 0;
    for h in &blocks {
        let expected = hash_header(h);
        if h.chain_id != "bitcoin-local-1" || h.height != height + 1 || (!previous.is_empty() && h.prev_hash != previous) || h.block_hash != expected {
            return Err(format!("第 {} 个区块验证失败", h.height));
        }
        previous = h.block_hash.clone(); height = h.height;
    }
    Ok(blocks.len())
}

/// Kotlin/JNI 可直接调用的 C ABI。返回 1 表示通过，0 表示失败。
#[no_mangle]
pub extern "C" fn bitcoin_verify_header(json: *const c_char, trusted_height: i64, trusted_hash: *const c_char) -> i32 {
    if json.is_null() || trusted_hash.is_null() { return 0; }
    let json = unsafe { CStr::from_ptr(json) }.to_string_lossy();
    let trusted = unsafe { CStr::from_ptr(trusted_hash) }.to_string_lossy();
    verify_header_json(&json, trusted_height, &trusted).map(|_| 1).unwrap_or(0)
}

/// 为 Kotlin 演示返回 UTF-8 错误/成功字符串；调用方需调用 bitcoin_string_free。
#[no_mangle]
pub extern "C" fn bitcoin_verify_header_message(json: *const c_char, trusted_height: i64, trusted_hash: *const c_char) -> *mut c_char {
    if json.is_null() || trusted_hash.is_null() { return ptr::null_mut(); }
    let j = unsafe { CStr::from_ptr(json) }.to_string_lossy(); let t = unsafe { CStr::from_ptr(trusted_hash) }.to_string_lossy();
    let message = match verify_header_json(&j, trusted_height, &t) { Ok(_) => "ok".to_string(), Err(e) => e };
    CString::new(message).unwrap().into_raw()
}

#[no_mangle]
pub extern "C" fn bitcoin_string_free(value: *mut c_char) { if !value.is_null() { unsafe { drop(CString::from_raw(value)); } } }

#[cfg(target_os = "android")]
mod android_jni {
    use super::*;
    use jni::objects::{JClass, JString};
    use jni::JNIEnv;

    #[no_mangle]
    pub extern "system" fn Java_com_example_mymobilechain_NativeLightClient_verifyHeader(
        mut env: JNIEnv, _class: JClass, json: JString, trusted_height: i64, trusted_hash: JString
    ) -> i32 {
        let j = match env.get_string(&json) { Ok(v) => v.to_string_lossy().into_owned(), Err(_) => return 0 };
        let t = match env.get_string(&trusted_hash) { Ok(v) => v.to_string_lossy().into_owned(), Err(_) => return 0 };
        verify_header_json(&j, trusted_height, &t).map(|_| 1).unwrap_or(0)
    }

    #[no_mangle]
    pub extern "system" fn Java_com_example_mymobilechain_NativeLightClient_verifyHeaderMessage(
        mut env: JNIEnv, _class: JClass, json: JString, trusted_height: i64, trusted_hash: JString
    ) -> jni::sys::jstring {
        let j = match env.get_string(&json) { Ok(v) => v.to_string_lossy().into_owned(), Err(_) => return ptr::null_mut() };
        let t = match env.get_string(&trusted_hash) { Ok(v) => v.to_string_lossy().into_owned(), Err(_) => return ptr::null_mut() };
        let msg = match verify_header_json(&j, trusted_height, &t) { Ok(_) => "ok".to_string(), Err(e) => e };
        match env.new_string(msg) { Ok(s) => s.into_raw(), Err(_) => ptr::null_mut() }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    fn block(height: i64, prev: &str, app: &str) -> Header {
        let mut h = Header { chain_id: "bitcoin-local-1".into(), height, time: format!("2026-01-01T00:00:0{height}Z"), prev_hash: prev.into(), app_hash: app.into(), block_hash: String::new() };
        h.block_hash = hash_header(&h); h
    }
    #[test]
    fn verifies_linked_headers() {
        let one = block(1, "", "app1"); let two = block(2, &one.block_hash, "app2");
        let json = serde_json::to_string(&vec![one, two]).unwrap(); assert_eq!(verify_chain_json(&json).unwrap(), 2);
    }
    #[test]
    fn rejects_tampered_header() {
        let mut one = block(1, "", "app1"); one.app_hash = "tampered".into();
        assert!(verify_header_json(&serde_json::to_string(&one).unwrap(), 0, "").is_err());
    }
}
