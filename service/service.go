package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/wyx0k/ekit/app"
)

type HttpServiceConf struct {
	ServiceName  string `yaml:"serviceName"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	MaxRetryPort int    `yaml:"maxRetryPort"`
	Mode         string `yaml:"mode"`
}

type ServiceOpt[H http.Handler] func(engine H, conf *HttpServiceConf)

type HttpService[H http.Handler, R HttpServiceRouterResolver[H]] struct {
	conf     *HttpServiceConf
	engine   H
	server   *http.Server
	options  []ServiceOpt[H]
	app      *app.AppContext
	routes   []app.Component
	resolver R
}

func NewHttpService[H http.Handler, R HttpServiceRouterResolver[H]](engine H, resolver R, opts ...ServiceOpt[H]) *HttpService[H, R] {
	c := &HttpService[H, R]{
		options:  opts,
		engine:   engine,
		resolver: resolver,
		routes:   make([]app.Component, 0),
	}
	return c
}

func (h *HttpService[H, R]) Init(app *app.AppContext, conf *app.ConfContext) error {
	c := HttpServiceConf{}
	err := conf.Value("service").Scan(&c)
	if err != nil {
		return err
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.MaxRetryPort == 0 {
		c.MaxRetryPort = c.Port
	}
	if c.MaxRetryPort < c.Port {
		return errors.New("MaxRetryPort must gretter or equal Port")
	}
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}

	h.conf = &c
	h.app = app
	for _, opt := range h.options {
		opt(h.engine, h.conf)
	}
	return nil
}

func (h *HttpService[H, R]) Close() error {
	return nil
}

func (h *HttpService[H, R]) Run(app *app.AppContext, conf *app.ConfContext) error {
	port, err := detectPort(h.conf.Port, h.conf.MaxRetryPort, app.MainLog)
	if err != nil {
		return errors.New("failed to listen port: " + err.Error())
	}
	h.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", h.conf.Host, port),
		Handler: h.engine,
	}
	app.MainLog.Infof("service[%s] listen %d", h.conf.ServiceName, port)
	// 服务连接
	if err2 := h.server.ListenAndServe(); err2 != nil && !errors.Is(err2, http.ErrServerClosed) {
		return err2
	}
	return nil
}

func (h *HttpService[H, R]) OnExit() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.server.Shutdown(ctx); err != nil {
		return errors.New("Server Shutdown:" + err.Error())
	}
	return nil
}

func detectPort(port int, maxAvailablePort int, logger app.Logger) (int, error) {
	if maxAvailablePort == 0 {
		maxAvailablePort = port
	}
	currentPort := port
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("0.0.0.0:%d", currentPort), 3*time.Second)
		// 正在使用
		if err == nil {
			conn.Close()
			currentPort++
			if currentPort > maxAvailablePort {
				logger.Errorf("port [%d] is conflict", port, maxAvailablePort)
				break
			}
			continue
		} else {
			return currentPort, nil
		}
	}
	return 0, fmt.Errorf("can not find free port from [%d - %d]", port, maxAvailablePort)
}
