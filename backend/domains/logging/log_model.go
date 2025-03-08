package logging

type Log struct {
	ID        int    `json:"id" db:"id"`
	Verbosity int    `json:"verbosity" db:"verbosity"`
	Timestamp int64  `json:"timestamp" db:"timestamp"`
	Message   string `json:"message" db:"message"`
}

// LogVerbosity represents log levels
const (
	Debug   = 0
	Info    = 1
	Warning = 2
	Error   = 3
)
