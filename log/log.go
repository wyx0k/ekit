package log

import "github.com/wyx0k/ekit/app"

var _logger app.Logger = new(app.DefaultOutputLog).WithComponent("default")

func Debug(args ...any) {
	_logger.Debug(args...)
}

func Debugf(format string, args ...any) {
	_logger.Debugf(format, args...)
}

func Info(args ...any) {
	_logger.Info(args...)
}

func Infof(format string, args ...any) {
	_logger.Infof(format, args...)
}

func Warn(args ...any) {
	_logger.Warn(args...)
}

func Warnf(format string, args ...any) {
	_logger.Warnf(format, args...)
}

func Error(args ...any) {
	_logger.Error(args...)
}

func Errorf(format string, args ...any) {
	_logger.Errorf(format, args...)
}

func Fatal(args ...any) {
	_logger.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	_logger.Fatalf(format, args...)
}
