package cmd

import (
	"fmt"

	"github.com/OpenFlowLabs/openindiana-autoinstaller/installd"
	"github.com/OpenFlowLabs/openindiana-autoinstaller/installservd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AddProfileCommand = &cobra.Command{
	Use:   "add-profile",
	Short: "add an profile to the Installation Server",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := installservd.Profile{
			Name:   args[0],
			Config: &installd.InstallConfiguration{},
		}

		if err := editProfileWithEditor(&profile); err != nil {
			return fmt.Errorf("could not edit Profile in $EDITOR: %s", err)
		}

		var reply string
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.AddProfile", profile, &reply); err != nil {
			return fmt.Errorf("could not conntact installservd: %s", err)
		}
		fmt.Println(reply)

		return nil
	},
}

func init() {
	viper.BindPFlags(AddProfileCommand.Flags())
	RootCmd.AddCommand(AddProfileCommand)
}
