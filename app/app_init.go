package app

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const MainLogger = "main"

func (r *RootComponent) initConf() error {
	if len(r.configLoaders) == 0 {
		r.configLoaders = append(r.configLoaders, withEmptyConfigLoader())
	}
	confContext := NewConfContext(r.configLoaders...)
	err := confContext.initConf()
	if err != nil {
		return err
	}
	r.conf = confContext
	return nil
}

func (r *RootComponent) initLog() error {
	if r.logInitFunc == nil {
		r.logInitFunc = withDefaultOutputLog(nil)
	}
	logger, err := r.logInitFunc.InitLog(r.appName, r.conf)
	if err != nil {
		return err
	}
	if logger == nil {
		return errors.New("logger must not be nil")
	}
	r.logger = logger.WithComponent(MainLogger)
	return nil
}

func (r *RootComponent) initAppContext() error {
	app := newAppContext(r.conf, r.exitNotifyCh, r.exitFinishedCh, r.logger, r.param)
	r.app = app
	return nil
}
func (r *RootComponent) initComponents() error {
	for _, c := range r.componentHolder {
		if count := componentDupCheck[c.ID()]; count > 1 {
			return errors.New("component duplicate: " + c.ID())
		}
		if c.IsSingleton() {
			t := string(c.componentType)
			if count := singletonComponentDupCheck[t]; count > 1 {
				return errors.New("singleton component duplicate: " + t)
			}
		}
	}
	ci := newComponentInitializer(r.componentHolder, r.app, r.conf)
	initSeq, err := ci.InitializeAll()
	if err != nil {
		return err
	}
	r.app.initSequence = initSeq
	r.componentHolder = nil
	return nil
}

func (r *RootComponent) closeComponents() error {
	for i := len(r.app.initSequence) - 1; i >= 0; i-- {
		id := r.app.initSequence[i]
		c := r.app.GetComponentById(id)
		if c == nil {
			return errors.New("component not found: " + id)
		}
		meta := r.app.Meta(c)
		if meta == nil {
			return errors.New("component meta not found: " + id)
		}
		meta.close()
	}
	err := r.conf.Close()
	if err != nil {
		return err
	}
	return nil
}

type ComponentInitializer struct {
	componentGraph       map[string]*ComponentMeta[Component]
	componentGroupByType map[string][]*ComponentMeta[Component]
	componentStatus      map[string]struct{}
	app                  *AppContext
	conf                 *ConfContext
	initSeq              []string
	initChain            []string
	logger               Logger
}

func newComponentInitializer(graph map[string]*ComponentMeta[Component], app *AppContext, conf *ConfContext) ComponentInitializer {
	m := map[string][]*ComponentMeta[Component]{}
	for _, c := range graph {
		if _, ok := m[string(c.Type())]; ok {
			m[string(c.Type())] = append(m[string(c.Type())], c)
		} else {
			m[string(c.Type())] = []*ComponentMeta[Component]{c}
		}
	}
	ci := ComponentInitializer{
		componentGraph:       graph,
		componentGroupByType: m,
		componentStatus:      map[string]struct{}{},
		app:                  app,
		conf:                 conf,
		logger:               app.MainLog,
	}
	return ci
}

func (ci *ComponentInitializer) InitializeAll() ([]string, error) {
	// init no
	for _, c := range ci.componentGraph {
		if len(c.Dependencies()) == 0 {
			if c.IsInitialized() {
				continue
			}
			err := ci.InitializeOne(c)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, c := range ci.componentGraph {
		if c.IsInitialized() {
			continue
		}
		ci.initChain = []string{}
		err := ci.InitializeOne(c)
		if err != nil {
			return nil, err
		}
	}
	return ci.initSeq, nil
}

func (ci *ComponentInitializer) InitializeOne(meta *ComponentMeta[Component]) error {
	if slices.Contains(ci.initChain, meta.ID()) {
		ci.initChain = append(ci.initChain, meta.ID())
		idx := slices.Index(ci.initChain, meta.ID())
		ci.initChain = ci.initChain[idx:]
		return fmt.Errorf("circular dependencies found:%s", strings.Join(ci.initChain, " -> "))
	}
	ci.initChain = append(ci.initChain, meta.ID())
	if meta.IsInitialized() {
		return nil
	}
	dependencies := meta.Dependencies()
	for _, t := range dependencies {
		metas := ci.componentGroupByType[t]
		for _, meta := range metas {
			m := ci.getMetaById(meta.ID())
			if m == nil {
				return fmt.Errorf("component[%s] required by [%s] not found", meta.ID(), meta.ID())
			}
			err := ci.InitializeOne(m)
			if err != nil {
				return err
			}
		}
	}
	err := meta.init(ci.app, ci.conf)
	if err != nil {
		return err
	}
	if meta.IsLazyInit() {
		ci.logger.Info("component", meta.ID(), "skip init cause lazy init")
	} else {
		ci.initSeq = append(ci.initSeq, meta.ID())
		ci.logger.Info("component", meta.ID(), "init success")
	}
	ci.app.addComponent(meta)
	return nil
}

func (ci *ComponentInitializer) getMetaById(id string) *ComponentMeta[Component] {
	return ci.componentGraph[id]
}
