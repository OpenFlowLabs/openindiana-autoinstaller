package cmd

import (
	"fmt"
	"os"

	"git.wegmueller.it/opencloud/opencloud/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var PingCommand = &cobra.Command{
	Use:   "ping",
	Short: "ping the server for debugging",
	Run: func(cmd *cobra.Command, args []string) {
		var reply string
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.Ping", "ping", &reply); err != nil {
			common.ExitWithErr("Could not conntact installservd: ", err)
		}
		fmt.Println(reply)
		os.Exit(0)
	},
}

func init() {
	PingCommand.Flags().String("socket", "/var/run/installservd.socket", "The RPC Socket of the installservd instance you want to contact")
	viper.BindPFlags(PingCommand.Flags())
	RootCmd.AddCommand(PingCommand)
}
