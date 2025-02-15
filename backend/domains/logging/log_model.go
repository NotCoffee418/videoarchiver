package logging

import "time"

type Log struct {
	ID        int       `json:"id" db:"id"`
	Verbosity int       `json:"verbosity" db:"verbosity"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Message   string    `json:"message" db:"message"`
}

// LogVerbosity represents log levels
const (
	Debug   = 0
	Info    = 1
	Warning = 2
	Error   = 3
)
