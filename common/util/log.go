package utility

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	LOG_LEVEL_WARN  int = 0
	LOG_LEVEL_INFO  int = 1
	LOG_LEVEL_DEBUG int = 2
	LOG_LEVEL_TRACE int = 3
	LOG_LEVEL_ERROR int = -1
	LOG_LEVEL_FATAL int = -2
)

var LogLevelName = map[int]string{
	0:  "WARN",
	1:  "INFO",
	2:  "DEBUG",
	3:  "TRACE",
	-1: "ERROR",
	-2: "FATAL",
}

var LogLevel = map[string]int{
	"WARN":  0,
	"INFO":  1,
	"DEBUG": 2,
	"TRACE": 3,
	"ERROR": -1,
	"FATAL": -2,
}

var LogLevelEnvName = "LOG_LEVEL"
var LogFileInitialized = false

func SetLogLevelEnvName(name string) {
	LogLevelEnvName = name
	fmt.Println("Current log level setting is:", os.Getenv(LogLevelEnvName))
}

func GetSysLogLevel() int {
	name := os.Getenv(LogLevelEnvName)
	if name != "" {
		return LogLevel[os.Getenv(LogLevelEnvName)]
	} else {
		return LOG_LEVEL_DEBUG
	}
}

func LogWithLevel(level int, loggeronly bool, v ...interface{}) {

	syslevel := GetSysLogLevel()
	if level > syslevel {
		return
	}

	levelName := LogLevelName[level]

	if !LogFileInitialized {
		loggeronly = true
	}

	log.Println(levelName, GetHeader(0), v)
	if !loggeronly {
		fmt.Println(levelName, GetHeader(0), v)
	}
	if level == LOG_LEVEL_FATAL {
		pc, filename, line, _ := runtime.Caller(1)
		log.Printf("Handled error in %s[%s:%d]\n", runtime.FuncForPC(pc).Name(), filename, line)
	}
}

func GetHeader(depth int) string {
	_, file, line, ok := runtime.Caller(2 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return fmt.Sprintf("%s(%d)", file, line)
}

func LogWithLevelJSON(level int, logger bool, v interface{}) {

	data, err := json.Marshal(v)
	if err != nil {
		return
	}

	LogWithLevel(level, logger, string(data))
}

func GetStructJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(data)
}

func InitLogFile(logowner string, minversion string, folder string) (*os.File, error) {
	if folder == "" {
		return nil, nil
	}
	day := time.Now().Format("20060102")
	localdir := folder + "/" + logowner + "/" + day
	hostdir := "/data/logs" + "/" + logowner + "/" + day
	os.MkdirAll(localdir, 0755)

	// The encoder can't get POD_NAME env, generate a random string.
	podname := os.Getenv("POD_NAME")
	if podname == "" {
		podname = RandStringLetters(5)
	}

	logfile := "/" + logowner + "_" + time.Now().Format("20060102-150405") + "_" + podname + ".log"
	fmt.Println("The log file in the pod: " + localdir + logfile)
	fmt.Println("The log redirected to file on the node: " + hostdir + logfile)

	fp, err := os.OpenFile(localdir+logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	log.SetOutput(fp)
	log.Printf(logowner+" version: %s\n", minversion)

	LogFileInitialized = true
	return fp, nil
}

func StdOutFormat(v ...interface{}) {
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")

	fmt.Println(now, v)
}

func StdOutFormatWithLevel(level int, v ...interface{}) {

	syslevel := LogLevel[os.Getenv(LogLevelEnvName)]
	if level > syslevel {
		return
	}

	levelName := LogLevelName[level]

	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	fmt.Println(now, levelName, GetHeader(0), v)
}

type Logger struct {
}

func (log *Logger) Debug(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_DEBUG, false, v)
}
func (log *Logger) Info(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_INFO, false, v)
}
func (log *Logger) Warn(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_WARN, false, v)
}
func (log *Logger) Error(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_ERROR, false, v)
}
func (log *Logger) Fatal(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_FATAL, false, v)
}

func (log *Logger) FDebug(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_DEBUG, true, v)
}
func (log *Logger) FInfo(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_INFO, true, v)
}
func (log *Logger) FWarn(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_WARN, true, v)
}
func (log *Logger) FError(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_ERROR, true, v)
}
func (log *Logger) FFatal(v ...interface{}) {
	LogWithLevel(LOG_LEVEL_FATAL, true, v)
}
