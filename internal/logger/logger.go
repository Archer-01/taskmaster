package logger

import (
	"fmt"
	"sync"
	"time"
)

type LogLevel uint8

const (
	InfoLevel LogLevel = iota
	WarnLevel
	ErrorLevel
	CriticalLevel
)

func (lvl LogLevel) String() string {
	switch lvl {
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERRO"
	case CriticalLevel:
		return "CRIT"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level LogLevel
	mutex sync.Mutex
}

var logger Logger = Logger{level: InfoLevel}

func SetLevel(level LogLevel) {
	logger.level = level
}

func Info(a any) {
	logger.log(InfoLevel, a)
}

func Infof(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	logger.log(InfoLevel, message)
}

func Warn(a any) {
	logger.log(WarnLevel, a)
}

func Warnf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	logger.log(WarnLevel, message)
}

func Error(a any) {
	logger.log(ErrorLevel, a)
}

func Errorf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	logger.log(ErrorLevel, message)
}

func Critical(a any) {
	logger.log(CriticalLevel, a)
}

func Criticalf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	logger.log(CriticalLevel, message)
}

func (l *Logger) log(level LogLevel, a any) {
	if level < l.level {
		return
	}

	date := time.Now().Format("2006-01-02 15:04:05,000")
	message := fmt.Sprintf("%s %s %s", date, level, a)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	fmt.Println(message)
}
