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

	"github.com/OpenFlowLabs/openindiana-autoinstaller/devprop"
	"github.com/OpenFlowLabs/openindiana-autoinstaller/fileutils"
	"github.com/OpenFlowLabs/openindiana-autoinstaller/installd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "installd",
	Short: "Install A System Based on a Command File",
	Long: `This Daemon will install a system based on the Configuration passed in the commandline or via devprop
	If executed without arguments it will start a instance of itself in daemon mode
	Use -i to stay in foreground
	`,
	PersistentPreRunE: preRun,
	RunE: func(cmd *cobra.Command, args []string) error {
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
				return fmt.Errorf("webserver functionality not implemented yet")
			}
			if err := runInstall(configFileName, noop); err != nil {
				return fmt.Errorf("error could not perform installation: %e", err)
			}

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
					return fmt.Errorf("could not start daemon: %e", err)
				}
			} else {
				return fmt.Errorf("could not launch Daemon: %e", err)
			}
		}

		return nil
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

func preRun(_ *cobra.Command, _ []string) error {
	loglevel := viper.GetString("loglevel")
	debug := viper.GetBool("debug")
	config := viper.GetString("config")

	if loglevel != "" {
		switch strings.ToLower(loglevel) {
		case "trace":
			logrus.SetLevel(logrus.TraceLevel)
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
		case "info":
			logrus.SetLevel(logrus.InfoLevel)
		case "warn":
			logrus.SetLevel(logrus.WarnLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if config != "" {
		viper.SetConfigFile(config)
	}

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return fmt.Errorf("fatal error config file: %e", err)
	}

	return nil
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
	var downloaded = false
	if strings.HasPrefix(configLocation, "http") || strings.HasPrefix(configLocation, "https") {
		var dlErr error
		if configLocation, dlErr = fileutils.HTTPDownloadTo(configLocation, "/tmp"); dlErr != nil {
			return dlErr
		}
		downloaded = true
	} else if strings.HasPrefix(configLocation, "nfs") {
		return fmt.Errorf("not Supported URL Type NFS")
	}
	if downloaded {
		_, configFileName := path.Split(configLocation)
		configLocation = filepath.Join("/tmp", configFileName)
	}

	if file, osReadErr = ioutil.ReadFile(configLocation); osReadErr != nil {
		return osReadErr
	}
	if err := json.Unmarshal(file, &confObj); err != nil {
		return err
	}
	confObj.FillUnSetValues()
	return installd.Install(confObj, noop)
}
