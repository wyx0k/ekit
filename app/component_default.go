package app

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
