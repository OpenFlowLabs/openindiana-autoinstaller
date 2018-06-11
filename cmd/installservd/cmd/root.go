package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"strings"

	"git.wegmueller.it/opencloud/installer/installservd"
	"git.wegmueller.it/opencloud/opencloud/common"
	"git.wegmueller.it/toasterson/daemon"
	"git.wegmueller.it/toasterson/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Daemon daemon.Daemon

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "installservd",
	Short: "Serves Files and Configurations for the Installation Daemon",
	Long: `Ths Daemon Server files and configurations to all installing servers. Or other Software wanting them.
	This Daemon is managed via installadm command. If you want to do anything with this Daemon use installadm.
	`,
	PersistentPreRun: preRun,
	Run: func(cmd *cobra.Command, args []string) {
		//Run As a Process we do not need to start another instance in the background
		if viper.GetBool("daemon") || viper.GetBool("interactive") {
			// Running in Daemon mode means we look to see if we can grab the config and try to execute what was passed to us.
			//configFileName := viper.GetString("config")
			server := installservd.New()
			listen := viper.GetString("listen")
			cert := viper.GetString("cert")
			if err := server.HandleRPC(viper.GetString("socket")); err != nil {
				common.ExitWithErr("Could not open rpc Server: %s", err)
			}
			if cert == "auto" {
				server.Echo.StartAutoTLS(listen)
			} else if cert != "no" {
				keyFile := strings.Replace(cert, ".crt", ".key", -1)
				server.Echo.StartTLS(listen, cert, keyFile)
			} else {
				server.Echo.Start(listen)
			}
			if err := server.StopRPC(); err != nil {
				common.ExitWithErr("Could not close rpc properly: %s", err)
			}
			os.Exit(0)
		} else {
			//If we are here it means we assume to have been called by init or by hand and should start the daemon
			if exeName, err := os.Executable(); err == nil {
				confFile := viper.GetString("config")
				args = []string{"--daemon"}
				if confFile != "" {
					args = append(args, "-c", confFile)
				}
				cmd := exec.Command(exeName, "--daemon")
				if err := cmd.Start(); err != nil {
					common.ExitWithErr("could not start daemon: %s", err)
				}
			} else {
				common.ExitWithErr("Could not launch Daemon: %s", err)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func preRun(_ *cobra.Command, _ []string) {
	loglevel := viper.GetString("loglevel")
	debug := viper.GetBool("debug")
	config := viper.GetString("config")
	if loglevel != "" {
		glog.SetLevelFromString(loglevel)
	}
	if debug {
		glog.SetLevel(glog.LOG_DEBUG)
	}
	if config != "" {
		viper.SetConfigFile(config)
	}
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		glog.Errf("Could not read Config: ", err)
	}
}

func init() {
	RootCmd.PersistentFlags().StringP("config", "c", "", "The Location of the Install Configuration file. Can be http.")
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable Debuging")
	RootCmd.PersistentFlags().String("loglevel", "", "Set the Log Level")
	RootCmd.PersistentFlags().Bool("daemon", false, "Only report what would be done.")
	RootCmd.PersistentFlags().BoolP("interactive", "i", false, "Run in Foreground")
	RootCmd.PersistentFlags().StringP("listen", "l", ":3000", "What Address/Port to listen on.")
	RootCmd.PersistentFlags().StringP("cert", "t", "no", "Which certificate to use, use auto for letsencrypt auto and no to disable. default: no")
	RootCmd.PersistentFlags().StringP("home", "H", "$HOME", "Which directory holds the installation images and other files")
	RootCmd.PersistentFlags().String("socket", "/var/run/installservd.socket", "Socket to listen on for RPC Commands")
	viper.BindPFlag("loglevel", RootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("daemon", RootCmd.PersistentFlags().Lookup("daemon"))
	viper.BindPFlag("interactive", RootCmd.PersistentFlags().Lookup("interactive"))
	viper.BindPFlag("listen", RootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("cert", RootCmd.PersistentFlags().Lookup("cert"))
	viper.BindPFlag("home", RootCmd.PersistentFlags().Lookup("home"))
	viper.BindPFlag("socket", RootCmd.PersistentFlags().Lookup("socket"))
}
