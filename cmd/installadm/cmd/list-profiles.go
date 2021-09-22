package cmd

import (
	"fmt"
	"os"

	"github.com/OpenFlowLabs/openindiana-autoinstaller/installservd"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ListProfileCommand = &cobra.Command{
	Use:     "list-profiles",
	Aliases: []string{"lsp"},
	Short:   "list profiles registered in the Installation Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		var profiles []installservd.Profile
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.ListProfiles", "", &profiles); err != nil {
			return fmt.Errorf("could not conntact installservd: %s", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "OS", "Kernel", "BootArchive"})
		for _, p := range profiles {
			table.Append([]string{p.Name, string(p.OS), p.Kernel.Path, p.BootArchive.Path})
		}
		table.Render()

		return nil
	},
}

func init() {
	viper.BindPFlags(ListProfileCommand.Flags())
	RootCmd.AddCommand(ListProfileCommand)
}
