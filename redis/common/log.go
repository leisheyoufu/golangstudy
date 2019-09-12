package common

import (
	"net/http"

	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	LOG_DEBUG = "debug"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_FATAL = "fatal"
	LOG_PANIC = "panic"
)

var (
	LOG_LEVEL = map[string]log.Level{
		LOG_DEBUG: log.DebugLevel,
		LOG_INFO:  log.InfoLevel,
		LOG_WARN:  log.WarnLevel,
		LOG_ERROR: log.ErrorLevel,
		LOG_FATAL: log.FatalLevel,
		LOG_PANIC: log.PanicLevel,
	}
	logFd *os.File
)

type Logger struct {
	plog *log.Entry
	pkg  string
}

func GetLogger(pkg string) *Logger {
	return &Logger{plog: log.WithFields(log.Fields{}), pkg: pkg}
}

func SetLogLevel(level string) {
	l := log.InfoLevel
	if _, ok := LOG_LEVEL[level]; !ok {
		keys := reflect.ValueOf(LOG_LEVEL).MapKeys()
		strkeys := make([]string, len(keys))
		for i := 0; i < len(keys); i++ {
			strkeys[i] = keys[i].String()
		}
		//plog.Warn(fmt.Sprintf("Error log level %s received. Only accept %s.", level, strings.Join(strkeys, " ")))
	} else {
		l = LOG_LEVEL[level]
	}
	log.SetLevel(l)
}

func InitLogger() {
	SetLogLevel(LOG_INFO)
	log.SetOutput(os.Stderr)
}

func CloseLogger() {
	if logFd != nil {
		logFd.Close()
		logFd = nil
	}
}

func (l *Logger) fileinfo() log.Fields {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return log.Fields{"file": fmt.Sprintf("%s/%s (%d)", l.pkg, file, line)}
}

func (l *Logger) HandleHttp(w http.ResponseWriter, req *http.Request, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	l.plog.WithFields(log.Fields{"code": code, "method": req.Method, "uri": req.Method})
	if msg != "" {
		l.plog.WithFields(l.fileinfo()).Error(msg)
		w.WriteHeader(code)
		fmt.Fprintf(w, "%s\n", msg)
	} else {
		l.plog.WithFields(l.fileinfo()).Info("OK")
		w.WriteHeader(http.StatusOK)
	}
}

func (l *Logger) ErrorNode(node string, message interface{}) {
	l.plog.WithFields(log.Fields{"node": node}).WithFields(l.fileinfo()).Error(message)
}

func (l *Logger) WarningNode(node string, message interface{}) {
	l.plog.WithFields(log.Fields{"node": node}).WithFields(l.fileinfo()).Warning(message)
}

func (l *Logger) DebugNode(node string, message interface{}) {
	l.plog.WithFields(log.Fields{"node": node}).WithFields(l.fileinfo()).Debug(message)
}

func (l *Logger) InfoNode(node string, message interface{}) {
	l.plog.WithFields(log.Fields{"node": node}).WithFields(l.fileinfo()).Info(message)
}

func (l *Logger) Info(message string) {
	l.plog.WithFields(l.fileinfo()).Info(message)
}

func (l *Logger) Warn(message string) {
	l.plog.WithFields(l.fileinfo()).Warn(message)
}

func (l *Logger) Error(err interface{}) {
	l.plog.WithFields(l.fileinfo()).Error(err)
}

func (l *Logger) Debug(err interface{}) {
	l.plog.WithFields(l.fileinfo()).Debug(err)
}
