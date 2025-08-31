package db

import (
	"database/sql"
	"sync"
	"videoarchiver/backend/domains/config"
	"videoarchiver/backend/domains/logging"

	_ "modernc.org/sqlite"
)

type DatabaseService struct {
	db           *sql.DB
	configSvc    *config.ConfigService
	logSvc       *logging.LogService
}

var (
	dbInstance *DatabaseService
	once       sync.Once // ensure singleton of `db` even with multiple instances of service
)

func NewDatabaseService(configSvc *config.ConfigService, logSvc *logging.LogService) (*DatabaseService, error) {
	var errInit error
	once.Do(func() {
		dbPath, err := configSvc.GetDatabasePath()
		if err != nil {
			errInit = err
			return
		}

		// Log the database location
		if logSvc != nil {
			logSvc.Info("Database location: " + dbPath)
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

		dbInstance = &DatabaseService{
			db:        db,
			configSvc: configSvc,
			logSvc:    logSvc,
		}
	})

	if errInit != nil {
		return nil, errInit
	}
	return dbInstance, nil
}

func (d *DatabaseService) GetDB() *sql.DB {
	return d.db
}
