package main

import (
	"fmt"
	"os"

	clientv2helpers "cosmossdk.io/client/v2/helpers"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/app"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/cmd/catcoind/cmd"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, clientv2helpers.EnvPrefix, app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
