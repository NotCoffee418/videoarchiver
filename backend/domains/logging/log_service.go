package logging

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type LogService struct {
	logDB  *LogDB
	logger *logrus.Logger
}

func NewLogService(logDB *LogDB) *LogService {
	logger := logrus.New()

	// Optional: Write logs to a file
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logger.SetOutput(file)
	} else {
		logger.SetOutput(os.Stdout)
	}

	logger.SetFormatter(&logrus.JSONFormatter{}) // Structured logs
	logger.SetLevel(logrus.DebugLevel)           // Default level

	return &LogService{
		logDB:  logDB,
		logger: logger,
	}
}

// Logs to database, stdout, and file (if enabled)
func (l *LogService) Log(verbosity logrus.Level, message string) {
	timestamp := time.Now()
	logEntry := l.logger.WithFields(logrus.Fields{
		"verbosity": verbosity,
		"timestamp": timestamp.Format(time.RFC3339),
	})

	logEntry.Log(verbosity, message)

	// Store in database if logDB is available
	if l.logDB != nil {
		l.logDB.AddLog(int(verbosity), message)
	}
}
