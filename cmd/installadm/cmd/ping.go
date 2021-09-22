package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var PingCommand = &cobra.Command{
	Use:   "ping",
	Short: "ping the server for debugging",
	RunE: func(cmd *cobra.Command, args []string) error {
		var reply string
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.Ping", "ping", &reply); err != nil {
			return fmt.Errorf("could not conntact installservd: %s", err)
		}
		fmt.Println(reply)

		return nil
	},
}

func init() {
	PingCommand.Flags().String("socket", "/var/run/installservd.socket", "The RPC Socket of the installservd instance you want to contact")
	viper.BindPFlags(PingCommand.Flags())
	RootCmd.AddCommand(PingCommand)
}
