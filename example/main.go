package main

import (
	"context"
	"fmt"
	"github.com/wyx0k/ekit/app"
	"github.com/wyx0k/ekit/log"
	"os"
	"path/filepath"
	"time"

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

			//demo.WithTitle("-> demo <-")
			demo.WithParam("configPath", configFile)
			demo.WithConfigLoader(app.NewFileConfigLoader(configFile))
			demo.WithLogger(log.WithSimpleLogger())
			demo.WithComponent(&Demo{})
			demo.WithComponent(&Demo2{A: "demo2 -----> ok"})
			demo.WithComponent(&RunnableDemo{})
			os.Exit(demo.Start())
		},
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./config.yaml", "config")
	rootCmd.Execute()

}

type RunnableDemo struct {
	app    *app.AppContext
	ctx    context.Context
	cancel context.CancelFunc
}

func (r *RunnableDemo) Init(app *app.AppContext, conf *app.ConfContext) error {
	r.app = app
	r.ctx, r.cancel = context.WithCancel(context.Background())
	app.MainLog.Info("runnable demo init......")
	return nil
}

func (r *RunnableDemo) Close() error {
	r.app.MainLog.Info("runnable demo close......")
	return nil
}

func (r *RunnableDemo) Run(app *app.AppContext, conf *app.ConfContext) error {
	for {
		app.MainLog.Info("runnable demo running......")
		select {
		case <-r.ctx.Done():
			app.MainLog.Info("runnable demo stop")
			return nil
		}
	}
	return nil
}

func (r *RunnableDemo) OnExit() error {
	time.Sleep(3 * time.Second)
	r.cancel()
	return nil
}

type Demo struct {
	Demo2 *Demo2 `ekit:"component;required:false"`
}

func (d Demo) Init(app *app.AppContext, conf *app.ConfContext) error {
	app.MainLog.Info(d.Demo2.A)
	app.MainLog.Info(app.GetParam("configPath"))
	return nil
}

func (d Demo) Close() error {
	return nil
}

type Demo2 struct {
	A string
}

func (d Demo2) Init(app *app.AppContext, conf *app.ConfContext) error {

	return nil
}

func (d Demo2) Close() error {
	return nil
}
