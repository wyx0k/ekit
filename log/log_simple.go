package log

import (
	"fmt"
	log "github.com/op/go-logging"
	"github.com/wyx0k/ekit/app"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
)

var format = log.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} %{level:.5s} %{shortfile} %{module}:%{color:reset} %{message}`,
)

type logConf struct {
	Path       string `json:"path"`
	MaxSize    int    `json:"maxSize"`
	MaxAge     int    `json:"maxAge"`
	MaxBackups int    `json:"maxBackups"`
	Compress   bool   `json:"compress"`
}
type SimpleLogger struct {
	path       string
	maxSize    int
	maxAge     int
	maxBackups int
	compress   bool
	logger     *log.Logger
	lumberjack *lumberjack.Logger
}

func module(name string) string {
	return fmt.Sprintf("[%s]", name)
}

func WithSimpleLogger() app.LogInitFunc {
	return func(appName string, conf *app.ConfContext) (app.Logger, error) {
		lc := logConf{}
		err := conf.Value("log").Scan(&lc)
		if err != nil {
			return nil, err
		}
		path := fmt.Sprintf("%s/%s.log", lc.Path, appName)
		p, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		sl := SimpleLogger{
			path:       p,
			maxSize:    lc.MaxAge,
			maxAge:     lc.MaxAge,
			maxBackups: lc.MaxBackups,
			compress:   lc.Compress,
		}

		sl.lumberjack = &lumberjack.Logger{
			Filename:   sl.path,
			MaxSize:    sl.maxSize,
			MaxAge:     sl.maxAge,
			MaxBackups: sl.maxBackups,
			LocalTime:  true,
			Compress:   sl.compress,
		}

		out := io.MultiWriter(sl.lumberjack, os.Stdout)
		backend := log.NewLogBackend(out, "", 0)
		backendFormatter := log.NewBackendFormatter(backend, format)
		log.SetBackend(backendFormatter)
		sl.logger = log.MustGetLogger(module(appName))
		sl.logger.ExtraCalldepth = 2
		_logger = &sl
		return &sl, nil
	}
}

func (s *SimpleLogger) WithComponent(name string) app.Logger {
	sl := &SimpleLogger{
		path:       s.path,
		maxSize:    s.maxSize,
		maxAge:     s.maxAge,
		maxBackups: s.maxBackups,
		compress:   s.compress,
	}
	sl.logger = log.MustGetLogger(module(name))
	sl.logger.ExtraCalldepth = 1
	return sl
}

func (s *SimpleLogger) Debug(args ...any) {
	s.logger.Debug(args...)
}

func (s *SimpleLogger) Debugf(format string, args ...any) {
	s.logger.Debugf(format, args...)
}

func (s *SimpleLogger) Info(args ...any) {
	s.logger.Info(args...)
}

func (s *SimpleLogger) Infof(format string, args ...any) {
	s.logger.Infof(format, args...)
}

func (s *SimpleLogger) Warn(args ...any) {
	s.logger.Warning(args...)
}

func (s *SimpleLogger) Warnf(format string, args ...any) {
	s.logger.Warningf(format, args...)
}

func (s *SimpleLogger) Error(args ...any) {
	s.logger.Error(args...)
}

func (s *SimpleLogger) Errorf(format string, args ...any) {
	s.logger.Errorf(format, args...)
}

func (s *SimpleLogger) Fatal(args ...any) {
	s.logger.Fatal(args...)
}

func (s *SimpleLogger) Fatalf(format string, args ...any) {
	s.logger.Fatalf(format, args...)
}
