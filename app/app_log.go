package app

import (
	"fmt"
	"log"
	"os"
)

type LogInitFuncInterface interface {
	InitLog(appName string, conf *ConfContext) (Logger, error)
}

type LogInitFunc func(appName string, conf *ConfContext) (Logger, error)

func (l LogInitFunc) InitLog(appName string, conf *ConfContext) (Logger, error) {
	return l(appName, conf)
}

type Logger interface {
	WithComponent(name string) Logger
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

type TracedLogger interface {
	Logger
	WithTraceId(trace_id string)
	WithSpanId(span_id string)
}

type DefaultOutputLog struct {
	component string
	format    string
	start     string
}

func withDefaultOutputLog(_ *ConfContext) LogInitFunc {
	log.SetFlags(log.Ldate | log.Ltime)
	return func(_ string, conf *ConfContext) (Logger, error) {
		return &DefaultOutputLog{
			component: "main",
			format:    "%s %s %s %s",
			start:     "=>",
		}, nil
	}
}

func (s *DefaultOutputLog) WithComponent(name string) Logger {
	return &DefaultOutputLog{
		component: name,
		format:    "%s %s %s %s",
		start:     "=>",
	}
}

func (s *DefaultOutputLog) print(level string, args ...any) {
	args = append([]any{s.component, level, s.start}, args...)
	log.Println(args...)
}

func (s *DefaultOutputLog) printf(level, format string, args ...any) {
	log.Printf(fmt.Sprintf(s.format, s.component, level, s.start, format), args...)
}

func (s *DefaultOutputLog) Debug(args ...any) {
	s.print("DEBUG", args...)
}

func (s *DefaultOutputLog) Debugf(format string, args ...any) {
	s.printf("DEBUG", format, args...)
}

func (s *DefaultOutputLog) Info(args ...any) {
	s.print("INFO", args...)
}

func (s *DefaultOutputLog) Infof(format string, args ...any) {
	s.printf("INFO", format, args...)
}

func (s *DefaultOutputLog) Warn(args ...any) {
	s.print("WARN", args...)
}

func (s *DefaultOutputLog) Warnf(format string, args ...any) {
	s.printf("WARN", format, args...)
}

func (s *DefaultOutputLog) Error(args ...any) {
	s.print("ERROR", args...)
}

func (s *DefaultOutputLog) Errorf(format string, args ...any) {
	s.printf("ERROR", format, args...)
}

func (s *DefaultOutputLog) Fatal(args ...any) {
	s.print("FATAL", args...)
	os.Exit(7)
}

func (s *DefaultOutputLog) Fatalf(format string, args ...any) {
	s.printf("FATAL", format, args...)
	os.Exit(7)
}
