package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "installadm",
	Short: "Command your installation server",
	Long: `This utility commands the Server of the Installation environment. Or Push configs to clients.
	`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringP("config", "c", "", "The Location of the Install Configuration file. Can be http.")
	RootCmd.PersistentFlags().String("socket", "./installservd.socket", "The RPC Socket of the installservd instance you want to contact")
	viper.BindPFlags(RootCmd.PersistentFlags())
}
