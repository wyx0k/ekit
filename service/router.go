package service

import (
	"net/http"

	"github.com/wyx0k/ekit/app"
)

type HttpServiceRouterResolver[T http.Handler] interface {
	Resolve(engine T, route app.Component) error
}

func (h *HttpService[T, R]) AddRoute(router app.Component) error {
	h.routes = append(h.routes, router)
	err := h.resolver.Resolve(h.engine, router)
	return err
}
func (h *HttpService[T, R]) EkitComponents() []app.Component {
	return h.routes
}
