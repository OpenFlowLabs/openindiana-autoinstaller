package cmd

import (
	"github.com/spf13/cobra"
)

var installDaemonCmd = &cobra.Command{
	Use:   "install-daemon",
	Short: "install the daemon into the system",
	Long: `Install this daemon into the system
	`,
	Run:  installCmdRun,
}

func init() {
	RootCmd.AddCommand(installDaemonCmd)
}

func installCmdRun(cmd *cobra.Command, args []string) {

}
