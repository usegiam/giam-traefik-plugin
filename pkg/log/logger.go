package log

import (
	"fmt"
	"log"
	"os"
)

// Log levels
const (
	DEBUG = iota
	INFO
	ERROR
	FATAL
)

// Logger struct to hold log level and output destination
type Logger struct {
	level  int
	logger *log.Logger
}

func New(level string) *Logger {
	logLevel := INFO

	switch level {
	case "INFO":
		logLevel = INFO
		break
	case "DEBUG":
		logLevel = DEBUG
		break
	case "ERROR":
		logLevel = ERROR
		break
	case "FATAL":
		logLevel = FATAL
		break
	default:
		logLevel = INFO
	}

	return &Logger{level: logLevel, logger: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)}
}

// Debug logs debug-level messages
func (l *Logger) Debug(msg string) {
	if l.level <= DEBUG {
		l.logger.Output(2, msg)
	}
}

// Debugf logs formatted messages at DEBUG level
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info logs info-level messages
func (l *Logger) Info(msg string) {
	if l.level <= INFO {
		l.logger.Output(2, msg)
	}
}

// Infof logs formatted messages at INFO level
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.logger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error logs error-level messages
func (l *Logger) Error(msg string) {
	if l.level <= ERROR {
		l.logger.Output(2, msg)
	}
}

// Errorf logs formatted message error-level
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.logger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Fatal logs fatal-level messages and exits the program
func (l *Logger) Fatal(msg string) {
	if l.level <= FATAL {
		l.logger.Output(2, msg)
		os.Exit(1)
	}
}

// Fatalf logs formatted messages at fatal-level and exits the program
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.level <= FATAL {
		l.logger.Output(2, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}
