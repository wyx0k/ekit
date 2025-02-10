package app

import (
	"errors"
	"testing"
)

var A, B, C, D ComponentType = "A", "B", "C", "D"

type TestFakeComponent struct {
	name string
}

func NewFakeComponent(name string) Component {
	c := &TestFakeComponent{name: name}
	return c
}

func (h *TestFakeComponent) Init(app *AppContext, conf *ConfContext) error {
	app.MainLog.Info("init", h.name)
	if h.name == "A" {
		b := app.GetComponent(B, "b")
		if b == nil {
			app.MainLog.Info("b not found")
			return errors.New("b not found")
		}
	}
	return nil
}
func (h *TestFakeComponent) Close() error {
	return nil
}
func (h *TestFakeComponent) Run() error {
	return nil
}

func TestComponentCircularDependencies(t *testing.T) {
	appName := "demo"
	app := App(appName)
	a := NewFakeComponent("A")
	b := NewFakeComponent("B")
	c := NewFakeComponent("C")
	d := NewFakeComponent("D")
	ma := NewComponentMeta(A, a, WithDependencyTypes[Component](B))
	mb := NewComponentMeta(B, b, WithDependencyTypes[Component](C))
	mc := NewComponentMeta(C, c, WithDependencyTypes[Component](D))
	md := NewComponentMeta(D, d, WithDependencyTypes[Component](C))
	app.WithComponentMeta("a", ma)
	app.WithComponentMeta("b", mb)
	app.WithComponentMeta("c", mc)
	app.WithComponentMeta("d", md)
	app.Start()
}
func TestComponentLazyInitDependencies(t *testing.T) {
	appName := "demo"
	app := App(appName)
	var A, B, C, D ComponentType = "A", "B", "C", "D"
	a := NewFakeComponent("A")
	b := NewFakeComponent("B")
	c := NewFakeComponent("C")
	d := NewFakeComponent("D")
	ma := NewComponentMeta(A, a, WithDependencyTypes[Component](B))
	mb := NewComponentMeta(B, b, WithDependencyTypes[Component](C), WithLazyInit[Component])
	mc := NewComponentMeta(C, c, WithDependencyTypes[Component](D))
	md := NewComponentMeta(D, d)
	app.WithComponentMeta("a", ma)
	app.WithComponentMeta("b", mb)
	app.WithComponentMeta("c", mc)
	app.WithComponentMeta("d", md)
	app.Start()
}
func TestComponentDependencies(t *testing.T) {
	appName := "demo"
	app := App(appName)
	var A, B ComponentType = "A", "B"
	a := NewFakeComponent("A")
	b := NewFakeComponent("B")
	c := NewFakeComponent("C")
	d := NewFakeComponent("D")
	ma := NewComponentMeta(A, a, WithDependencyTypes[Component](B))
	mb := NewComponentMeta(A, b, WithDependencyTypes[Component](B))
	mc := NewComponentMeta(A, c, WithDependencyTypes[Component](B))
	md := NewComponentMeta(B, d)
	app.WithComponentMeta("a", ma)
	app.WithComponentMeta("b", mb)
	app.WithComponentMeta("c", mc)
	app.WithComponentMeta("d", md)
	app.Start()
}
