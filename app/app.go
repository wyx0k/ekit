package app

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

const defaultTitle = `
  ______   _  __  _____   _______ 
 |  ____| | |/ / |_   _| |__   __|
 | |__    | ' /    | |      | |   
 |  __|   |  <     | |      | |   
 | |____  | . \   _| |_     | |   
 |______| |_|\_\ |_____|    |_|   
                                                                 
`

var componentDupCheck map[string]int = map[string]int{}
var singletonComponentDupCheck map[string]int = map[string]int{}

type RootComponent struct {
	startTime               *time.Time
	runningTime             *time.Time
	appName                 string
	title                   string
	gracefulShutdownTimeout time.Duration

	app   *AppContext
	conf  *ConfContext
	param map[string]any

	componentHolder   map[string]*ComponentMeta[Component]
	setupComponentErr []error
	logInitFunc       LogInitFuncInterface
	configLoaders     []ConfigLoader
	logger            Logger
	runningWg         sync.WaitGroup
	exitNotifyCh      chan string
	exitFinishedCh    chan struct{}
}

func App(name string) *RootComponent {

	now := time.Now()
	root := &RootComponent{
		startTime:               &now,
		appName:                 name,
		gracefulShutdownTimeout: 0,
		componentHolder:         map[string]*ComponentMeta[Component]{},
		param:                   map[string]any{},
		exitNotifyCh:            make(chan string),
		exitFinishedCh:          make(chan struct{}),
	}
	return root
}

func (r *RootComponent) Start() (exitCode int) {
	r.printStart()
	err := r.initConf()
	if err != nil {
		fmt.Println("failed to initialize config:", err.Error())
		exitCode = 1
		return
	}
	err = r.initLog()
	if err != nil {
		fmt.Println("failed to initialize log:", err.Error())
		exitCode = 2
		return
	}
	err = r.initAppContext()
	if err != nil {
		r.logger.Error("failed to initialize app context:", err.Error())
		exitCode = 3
		return
	}
	if len(r.setupComponentErr) > 0 {
		r.logger.Error("failed to setup component:", errors.Join(r.setupComponentErr...))
		exitCode = 4
		return
	}
	err = r.initComponents()
	if err != nil {
		r.logger.Error("failed to initialize components:", err.Error())
		exitCode = 4
		return
	}
	now := time.Now()
	r.runningTime = &now
	r.printTitle()
	err = r.runAll()
	if err != nil {
		r.logger.Error("failed to run components:", err.Error())
		exitCode = 5
	}
	err = r.closeComponents()
	if err != nil {
		r.logger.Error("failed to close components:", err.Error())
		exitCode = 6
		return
	}
	r.logger.Info("app exit successfully")
	return
}

func (r *RootComponent) WithTitle(title string) {
	r.title = title
}

func (r *RootComponent) WithGracefulShutdownTimeout(d time.Duration) {
	r.gracefulShutdownTimeout = d
}

func (r *RootComponent) WithParam(name string, value any) {
	r.param[name] = value
}

func (r *RootComponent) WithLogger(logInitFunc LogInitFunc) {
	r.logInitFunc = logInitFunc
}

func (r *RootComponent) WithConfigLoader(loader ConfigLoader) {
	r.configLoaders = append(r.configLoaders, loader)
}

func (r *RootComponent) WithComponentMeta(name string, componentMeta *ComponentMeta[Component]) {
	err := componentMeta.preInit(name)
	if err != nil {
		r.logger.Error(err)
		os.Exit(1)
	}
	r.componentHolder[componentMeta.ID()] = componentMeta
	componentDupCheck[componentMeta.ID()] = componentDupCheck[componentMeta.ID()] + 1
	t := string(componentMeta.componentType)
	singletonComponentDupCheck[t] = singletonComponentDupCheck[t] + 1
}

func (r *RootComponent) WithComponent(component Component, options ...ComponentMetaOption[Component]) {
	r.WithNamedComponent("", component)
}

func (r *RootComponent) WithNamedComponent(name string, component Component, options ...ComponentMetaOption[Component]) {
	typeName, types, instancies, fields, err := resolveDependencies(component)
	if err != nil {
		r.setupComponentErr = append(r.setupComponentErr, err)
		return
	}
	options = append(options, withDependencyTypes[Component](types...),
		withDependencies[Component](instancies...),
		withFieldInfo[Component](fields))
	meta := NewComponentMeta(ComponentType(typeName), component, options...)
	if name == "" {
		name = typeName
	}
	r.WithComponentMeta(name, meta)
}

func (r *RootComponent) printStart() {
	fmt.Println(r.appName, "starting at", r.startTime.Format(time.DateTime))
}
func (r *RootComponent) printTitle() {
	if r.title == "" {
		r.title = defaultTitle
	}
	r.logger.Info(r.title)
	r.logger.Info(r.appName, "running at", r.runningTime.Format(time.DateTime))
}
