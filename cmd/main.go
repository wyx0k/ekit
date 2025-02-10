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
		err        error
	)
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: appName,
		Run: func(cmd *cobra.Command, args []string) {
			configFile, err = filepath.Abs(configFile)
			if err != nil {
				fmt.Println(err.Error())
			}
			demo := app.App(appName)

			demo.WithTitle("-> demo <-")
			demo.WithParam("configPath", configFile)
			demo.WithConfigLoader(app.NewFileConfigLoader(configFile))
			demo.WithLogger(log.WithSimpleLogger())
			demo.WithComponent(&Demo{}, app.WithDependencyTypes[app.Component]("Demo2"))
			os.Exit(demo.Start())
		},
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./config.yaml", "config")
	rootCmd.Execute()

}

type Demo struct {
	demo *Demo2 `ekit:"component"`
}

func (d Demo) Init(app *app.AppContext, conf *app.ConfContext) error {

	return nil
}

func (d Demo) Close() error {
	return nil
}

type Demo2 struct {
}

func (d Demo2) Init(app *app.AppContext, conf *app.ConfContext) error {

	return nil
}

func (d Demo2) Close() error {
	return nil
}
