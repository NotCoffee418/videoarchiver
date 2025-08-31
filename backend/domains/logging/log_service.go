package logging

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type LogService struct {
	logDB  *LogDB
	logger *logrus.Logger
	mode   string
}

// NewLogService creates a new log service with mode-specific log files
// mode should be "daemon" or "ui"
func NewLogService(logDB *LogDB, mode string) *LogService {
	logger := logrus.New()

	// Create mode-specific log file
	var logFileName string
	if mode == "daemon" {
		logFileName = "daemon.log"
	} else {
		logFileName = "ui.log"
	}

	// Write logs to both file and stdout
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		// Multi-writer to write to both file and stdout
		multiWriter := io.MultiWriter(file, os.Stdout)
		logger.SetOutput(multiWriter)
	} else {
		logger.SetOutput(os.Stdout)
	}

	logger.SetFormatter(&logrus.JSONFormatter{}) // Structured logs
	logger.SetLevel(logrus.DebugLevel)           // Default level

	return &LogService{
		logDB:  logDB,
		logger: logger,
		mode:   mode,
	}
}

// Logs to database, stdout, and file (if enabled)
func (l *LogService) Log(verbosity logrus.Level, message string) {
	timestamp := time.Now()
	logEntry := l.logger.WithFields(logrus.Fields{
		"verbosity": verbosity,
		"timestamp": timestamp.Format(time.RFC3339),
		"mode":      l.mode,
	})

	logEntry.Log(verbosity, message)

	// Store in database if logDB is available
	if l.logDB != nil {
		l.logDB.AddLog(int(verbosity), message)
	}
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
