package main

import (
	"fmt"
	"github.com/li127103/frp/pkg/config"
	"github.com/li127103/frp/pkg/config/v1/validation"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(verifyCmd)
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify that the configures is valid",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgFile == "" {
			fmt.Println("frps: the configuration file is not specified")
			return nil
		}
		svrCfg, _, err := config.LoadServerConfig(cfgFile, strictConfigMode)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		warning, err := validation.ValidateServerConfig(svrCfg)
		if warning != nil {
			fmt.Printf("WARNING: %v\n", warning)
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("frps: the configuration file %s syntax is ok\n", cfgFile)
		return nil
	},
}
