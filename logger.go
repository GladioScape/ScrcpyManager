package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarning
	LogError
)

var logLevelNames = map[LogLevel]string{
	LogDebug:   "DEBUG",
	LogInfo:    "INFO",
	LogWarning: "WARN",
	LogError:   "ERROR",
}

// Logger provides structured logging with different levels
type Logger struct {
	mu      sync.Mutex
	level   LogLevel
	file    *os.File
	logger  *log.Logger
	console bool
}

var appLogger *Logger

// InitLogger initializes the application logger
func InitLogger(level LogLevel, enableConsole bool) error {
	logDir := filepath.Join(os.Getenv("TEMP"), "ScrcpyManager")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("scrcpy-manager-%s.log",
		time.Now().Format("2006-01-02")))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	appLogger = &Logger{
		level:   level,
		file:    file,
		logger:  log.New(file, "", 0),
		console: enableConsole,
	}

	appLogger.Info("Logger initialized", "file", logFile)
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if appLogger != nil && appLogger.file != nil {
		appLogger.file.Close()
	}
}

// log writes a log message at the specified level
func (l *Logger) log(level LogLevel, message string, keyvals ...interface{}) {
	if l == nil || level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := logLevelNames[level]

	logMsg := fmt.Sprintf("[%s] [%s] %s", timestamp, levelStr, message)

	// Add key-value pairs
	if len(keyvals) > 0 {
		logMsg += " |"
		for i := 0; i < len(keyvals); i += 2 {
			if i+1 < len(keyvals) {
				logMsg += fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1])
			}
		}
	}

	if l.logger != nil {
		l.logger.Println(logMsg)
	}

	if l.console {
		fmt.Println(logMsg)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, keyvals ...interface{}) {
	l.log(LogDebug, message, keyvals...)
}

// Info logs an info message
func (l *Logger) Info(message string, keyvals ...interface{}) {
	l.log(LogInfo, message, keyvals...)
}

// Warning logs a warning message
func (l *Logger) Warning(message string, keyvals ...interface{}) {
	l.log(LogWarning, message, keyvals...)
}

// Error logs an error message
func (l *Logger) Error(message string, keyvals ...interface{}) {
	l.log(LogError, message, keyvals...)
}

// Global logging helper functions
func Log(level LogLevel, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if appLogger != nil {
		appLogger.log(level, message)
	}
}
