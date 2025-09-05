package fileregistry

import (
	"database/sql"
	"time"
	"videoarchiver/backend/domains/db"
)

type FileRegistryService struct {
	db *sql.DB
}

func NewFileRegistryService(dbService *db.DatabaseService) *FileRegistryService {
	return &FileRegistryService{db: dbService.GetDB()}
}

// GetByMD5 returns the first registered file with the given MD5 hash
func (f *FileRegistryService) GetByMD5(md5Hash string) (*RegisteredFile, error) {
	row := f.db.QueryRow(
		"SELECT id, filename, file_path, md5, registered_at FROM file_registry WHERE md5 = ? LIMIT 1",
		md5Hash,
	)
	
	var file RegisteredFile
	err := row.Scan(&file.ID, &file.Filename, &file.FilePath, &file.MD5, &file.RegisteredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No duplicate found
		}
		return nil, err
	}
	
	return &file, nil
}

// RegisterFile adds a new file to the registry
func (f *FileRegistryService) RegisterFile(filename, filePath, md5Hash string) error {
	_, err := f.db.Exec(
		"INSERT INTO file_registry (filename, file_path, md5, registered_at) VALUES (?, ?, ?, ?)",
		filename, filePath, md5Hash, time.Now().Unix(),
	)
	return err
}

// GetAllPaginated returns a paginated list of registered files
func (f *FileRegistryService) GetAllPaginated(offset, limit int) ([]RegisteredFile, error) {
	rows, err := f.db.Query(
		"SELECT id, filename, file_path, md5, registered_at FROM file_registry ORDER BY registered_at DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var files []RegisteredFile
	for rows.Next() {
		var file RegisteredFile
		err := rows.Scan(&file.ID, &file.Filename, &file.FilePath, &file.MD5, &file.RegisteredAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	
	return files, nil
}

// ClearAll removes all registered files from the database
func (f *FileRegistryService) ClearAll() error {
	_, err := f.db.Exec("DELETE FROM file_registry")
	return err
}