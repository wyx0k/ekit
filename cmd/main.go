package main

import (
	"fmt"
	"os"
	"path/filepath"
	"wyx0k/ekit/app"
	"wyx0k/ekit/log"

	"github.com/spf13/cobra"
)

var appName = "demo"

func main() {
	var (
		configFile string
	)
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: appName,
		Run: func(cmd *cobra.Command, args []string) {
			configFile, err := filepath.Abs(configFile)
			if err != nil {
				fmt.Println(err.Error())
			}
			demo := app.App(appName)

			demo.WithTitle("-> demo <-")
			demo.WithParam("configPath", configFile)
			demo.WithConfigLoader(app.NewFileConfigLoader(configFile))
			demo.WithLogger(log.WithSimpleLogger())
			os.Exit(demo.Start())
		},
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./config.yaml", "config")
	rootCmd.Execute()

}
