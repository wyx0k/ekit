package app

import (
	"errors"
	"fmt"
	"reflect"
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
	ci, err := newComponentInitializer(r.componentHolder, r.app, r.conf)
	if err != nil {
		return err
	}
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
		meta := r.app.GetComponentMetaById(id)
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
	componentGraph        map[string]*ComponentMeta[Component]
	componentGroupByType  map[string][]*ComponentMeta[Component]
	componentPrimaryGraph map[string]*ComponentMeta[Component]
	componentStatus       map[string]struct{}
	app                   *AppContext
	conf                  *ConfContext
	initSeq               []string
	initChain             []string
	logger                Logger
}

func newComponentInitializer(graph map[string]*ComponentMeta[Component], app *AppContext, conf *ConfContext) (*ComponentInitializer, error) {
	m := map[string][]*ComponentMeta[Component]{}
	primary := map[string]*ComponentMeta[Component]{}
	for _, c := range graph {
		if _, ok := m[string(c.Type())]; ok {
			m[string(c.Type())] = append(m[string(c.Type())], c)
		} else {
			m[string(c.Type())] = []*ComponentMeta[Component]{c}
		}
		if c.IsPrimary() {
			if meta, exist := primary[string(c.Type())]; exist {
				return nil, errors.New("duplicated primary componet " + string(c.Type()) + ": " + c.componentID + ", " + meta.componentID)
			}
			primary[string(c.Type())] = c
		}
	}
	ci := &ComponentInitializer{
		componentGraph:        graph,
		componentGroupByType:  m,
		componentPrimaryGraph: primary,
		componentStatus:       map[string]struct{}{},
		app:                   app,
		conf:                  conf,
		logger:                app.MainLog,
	}
	return ci, nil
}

func (ci *ComponentInitializer) InitializeAll() ([]string, error) {
	// init no
	for _, c := range ci.componentGraph {
		if len(c.Dependencies()) == 0 && len(c.DependencyTypes()) == 0 {
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

func (ci *ComponentInitializer) InitializeOne(inMeta *ComponentMeta[Component]) error {
	if slices.Contains(ci.initChain, inMeta.ID()) {
		ci.initChain = append(ci.initChain, inMeta.ID())
		idx := slices.Index(ci.initChain, inMeta.ID())
		ci.initChain = ci.initChain[idx:]
		return fmt.Errorf("circular dependencies found:%s", strings.Join(ci.initChain, " -> "))
	}
	ci.initChain = append(ci.initChain, inMeta.ID())
	defer func() {
		ci.initChain = ci.initChain[:len(ci.initChain)-1]
	}()
	if inMeta.IsInitialized() {
		return nil
	}
	types := inMeta.DependencyTypes()
	for _, t := range types {
		metas := ci.componentGroupByType[t]
		if len(metas) == 0 && inMeta.IsAdditionalDepends(t) {
			return fmt.Errorf("component type[%s] required by [%s] but found 0 candidates", t, inMeta.ID())
		}
		for _, meta := range metas {
			m := ci.getMetaById(meta.ID())
			if m == nil {
				return fmt.Errorf("component[%s] required by [%s] not found", meta.ID(), inMeta.ID())
			}
			err := ci.InitializeOne(m)
			if err != nil {
				return err
			}
		}
	}
	instancies := inMeta.Dependencies()
	for _, instance := range instancies {
		m := ci.getMetaById(instance)
		if m == nil {
			return fmt.Errorf("component[%s] required by [%s] not found", instance, inMeta.ID())
		}
		err := ci.InitializeOne(m)
		if err != nil {
			return err
		}
	}
	// dependency inject
	err := ci.dependencyInject(inMeta)
	if err != nil {
		return err
	}
	err = inMeta.init(ci.app, ci.conf)
	if err != nil {
		return err
	}
	if inMeta.IsLazyInit() {
		ci.logger.Info("component", inMeta.ID(), "skip init cause lazy init")
	} else {
		ci.initSeq = append(ci.initSeq, inMeta.ID())
		ci.logger.Info("component", inMeta.ID(), "init success")
	}
	ci.app.addComponent(inMeta)
	return nil
}

func (ci *ComponentInitializer) getMetaById(id string) *ComponentMeta[Component] {
	return ci.componentGraph[id]
}

func (ci *ComponentInitializer) dependencyInject(inMeta *ComponentMeta[Component]) error {
	fields := inMeta.fieldMap()
	if len(fields) == 0 {
		return nil
	}
	t := reflect.TypeOf(inMeta.component).Elem()
	v := reflect.ValueOf(inMeta.component)

	tv := v.Elem()
	for i := 0; i < tv.NumField(); i++ {
		field := tv.Field(i)
		fieldT := t.Field(i)
		if field.CanSet() {
			if diInfo, ok := fields[fieldT.Name]; ok {
				if ok {
					var components []Component
					componentMap := map[string]Component{}
					if diInfo.IsDependAll {
						if metas, ok2 := ci.componentGroupByType[string(diInfo.DependType)]; ok2 {
							for _, meta := range metas {
								components = append(components, meta.component)
								componentMap[meta.componentName] = meta.component
							}
						}
					} else {
						for _, id := range diInfo.DependIds {
							if meta, ok2 := ci.componentGraph[id]; ok2 {
								components = append(components, meta.component)
								componentMap[meta.componentName] = meta.component
							}
						}
					}
					if len(components) == 0 {
						if diInfo.Required {
							return errors.New(t.Name() + "." + diInfo.FieldName + " can not set, found 0 candidates, but field is required, or you can add \"required:false\" on field tag to avoid this")
						}
						continue
					}
					switch diInfo.FieldKind {
					case reflect.Slice:
						targetSlice := reflect.MakeSlice(diInfo.FieldType, len(components), len(components))
						for i, c := range components {
							cv := reflect.ValueOf(c)
							targetSlice.Index(i).Set(cv)
						}
						field.Set(targetSlice)
					case reflect.Map:
						targetMap := reflect.MakeMap(diInfo.FieldType)
						for k, c := range componentMap {
							cv := reflect.ValueOf(c)
							targetMap.SetMapIndex(reflect.ValueOf(k), cv)
						}
						field.Set(targetMap)
					case reflect.Ptr:
						if len(components) > 1 {
							if p, ok := ci.componentPrimaryGraph[string(diInfo.DependType)]; ok {
								vv := reflect.ValueOf(p.component)
								field.Set(vv)
							} else {
								return errors.New(t.Name() + "." + diInfo.FieldName + " can not set, found more than 1 candidates, please specify name on tag or set primary when register component")
							}
						} else {
							vv := reflect.ValueOf(components[0])
							field.Set(vv)
						}
					}
				}
			}
		}
	}
	return nil
}
