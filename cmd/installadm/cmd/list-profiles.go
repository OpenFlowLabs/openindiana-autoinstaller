package cmd

import (
	"os"

	"git.wegmueller.it/opencloud/installer/installservd"
	"git.wegmueller.it/opencloud/opencloud/common"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ListProfileCommand = &cobra.Command{
	Use:     "list-profiles",
	Aliases: []string{"lsp"},
	Short:   "list profiles registered in the Installation Server",
	Run: func(cmd *cobra.Command, args []string) {
		var profiles []installservd.Profile
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.ListProfiles", "", &profiles); err != nil {
			common.ExitWithErr("Could not conntact installservd: ", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "OS", "Kernel", "BootArchive"})
		for _, p := range profiles {
			table.Append([]string{p.Name, string(p.OS), p.Kernel.Path, p.BootArchive.Path})
		}
		table.Render()
	},
}

func init() {
	viper.BindPFlags(ListProfileCommand.Flags())
	RootCmd.AddCommand(ListProfileCommand)
}
