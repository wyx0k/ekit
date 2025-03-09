package app

import (
	"fmt"
	"sync"
)

type ExitStatus struct {
	HasErr bool
}

type AppContext struct {
	// id - meta
	components map[string]*ComponentMeta[Component]
	// type - meta
	singletonComponents map[string]*ComponentMeta[Component]
	componentMetas      map[Component]*ComponentMeta[Component]
	param               map[string]any
	conf                *ConfContext
	initSequence        []string
	MainLog             Logger
	exitNotifyCh        chan<- string
	exitFinishedCh      chan<- struct{}
	exitErrCh           chan<- error
	exitErrs            []error
	exitActionWg        sync.WaitGroup
}

func newAppContext(conf *ConfContext, exitNotifyCh chan<- string, exitFinishedCh chan<- struct{}, logger Logger, param map[string]any) *AppContext {
	exitErrCh := make(chan error)
	ac := &AppContext{
		components:          map[string]*ComponentMeta[Component]{},
		singletonComponents: map[string]*ComponentMeta[Component]{},
		componentMetas:      map[Component]*ComponentMeta[Component]{},
		param:               param,
		conf:                conf,
		MainLog:             logger,
		exitNotifyCh:        exitNotifyCh,
		exitFinishedCh:      exitFinishedCh,
		exitErrCh:           exitErrCh,
	}
	go func() {
		for err := range exitErrCh {
			if err != nil {
				ac.exitErrs = append(ac.exitErrs, err)
			}
			ac.exitActionWg.Done()
		}
	}()
	return ac
}

func (a *AppContext) addComponent(meta *ComponentMeta[Component]) {
	a.components[meta.ID()] = meta
	if meta.IsSingleton() {
		a.singletonComponents[string(meta.componentType)] = meta
	}
	a.componentMetas[meta.component] = meta
}

func (a *AppContext) GetParam(name string) (d any, ok bool) {
	d, ok = a.param[name]
	return
}

func (a *AppContext) SetParam(name string, value string) {
	a.param[name] = value
}

func (a *AppContext) Meta(c Component) *ComponentMeta[Component] {
	return a.componentMetas[c]
}

func (a *AppContext) GetComponentMeta(componentType ComponentType, name string) *ComponentMeta[Component] {
	id := getComponentID(componentType, name)
	return a.GetComponentMetaById(id)
}
func (a *AppContext) GetComponentMetaById(id string) *ComponentMeta[Component] {
	meta, ok := a.components[id]
	if !ok {
		return nil
	}
	if meta.IsLazyInit() {
		if !meta.IsLazyInitialized() {
			err := meta.lazyinit(a, a.conf)
			if err != nil {
				a.MainLog.Warnf("fail to lazy init component meta: %v", err)
			}
			a.initSequence = append(a.initSequence, meta.ID())
		}
	} else if !meta.IsInitialized() {
		return nil
	}
	return meta
}
func (a *AppContext) GetComponent(componentType ComponentType, name string) Component {
	id := getComponentID(componentType, name)
	return a.GetComponentById(id)
}

func (a *AppContext) GetComponentById(id string) Component {
	meta := a.GetComponentMetaById(id)
	return meta.component
}

func (a *AppContext) GetSingletonComponent(componentType string) Component {
	meta, ok := a.singletonComponents[componentType]
	if !ok {
		return nil
	}
	if !meta.IsInitialized() {
		if meta.IsLazyInit() {
			err := meta.lazyinit(a, a.conf)
			if err != nil {
				a.MainLog.Warnf("fail to lazy init component meta: %v", err)
			}
			a.initSequence = append(a.initSequence, meta.ID())
		} else {
			return nil
		}
	}
	return meta.component
}
func (a *AppContext) Exit(msg ...string) {
	for _, c := range a.components {
		if rc, ok := c.component.(RunnableComponent); ok {
			a.exitActionWg.Add(1)
			go func() {
				var rErr error
				defer func() {
					if err := recover(); err != nil {
						rErr = fmt.Errorf("component[%s] panic when exit:%v", c.ID(), err)
					}
					if rErr != nil {
						a.exitErrCh <- rErr
					} else {
						a.exitErrCh <- nil
					}
				}()
				rErr = rc.OnExit()
			}()
		}
	}
	go func() {
		a.exitActionWg.Wait()
		a.exitFinishedCh <- struct{}{}
	}()
	if len(msg) > 0 {
		a.exitNotifyCh <- msg[0]
	} else {
		a.exitNotifyCh <- ""
	}
}
