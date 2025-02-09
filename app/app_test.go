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
	ma := NewComponentMeta(A, a, WithDependencies[Component](B))
	mb := NewComponentMeta(B, b, WithDependencies[Component](C))
	mc := NewComponentMeta(C, c, WithDependencies[Component](D))
	md := NewComponentMeta(D, d, WithDependencies[Component](C))
	app.WithComponent("a", ma)
	app.WithComponent("b", mb)
	app.WithComponent("c", mc)
	app.WithComponent("d", md)
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
	ma := NewComponentMeta(A, a, WithDependencies[Component](B))
	mb := NewComponentMeta(B, b, WithDependencies[Component](C), WithLazyInit[Component])
	mc := NewComponentMeta(C, c, WithDependencies[Component](D))
	md := NewComponentMeta(D, d)
	app.WithComponent("a", ma)
	app.WithComponent("b", mb)
	app.WithComponent("c", mc)
	app.WithComponent("d", md)
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
	ma := NewComponentMeta(A, a, WithDependencies[Component](B))
	mb := NewComponentMeta(A, b, WithDependencies[Component](B))
	mc := NewComponentMeta(A, c, WithDependencies[Component](B))
	md := NewComponentMeta(B, d)
	app.WithComponent("a", ma)
	app.WithComponent("b", mb)
	app.WithComponent("c", mc)
	app.WithComponent("d", md)
	app.Start()
}
