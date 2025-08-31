package logging

import (
	"io"
	"os"
	"sync"
	"time"
	"videoarchiver/backend/domains/pathing"

	"github.com/sirupsen/logrus"
)

type LogService struct {
	logger *logrus.Logger
	mode   string
	file   *os.File
	mu     sync.Mutex
}

// NewLogService creates a new log service with mode-specific log files
// mode should be "daemon" or "ui"
func NewLogService(mode string) *LogService {
	logger := logrus.New()

	// Create mode-specific log file using proper pathing
	var logFileName string
	if mode == "daemon" {
		logFileName = "daemon.log"
	} else {
		logFileName = "ui.log"
	}

	// Get proper log file path using pathing system
	logFilePath, err := pathing.GetWorkingFile(logFileName)
	if err != nil {
		// Fallback to stdout only if pathing fails
		logger.SetOutput(os.Stdout)
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.DebugLevel)
		
		service := &LogService{
			logger: logger,
			mode:   mode,
			file:   nil,
		}
		service.Info("LogService initialized for mode: " + mode + " (file logging disabled due to pathing error)")
		return service
	}

	// Write logs to both file and stdout
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	var multiWriter io.Writer
	if err == nil {
		// Multi-writer to write to both file and stdout
		multiWriter = io.MultiWriter(file, os.Stdout)
		logger.SetOutput(multiWriter)
	} else {
		logger.SetOutput(os.Stdout)
		file = nil
	}

	logger.SetFormatter(&logrus.JSONFormatter{}) // Structured logs
	logger.SetLevel(logrus.DebugLevel)           // Default level

	service := &LogService{
		logger: logger,
		mode:   mode,
		file:   file,
	}

	// Test log entry to verify logging is working
	service.Info("LogService initialized for mode: " + mode)
	
	return service
}

// Logs to database, stdout, and file (if enabled)
func (l *LogService) Log(verbosity logrus.Level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	timestamp := time.Now()
	logEntry := l.logger.WithFields(logrus.Fields{
		"verbosity": verbosity,
		"timestamp": timestamp.Format(time.RFC3339),
		"mode":      l.mode,
	})

	logEntry.Log(verbosity, message)

	// Force sync to disk if file is available
	if l.file != nil {
		l.file.Sync()
	}
}

// Close closes the log file
func (l *LogService) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Info logs an info level message
func (l *LogService) Info(message string) {
	l.Log(logrus.InfoLevel, message)
}

// Debug logs a debug level message
func (l *LogService) Debug(message string) {
	l.Log(logrus.DebugLevel, message)
}

// Warn logs a warning level message
func (l *LogService) Warn(message string) {
	l.Log(logrus.WarnLevel, message)
}

// Error logs an error level message
func (l *LogService) Error(message string) {
	l.Log(logrus.ErrorLevel, message)
}
