package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

type Logger struct {
	*log.Logger
}

var (
	Info  *Logger
	Error *Logger
	Debug *Logger
)

func init() {
	Info = &Logger{log.New(os.Stdout, "", 0)}
	Error = &Logger{log.New(os.Stderr, "", 0)}
	Debug = &Logger{log.New(os.Stdout, "", 0)}
}

func (l *Logger) logWithContext(level string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Extract just the filename from the full path
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s:%d] [%s]", timestamp, file, line, level)

	message := fmt.Sprint(v...)
	l.Logger.Printf("%s %s", prefix, message)
}

func (l *Logger) logWithContextf(level string, format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Extract just the filename from the full path
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s:%d] [%s]", timestamp, file, line, level)

	message := fmt.Sprintf(format, v...)
	l.Logger.Printf("%s %s", prefix, message)
}

func (l *Logger) logErrorWithTraceback(v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Extract just the filename from the full path
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s:%d] [ERROR]", timestamp, file, line)

	message := fmt.Sprint(v...)
	l.Logger.Printf("%s %s", prefix, message)

	// Print stack trace
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	l.Logger.Printf("[%s] [%s:%d] [TRACEBACK]\n%s", timestamp, file, line, string(buf))
}

// Info logging methods
func (l *Logger) Info(v ...interface{}) {
	l.logWithContext("INFO", v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.logWithContextf("INFO", format, v...)
}

// Error logging methods
func (l *Logger) Error(v ...interface{}) {
	l.logErrorWithTraceback(v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Extract just the filename from the full path
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s:%d] [ERROR]", timestamp, file, line)

	message := fmt.Sprintf(format, v...)
	l.Logger.Printf("%s %s", prefix, message)

	// Print stack trace
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	l.Logger.Printf("[%s] [%s:%d] [TRACEBACK]\n%s", timestamp, file, line, string(buf))
}

// Debug logging methods
func (l *Logger) Debug(v ...interface{}) {
	l.logWithContext("DEBUG", v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logWithContextf("DEBUG", format, v...)
}

// Convenience functions
func LogInfo(v ...interface{}) {
	Info.Info(v...)
}

func LogInfof(format string, v ...interface{}) {
	Info.Infof(format, v...)
}

func LogError(v ...interface{}) {
	Error.Error(v...)
}

func LogErrorf(format string, v ...interface{}) {
	Error.Errorf(format, v...)
}

func LogDebug(v ...interface{}) {
	Debug.Debug(v...)
}

func LogDebugf(format string, v ...interface{}) {
	Debug.Debugf(format, v...)
}
