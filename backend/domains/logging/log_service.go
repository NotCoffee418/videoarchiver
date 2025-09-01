package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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

// Fatal logs a fatal level message and exits
func (l *LogService) Fatal(message string) {
	l.Log(logrus.FatalLevel, message)
	os.Exit(1)
}

// GetLogLinesFromFile reads the last N lines from a log file using proper pathing
func (l *LogService) GetLogLinesFromFile(filename string, lines int) ([]string, error) {
	// Get proper log file path using pathing system
	logFilePath, err := pathing.GetWorkingFile(filename)
	if err != nil {
		return []string{fmt.Sprintf("Error getting log file path: %v", err)}, nil
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"Log file does not exist yet"}, nil
		}
		return nil, fmt.Errorf("failed to open log file %s: %v", logFilePath, err)
	}
	defer file.Close()

	// Read all content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file %s: %v", logFilePath, err)
	}

	if len(content) == 0 {
		return []string{"Log file is empty"}, nil
	}

	// Split into lines and get the last N lines
	allLines := strings.Split(string(content), "\n")

	// Remove empty last line if it exists
	if len(allLines) > 0 && allLines[len(allLines)-1] == "" {
		allLines = allLines[:len(allLines)-1]
	}

	// Get last N lines
	startIndex := 0
	if len(allLines) > lines {
		startIndex = len(allLines) - lines
	}

	result := allLines[startIndex:]
	if len(result) == 0 {
		return []string{"No log entries found"}, nil
	}

	return result, nil
}

// GetLogLinesFromFileWithLevel reads the last N lines from a log file and filters by minimum log level
func (l *LogService) GetLogLinesFromFileWithLevel(filename string, lines int, minLevelStr string) ([]string, error) {
	// Parse minimum log level
	var minLevel logrus.Level
	switch strings.ToLower(minLevelStr) {
	case "debug":
		minLevel = logrus.DebugLevel
	case "info":
		minLevel = logrus.InfoLevel
	case "warn", "warning":
		minLevel = logrus.WarnLevel
	case "error":
		minLevel = logrus.ErrorLevel
	case "fatal":
		minLevel = logrus.FatalLevel
	default:
		minLevel = logrus.InfoLevel // Default to info
	}

	// Get proper log file path using pathing system
	logFilePath, err := pathing.GetWorkingFile(filename)
	if err != nil {
		return []string{fmt.Sprintf("Error getting log file path: %v", err)}, nil
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"Log file does not exist yet"}, nil
		}
		return nil, fmt.Errorf("failed to open log file %s: %v", logFilePath, err)
	}
	defer file.Close()

	// Read all content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file %s: %v", logFilePath, err)
	}

	if len(content) == 0 {
		return []string{"Log file is empty"}, nil
	}

	// Split into lines
	allLines := strings.Split(string(content), "\n")

	// Remove empty last line if it exists
	if len(allLines) > 0 && allLines[len(allLines)-1] == "" {
		allLines = allLines[:len(allLines)-1]
	}

	// Filter lines by log level
	var filteredLines []string
	for _, line := range allLines {
		if line == "" {
			continue
		}

		// Try to parse JSON log entry
		var logEntry struct {
			Level string `json:"level"`
		}

		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			// If not JSON, include the line (could be plain text log)
			filteredLines = append(filteredLines, line)
			continue
		}

		// Parse log level from JSON
		var logLevel logrus.Level
		switch strings.ToLower(logEntry.Level) {
		case "debug":
			logLevel = logrus.DebugLevel
		case "info":
			logLevel = logrus.InfoLevel
		case "warn", "warning":
			logLevel = logrus.WarnLevel
		case "error":
			logLevel = logrus.ErrorLevel
		case "fatal":
			logLevel = logrus.FatalLevel
		default:
			logLevel = logrus.InfoLevel
		}

		// Include line if it meets minimum level requirement
		// In logrus, lower numbers = higher severity, so check <=
		if logLevel <= minLevel {
			filteredLines = append(filteredLines, line)
		}
	}

	// Get last N filtered lines
	startIndex := 0
	if len(filteredLines) > lines {
		startIndex = len(filteredLines) - lines
	}

	result := filteredLines[startIndex:]
	if len(result) == 0 {
		return []string{"No log entries found matching the specified level"}, nil
	}

	return result, nil
}
