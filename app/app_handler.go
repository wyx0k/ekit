package app

type AfterInitHandler func(app *AppContext, conf *ConfContext, target Component)
type BeforeInitHandler func(app *AppContext, conf *ConfContext)

func (r *RootComponent) AfterComponentTypeInit(componentType ComponentType, handler AfterInitHandler) {
	ct := string(componentType)
	if _, ok := r.afterHandlers[ct]; ok {
		r.afterHandlers[ct] = append(r.afterHandlers[ct], handler)
	} else {
		r.afterHandlers[ct] = []AfterInitHandler{handler}
	}

}

func (r *RootComponent) BeforeComponentTypeInit(componentType ComponentType, handler BeforeInitHandler) {
	ct := string(componentType)
	if _, ok := r.beforeHandlers[ct]; ok {
		r.beforeHandlers[ct] = append(r.beforeHandlers[ct], handler)
	} else {
		r.beforeHandlers[ct] = []BeforeInitHandler{handler}
	}
}
