package logging

import (
	"github.com/sky-as-code/nikki-erp/modules"
)

type Lvl int8

const (
	LVL_WARN  Lvl = 0
	LVL_INFO  Lvl = 1
	LVL_DEBUG Lvl = 2
	LVL_TRACE Lvl = 3
	LVL_ERROR Lvl = -1
	LVL_FATAL Lvl = -2
)

var LvlName = map[Lvl]string{
	LVL_WARN:  "WARN",
	LVL_INFO:  "INFO",
	LVL_DEBUG: "DEBUG",
	LVL_TRACE: "TRACE",
	LVL_ERROR: "ERROR",
	LVL_FATAL: "FATAL",
}

var LvlMap = map[string]Lvl{
	"WARN":  LVL_WARN,
	"INFO":  LVL_INFO,
	"DEBUG": LVL_DEBUG,
	"TRACE": LVL_TRACE,
	"ERROR": LVL_ERROR,
	"FATAL": LVL_FATAL,
}

type LoggerService interface {
	Level() Lvl
	SetLevel(level Lvl)
	Debug(i ...interface{})
	Debugf(format string, args ...interface{})
	Debugj(j modules.JSON)
	Info(i ...interface{})
	Infof(format string, args ...interface{})
	Infoj(j modules.JSON)
	Warn(i ...interface{})
	Warnf(format string, args ...interface{})
	Warnj(j modules.JSON)
	Error(i ...interface{})
	Errorf(format string, args ...interface{})
	Errorj(j modules.JSON)
	IfError(err error, i ...interface{})
	IfErrorf(err error, format string, args ...interface{})
	Fatal(i ...interface{})
	Fatalj(j modules.JSON)
	Fatalf(format string, args ...interface{})
	WithRequestId(requestId string)
}
