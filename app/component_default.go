package app

import (
	"context"
)

type SimpleComponent struct {
	app  *AppContext
	conf *ConfContext
}

func (g *SimpleComponent) Init(app *AppContext, conf *ConfContext) error {
	g.conf = conf
	g.app = app
	return nil
}

func (g *SimpleComponent) Close() error {
	return nil
}

type SimpleRunnableComponent struct {
	app    *AppContext
	ctx    context.Context
	cancel context.CancelFunc
}

func (r *SimpleRunnableComponent) Init(app *AppContext, conf *ConfContext) error {
	r.app = app
	r.ctx, r.cancel = context.WithCancel(context.Background())
	return nil
}

func (r *SimpleRunnableComponent) Close() error {
	return nil
}

func (r *SimpleRunnableComponent) Run(app *AppContext, conf *ConfContext) error {
	<-r.ctx.Done()
	return nil
}

func (r *SimpleRunnableComponent) OnExit() error {
	r.cancel()
	return nil
}

func (r *SimpleRunnableComponent) Done() <-chan struct{} {
	return r.ctx.Done()
}
