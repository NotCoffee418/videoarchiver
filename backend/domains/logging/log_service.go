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

// ClearLogsOlderThanDays removes log entries older than the specified number of days
// It modifies the original file in place by truncating and rewriting filtered content
func (l *LogService) ClearLogsOlderThanDays(filename string, days int) error {
	// Get proper log file path using pathing system
	logFilePath, err := pathing.GetWorkingFile(filename)
	if err != nil {
		return fmt.Errorf("failed to get log file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		// File doesn't exist, nothing to clean
		return nil
	}

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Open the original file for reading
	file, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFilePath, err)
	}
	defer file.Close()

	// Read all content
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read log file %s: %w", logFilePath, err)
	}

	if len(content) == 0 {
		// Empty file, nothing to clean
		return nil
	}

	// Split into lines and filter by date
	allLines := strings.Split(string(content), "\n")
	var filteredLines []string
	entriesRemoved := 0

	for _, line := range allLines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		// Try to parse JSON log entry to get timestamp
		var logEntry struct {
			Time string `json:"time"`
		}

		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			// If not valid JSON, keep the line (could be plain text log or corrupted entry)
			filteredLines = append(filteredLines, line)
			continue
		}

		// Parse the timestamp
		if logEntry.Time == "" {
			// No timestamp field, keep the line
			filteredLines = append(filteredLines, line)
			continue
		}

		logTime, err := time.Parse(time.RFC3339, logEntry.Time)
		if err != nil {
			// Invalid timestamp format, keep the line
			filteredLines = append(filteredLines, line)
			continue
		}

		// Keep entries newer than cutoff date
		if logTime.After(cutoffDate) {
			filteredLines = append(filteredLines, line)
		} else {
			entriesRemoved++
		}
	}

	// If no entries were removed, no need to rewrite the file
	if entriesRemoved == 0 {
		return nil
	}

	// Open the original file for writing (truncates the file)
	writeFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file for writing: %w", err)
	}
	defer writeFile.Close()

	// Write filtered content directly to the original file
	for i, line := range filteredLines {
		if i > 0 {
			if _, err := writeFile.WriteString("\n"); err != nil {
				return fmt.Errorf("failed to write to log file: %w", err)
			}
		}
		if _, err := writeFile.WriteString(line); err != nil {
			return fmt.Errorf("failed to write to log file: %w", err)
		}
	}

	// Add final newline if we have content
	if len(filteredLines) > 0 {
		if _, err := writeFile.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write final newline to log file: %w", err)
		}
	}

	return nil
}
