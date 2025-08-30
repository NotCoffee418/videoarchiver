package db

import (
	"database/sql"
	"path/filepath"
	"sync"
	"videoarchiver/backend/domains/pathing"

	_ "modernc.org/sqlite"
)

type DatabaseService struct {
	db *sql.DB
}

var (
	dbInstance *DatabaseService
	once       sync.Once // ensure singleton of `db` even with multiple instances of service
)

func NewDatabaseService() (*DatabaseService, error) {
	var errInit error
	once.Do(func() {
		dbPath, err := getDatabasePath()
		if err != nil {
			errInit = err
			return
		}

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			errInit = err
			return
		}

		// Set a busy timeout to handle database locks gracefully
		_, err = db.Exec("PRAGMA busy_timeout = 30000")
		if err != nil {
			errInit = err
			return
		}

		dbInstance = &DatabaseService{db: db}
	})

	if errInit != nil {
		return nil, errInit
	}
	return dbInstance, nil
}

func (d *DatabaseService) GetDB() *sql.DB {
	return d.db
}

func getDatabasePath() (string, error) {
	workingDir, err := pathing.GetWorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(workingDir, "db.sqlite"), nil
}
