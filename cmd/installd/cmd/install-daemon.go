package cmd

import (
	"fmt"
	"os"

	"git.wegmueller.it/opencloud/opencloud/common"
	"git.wegmueller.it/opencloud/opencloud/pod"
	"github.com/appc/spec/schema/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createCmd represents the create command
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
