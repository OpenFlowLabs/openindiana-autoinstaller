package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"

	"fmt"

	"github.com/OpenFlowLabs/openindiana-autoinstaller/installservd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AddAssetCommand = &cobra.Command{
	Use:   "add-asset",
	Short: "add an asset to the Installation Server",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		src := args[1]
		rpcArgs := installservd.AddAssetArg{}
		if strings.Contains(src, "http") {
			rpcArgs.Source = src
		} else {
			var srcFile *os.File
			var err error
			if srcFile, err = os.Open(src); err != nil {
				return fmt.Errorf("could not read asset: %s", err)
			}
			defer srcFile.Close()
			buff := bytes.NewBuffer(rpcArgs.Content)
			if _, err = io.Copy(buff, srcFile); err != nil {
				return fmt.Errorf("could not read asset: %s", err)
			}
			rpcArgs.Content = buff.Bytes()
		}

		rpcArgs.Path = target
		rpcArgs.Type = viper.GetString("type")

		var reply string
		if err := rpcDialServer(viper.GetString("socket"), "InstallservdRPCReceiver.AddAsset", rpcArgs, &reply); err != nil {
			return fmt.Errorf("could not conntact installservd: %s", err)
		}
		fmt.Println(reply)

		return nil
	},
}

func init() {
	AddAssetCommand.Flags().StringP("type", "t", "", "Provide a type to make additional PostProcessing. Use \"template\" for templates and \"image\" for a known image format (known are iso and aci)")
	viper.BindPFlags(AddAssetCommand.Flags())
	RootCmd.AddCommand(AddAssetCommand)
}
