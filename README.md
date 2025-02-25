# Ekit

## Introduction
A easy and simple tool-kit for building apps.
```go
ekit_is:=`
  ______   _  __  _____   _______ 
 |  ____| | |/ / |_   _| |__   __|
 | |__    | ' /    | |      | |   
 |  __|   |  <     | |      | |   
 | |____  | . \   _| |_     | |   
 |______| |_|\_\ |_____|    |_|   
                                                                 
`
```
## Features
- support dependency inject for ekit component while register component
- support dependency inject for ekit component by field tag
- support default log and config loader , extension point also provided for those so that user can implement by self
- support simple log implement by go-logging for advanced option
- support Component and RunnableComponent for different usecase
- user can custom slogan in console

## Getting Started
### Dependency Injection
You can inject registered component by tag "ekit", and specify "component" like this:
```go
package main

import (
	"github.com/wyx0k/ekit/app"
	"os"
)

type Demo struct {
	Demo2 *Demo2 `ekit:"component;required:false"`
}

func (d Demo) Init(app *app.AppContext, conf *app.ConfContext) error {
	app.MainLog.Info(d.Demo2.A)
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

func main() {

	demo := app.App("demo")
	demo.WithComponent(&Demo{})
	demo.WithComponent(&Demo2{A: "demo2 -----> ok"})
	os.Exit(demo.Start())
}
```
### Runnable Component
Runnable component also support for which such as web server
```go
package main

import (
"context"
"github.com/wyx0k/ekit/app"
"os"
"time"
)

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

func (r RunnableDemo) Close() error {
	r.app.MainLog.Info("runnable demo close......")
	return nil
}

func (r RunnableDemo) Run(app *app.AppContext, conf *app.ConfContext) error {
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

func (r RunnableDemo) OnExit() error {
	time.Sleep(3 * time.Second)
	r.cancel()
	return nil
}

func main() {
	demo := app.App("demo")
	demo.WithComponent(&RunnableDemo{})
	os.Exit(demo.Start())
}
```
### Config Loader
You can use default file config loader or implement your own config loader.
Default file config loader support hot reload.

interface
```go
type ConfigLoader interface {
	Load(updater *ConfigUpdater) error
}
```

usage
```go
demo := app.App("demo")
demo.WithConfigLoader(app.NewFileConfigLoader(filePath))
```

### Log
You can use default log or implement your own logger.Default log dosen't write to file,so we provide another simple logger in ekit/log package.

interface
```go
type Logger interface {
	WithComponent(name string) Logger
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}
```
usage
```go
demo := app.App("demo")
demo.WithLogger(log.WithSimpleLogger())
```

### Slogan
You can change default slogan like this:

```go
demo := app.App("demo")
demo.WithTitle("-> demo <-")
```