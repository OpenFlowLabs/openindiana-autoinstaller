package cmd

import (
	"os"

	"fmt"

	"git.wegmueller.it/opencloud/installer/installd"
	"git.wegmueller.it/opencloud/installer/installservd"
	"git.wegmueller.it/opencloud/opencloud/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AddProfileCommand = &cobra.Command{
	Use:   "add-profile",
	Short: "add an profile to the Installation Server",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := installservd.Profile{
			Name:   args[0],
			Config: &installd.InstallConfiguration{},
		}

		if err := editProfileWithEditor(&profile); err != nil {
			common.ExitWithErr("Could not edit Profile in $EDITOR: ", err)
		}

		var reply string
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.AddProfile", profile, &reply); err != nil {
			common.ExitWithErr("Could not conntact installservd: ", err)
		}
		fmt.Println(reply)

		os.Exit(0)
	},
}

func init() {
	viper.BindPFlags(AddProfileCommand.Flags())
	RootCmd.AddCommand(AddProfileCommand)
}
