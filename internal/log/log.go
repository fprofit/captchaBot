package log

import (
	stdlog "log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Logger struct {
	logger *stdlog.Logger
	file   *os.File
	level  Level
}

func New(debug bool, level string) (*Logger, error) {
	l := &Logger{
		level: parseLevel(level),
	}

	var output *os.File

	if debug {
		output = os.Stdout
	} else {
		if err := os.MkdirAll("logs", 0755); err != nil {
			return nil, err
		}

		filename := filepath.Join(
			"logs",
			time.Now().Format("02.01.2006")+".log",
		)

		file, err := os.OpenFile(
			filename,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0644,
		)
		if err != nil {
			return nil, err
		}

		l.file = file
		output = file
	}

	l.logger = stdlog.New(
		output,
		"",
		stdlog.Ldate|stdlog.Ltime|stdlog.Lshortfile,
	)

	return l, nil
}

func (l *Logger) Close() error {
	if l.file == nil {
		return nil
	}

	err := l.file.Close()
	l.file = nil

	return err
}

func (l *Logger) Debug(message string) {
	if l.level > DebugLevel {
		return
	}

	l.logger.Printf("[DEBUG] %s", message)
}

func (l *Logger) Info(message string) {
	if l.level > InfoLevel {
		return
	}

	l.logger.Printf("[INFO] %s", message)
}

func (l *Logger) Warn(message string) {
	if l.level > WarnLevel {
		return
	}

	l.logger.Printf("[WARN] %s", message)
}

func (l *Logger) Error(message string) {
	if l.level > ErrorLevel {
		return
	}

	l.logger.Printf("[ERROR] %s", message)
}

func (l *Logger) Debugf(format string, v ...any) {
	if l.level > DebugLevel {
		return
	}

	l.logger.Printf("[DEBUG] "+format, v...)
}

func (l *Logger) Infof(format string, v ...any) {
	if l.level > InfoLevel {
		return
	}

	l.logger.Printf("[INFO] "+format, v...)
}

func (l *Logger) Warnf(format string, v ...any) {
	if l.level > WarnLevel {
		return
	}

	l.logger.Printf("[WARN] "+format, v...)
}

func (l *Logger) Errorf(format string, v ...any) {
	if l.level > ErrorLevel {
		return
	}

	l.logger.Printf("[ERROR] "+format, v...)
}

func parseLevel(value string) Level {
	switch strings.ToUpper(value) {
	case "DEBUG":
		return DebugLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		return InfoLevel
	}
}
