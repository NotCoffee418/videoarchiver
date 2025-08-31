package logging

import (
	"database/sql"
	"time"
	"videoarchiver/backend/domains/db"
)

type LogDB struct {
	db *sql.DB
}

func NewLogDB(dbService *db.DatabaseService) *LogDB {
	return &LogDB{db: dbService.GetDB()}
}

// GetLogs retrieves logs with a minimum verbosity level and optional limit
func (l *LogDB) GetLogs(minVerbosity, limit int) ([]Log, error) {
	rows, err := l.db.Query("SELECT * FROM logs WHERE verbosity >= ? ORDER BY timestamp DESC LIMIT ?", minVerbosity, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var logEntry Log
		err := rows.Scan(&logEntry.ID, &logEntry.Verbosity, &logEntry.Timestamp, &logEntry.Message)
		if err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}
	return logs, nil
}

// AddLog inserts a log entry into the database
func (l *LogDB) AddLog(verbosity int, message string) error {
	timestamp := time.Now()

	_, err := l.db.Exec("INSERT INTO logs (verbosity, timestamp, message) VALUES (?, ?, ?)", verbosity, timestamp.Unix(), message)
	return err
}
