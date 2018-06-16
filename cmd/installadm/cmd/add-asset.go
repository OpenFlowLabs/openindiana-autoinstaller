package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"

	"fmt"

	"git.wegmueller.it/opencloud/installer/installservd"
	"git.wegmueller.it/opencloud/opencloud/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AddAssetCommand = &cobra.Command{
	Use:   "add-asset",
	Short: "add an asset to the Installation Server",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		src := args[1]
		rpcArgs := installservd.AddAssetArg{}
		if strings.Contains(src, "http") {
			rpcArgs.Source = src
		} else {
			var srcFile *os.File
			var err error
			if srcFile, err = os.Open(src); err != nil {
				common.ExitWithErr("Could not read asset: ", err)
			}
			defer srcFile.Close()
			buff := bytes.NewBuffer(rpcArgs.Content)
			if _, err = io.Copy(buff, srcFile); err != nil {
				common.ExitWithErr("Could not read asset: ", err)
			}
			rpcArgs.Content = buff.Bytes()
		}

		rpcArgs.Path = target
		rpcArgs.Type = viper.GetString("type")

		var reply string
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.AddAsset", rpcArgs, &reply); err != nil {
			common.ExitWithErr("Could not conntact installservd: ", err)
		}
		fmt.Println(reply)

		fmt.Println(reply)
		os.Exit(0)
	},
}

func init() {
	AddAssetCommand.Flags().StringP("type", "t", "", "Provide a type to make additional PostProcessing. Use \"template\" for templates and \"image\" for a known image format (known are iso and aci)")
	viper.BindPFlags(AddAssetCommand.Flags())
	RootCmd.AddCommand(AddAssetCommand)
}
