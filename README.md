# Ekit
A easy and simple tool-kit for building apps
## feature
- support dependency inject for ekit component while register component
- support dependency inject for ekit component by field tag
- support default log and config loader , extension point also provided for those so that user can implement by self
- support simple log implement by go-logging for advanced option
- support Component and RunnableComponent for different usecase
- user can custom slogan in console

## example
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
