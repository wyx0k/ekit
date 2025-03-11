package app

import (
	"context"
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
	afterHandlers     map[string][]AfterInitHandler
	beforeHandlers    map[string][]BeforeInitHandler
	setupComponentErr []error
	logInitFunc       LogInitFuncInterface
	configLoaders     []ConfigLoader
	logger            Logger
	runningWg         sync.WaitGroup
	exitNotifyCh      chan string
	exitFinishedCh    chan struct{}
	initializedCtx    context.Context
	initializedCancel context.CancelFunc
	initialized       bool
	rootCtx           context.Context
	rootCtxCancel     context.CancelFunc
}

func App(name string, ctx ...context.Context) *RootComponent {
	now := time.Now()
	root := &RootComponent{
		startTime:               &now,
		appName:                 name,
		gracefulShutdownTimeout: 0,
		componentHolder:         map[string]*ComponentMeta[Component]{},
		beforeHandlers:          map[string][]BeforeInitHandler{},
		afterHandlers:           map[string][]AfterInitHandler{},
		param:                   map[string]any{},
		exitNotifyCh:            make(chan string),
		exitFinishedCh:          make(chan struct{}),
	}
	root.initializedCtx, root.initializedCancel = context.WithCancel(context.Background())
	if len(ctx) > 1 {
		fmt.Println("context must not be more than one")
		os.Exit(1)
	}
	if len(ctx) == 0 {
		root.rootCtx, root.rootCtxCancel = context.WithCancel(context.Background())
	} else {
		root.rootCtx, root.rootCtxCancel = context.WithCancel(ctx[0])
	}

	go func() {
		select {
		case <-root.initializedCtx.Done():
			root.initialized = true
		}
	}()
	return root
}

func (r *RootComponent) Start() (exitCode int) {
	defer r.rootCtxCancel()
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
	r.initializedCancel()
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

func (r *RootComponent) Exit(msg ...string) {
	r.app.Exit(msg...)
}

func (r *RootComponent) WaitUntilInitialized() {
	if r.initialized {
		return
	}
	select {
	case <-r.initializedCtx.Done():
	}
}

func (r *RootComponent) WaitUntilExit() {
	<-r.rootCtx.Done()
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
	typeName, types, instances, fields, err := resolveDependencies(component)
	if err != nil {
		r.setupComponentErr = append(r.setupComponentErr, err)
		return
	}
	if dependenciesExtendComponent, ok := component.(DependenciesExtendComponent); ok {
		ts, its := dependenciesExtendComponent.EkitDependencies()
		if len(ts) > 0 {
			types = append(types, ts...)
		}
		if len(its) > 0 {
			instances = append(instances, its...)
		}
	}
	if componentProvider, ok := component.(ComponentProvider); ok {
		components := componentProvider.EkitComponents()
		if len(components) > 0 {
			for _, c := range components {
				r.WithComponent(c)
			}
		}
	}
	options = append(options, withDependencyTypes[Component](types...),
		withDependencies[Component](instances...),
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
