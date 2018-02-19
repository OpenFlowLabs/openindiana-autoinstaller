package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"path"
	"path/filepath"

	"git.wegmueller.it/opencloud/installer/devprop"
	"git.wegmueller.it/opencloud/installer/installd"
	"git.wegmueller.it/opencloud/opencloud/common"
	"git.wegmueller.it/toasterson/daemon"
	"git.wegmueller.it/toasterson/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Daemon daemon.Daemon

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "installd",
	Short: "Install A System Based on a Command File",
	Long: `This Daemon will install a system based on the Configuration passed in the commandline or via devprop
	If executed without arguments it will start a instance of itself in daemon mode
	Use -i to stay in foreground
	`,
	PersistentPreRun: preRun,
	Run: func(cmd *cobra.Command, args []string) {
		//Run As a Process we do not need to start another instance in the background
		if viper.GetBool("daemon") || viper.GetBool("interactive") {
			// Running in Daemon mode means we look to see if we can grab the config and try to execute what was passed to us.
			// If Config was not passed just launch the HTTP Server and wait.
			var configFileName string
			configArg := viper.GetString("config")
			noop := viper.GetBool("noop")
			if configArg == "" {
				configFileName = devprop.GetValue("install_config")
			} else {
				configFileName = configArg
			}
			if configFileName == "" {
				//TODO Webserver to listen for commands
				os.Exit(0)
			}
			if err := runInstall(configFileName, noop); err != nil {
				common.ExitWithErr("could not perform installation")
			}
			os.Exit(0)
		} else {
			//If we are here it means we assume to have been called by init or by hand and should start the daemon
			if exeName, err := os.Executable(); err == nil {
				confFile := viper.GetString("config")
				noop := viper.GetBool("noop")
				args = []string{"--daemon"}
				if confFile != "" {
					args = append(args, "-c", confFile)
				}
				if noop {
					args = append(args, "-n")
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

func preRun(cmd *cobra.Command, args []string) {
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
		common.ExitWithErr("Fatal error config file: %s", err)
	}
	if viper.GetBool("daemon") {
		if Daemon, err = daemon.New("installd", "illumos Installation Daemon", "svc:/milestone/network:default"); err != nil {
			common.ExitWithErr("could not create new Daemon Instance %s", err)
		}
	}
}

func init() {
	RootCmd.PersistentFlags().StringP("config", "c", "", "The Location of the Install Configuration file. Can be http.")
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable Debuging")
	RootCmd.PersistentFlags().String("loglevel", "", "Set the Log Level")
	RootCmd.PersistentFlags().BoolP("noop", "n", false, "Do nothing at all.")
	RootCmd.PersistentFlags().Bool("daemon", false, "Only report what would be done.")
	RootCmd.PersistentFlags().BoolP("interactive", "i", false, "Run in Foreground")
	viper.BindPFlag("loglevel", RootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("noop", RootCmd.PersistentFlags().Lookup("noop"))
	viper.BindPFlag("daemon", RootCmd.PersistentFlags().Lookup("daemon"))
	viper.BindPFlag("interactive", RootCmd.PersistentFlags().Lookup("interactive"))
}

func runInstall(configLocation string, noop bool) error {
	var file []byte
	var osReadErr error
	var confObj installd.InstallConfiguration
	if strings.HasPrefix(configLocation, "http") || strings.HasPrefix(configLocation, "https") {
		var dlErr error
		if configLocation, dlErr = installd.HTTPDownloadTo(configLocation, "/tmp"); dlErr != nil {
			return dlErr
		}
	} else if strings.HasPrefix(configLocation, "nfs") {
		return fmt.Errorf("not Supported URL Type NFS")
	}
	_, configFileName := path.Split(configLocation)
	if file, osReadErr = ioutil.ReadFile(filepath.Join("/tmp", configFileName)); osReadErr != nil {
		return osReadErr
	}
	if err := json.Unmarshal(file, &confObj); err != nil {
		return err
	}
	//Assume that we want the Media URL from devprop if it is not in the config
	if confObj.InstallImage.URL == "" {
		confObj.InstallImage.URL = devprop.GetValue("install_media")
	}
	return installd.Install(confObj, noop)
}
