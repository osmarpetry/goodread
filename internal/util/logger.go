package util

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel represents the logging level.
type LogLevel int

const (
	// LogLevelQuiet suppresses all logs except errors.
	LogLevelQuiet LogLevel = iota
	// LogLevelNormal shows info and errors.
	LogLevelNormal
	// LogLevelVerbose shows debug, info, and errors.
	LogLevelVerbose
)

// Logger provides structured logging with levels.
type Logger struct {
	level      LogLevel
	infoLogger *log.Logger
	errLogger  *log.Logger
}

// NewLogger creates a new logger with the specified level.
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:      level,
		infoLogger: log.New(os.Stdout, "", 0),
		errLogger:  log.New(os.Stderr, "ERROR: ", 0),
	}
}

// NewQuietLogger creates a logger that only outputs errors.
func NewQuietLogger() *Logger {
	return NewLogger(LogLevelQuiet)
}

// SetOutput sets the output writer for info logs.
func (l *Logger) SetOutput(w io.Writer) {
	l.infoLogger.SetOutput(w)
}

// Debug logs a debug message (only in verbose mode).
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level >= LogLevelVerbose {
		l.infoLogger.Printf("DEBUG: "+format, v...)
	}
}

// Info logs an info message.
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level >= LogLevelNormal {
		l.infoLogger.Printf(format, v...)
	}
}

// Error logs an error message.
func (l *Logger) Error(format string, v ...interface{}) {
	l.errLogger.Printf(format, v...)
}

// Fatal logs a fatal error and exits.
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.errLogger.Printf(format, v...)
	os.Exit(1)
}

// Success logs a success message with a checkmark.
func (l *Logger) Success(format string, v ...interface{}) {
	if l.level >= LogLevelNormal {
		l.infoLogger.Printf("✓ "+format, v...)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level >= LogLevelNormal {
		l.infoLogger.Printf("⚠ "+format, v...)
	}
}

// Progress logs a progress message (overwrites the previous line in verbose mode).
func (l *Logger) Progress(format string, v ...interface{}) {
	if l.level >= LogLevelVerbose {
		fmt.Printf("\r"+format, v...)
	}
}
