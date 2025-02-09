package app

import (
	"errors"
	"reflect"
)

// will block main go routine
type RunnableComponent interface {
	Component
	Run(app *AppContext, conf *ConfContext) error
	OnExit() error
}

type Component interface {
	Init(app *AppContext, conf *ConfContext) error
	Close() error
}
type ComponentType string

type ComponentMeta[T Component] struct {
	componentID       string
	componentName     string
	componentType     ComponentType
	dependencies      []string
	singleton         bool
	ignoreError       bool
	lazyInit          bool
	_initialized      bool
	_lazy_initialized bool

	component T
}

type ComponentMetaOption[T Component] func(meta *ComponentMeta[T])

func WithSingleton[T Component](meta *ComponentMeta[T]) {
	meta.singleton = true
}
func WithIgnoreError[T Component](meta *ComponentMeta[T]) {
	meta.ignoreError = true
}
func WithLazyInit[T Component](meta *ComponentMeta[T]) {
	meta.lazyInit = true
}
func WithDependencies[T Component](dependencies ...ComponentType) ComponentMetaOption[T] {
	return func(meta *ComponentMeta[T]) {
		var lst []string
		for _, d := range dependencies {
			lst = append(lst, string(d))
		}
		meta.dependencies = lst
	}
}

func NewComponentMeta[T Component](componentType ComponentType, component T, options ...ComponentMetaOption[T]) *ComponentMeta[T] {
	cm := &ComponentMeta[T]{
		componentType: componentType,
		component:     component,
	}
	for _, option := range options {
		option(cm)
	}
	return cm
}

func (cm *ComponentMeta[T]) preInit(name string) error {
	if cm == nil {
		return errors.New("component-meta must not be nil")
	}
	if reflect.ValueOf(cm.component).IsNil() {
		return errors.New("component in component-meta must not be nil")
	}
	cm.componentName = name
	cm.componentID = getComponentID(cm.componentType, name)
	return nil
}

func (cm *ComponentMeta[T]) ID() string {
	return cm.componentID
}
func (cm *ComponentMeta[T]) Name() string {
	return cm.componentName
}
func (cm *ComponentMeta[T]) Type() ComponentType {
	return cm.componentType
}
func (cm *ComponentMeta[T]) Dependencies() []string {
	return cm.dependencies
}
func (cm *ComponentMeta[T]) IsSingleton() bool {
	return cm.singleton
}
func (cm *ComponentMeta[T]) IsIgnoreError() bool {
	return cm.ignoreError
}
func (cm *ComponentMeta[T]) IsLazyInit() bool {
	return cm.lazyInit
}
func (cm *ComponentMeta[T]) IsInitialized() bool {
	return cm._initialized
}
func (cm *ComponentMeta[T]) IsLazyInitialized() bool {
	return cm._lazy_initialized
}

func (cm *ComponentMeta[T]) init(app *AppContext, conf *ConfContext) error {
	if cm.lazyInit {
		cm._initialized = true
		return nil
	}
	err := cm.component.Init(app, conf)
	if err != nil {
		return err
	}
	cm._initialized = true
	return nil
}
func (cm *ComponentMeta[T]) lazyinit(app *AppContext, conf *ConfContext) error {
	if !cm.lazyInit {
		return nil
	}
	err := cm.component.Init(app, conf)
	if err != nil {
		return err
	}
	cm._lazy_initialized = true
	return nil
}
func (cm *ComponentMeta[T]) close() error {
	if cm.IsLazyInit() {
		if cm._lazy_initialized {
			return cm.component.Close()
		}
	} else {
		if cm._initialized {
			return cm.component.Close()
		}
	}
	return nil
}

func getComponentID(componentType ComponentType, componentName string) string {
	return string(componentType) + ":" + componentName
}
