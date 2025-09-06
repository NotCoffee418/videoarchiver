package download

import (
	"database/sql"
	"testing"
	_ "modernc.org/sqlite"
)

func TestCheckForDuplicateInDownloads(t *testing.T) {
	// Create a temporary database for testing
	tmpDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer tmpDB.Close()
	
	// Create the downloads table manually for testing
	_, err = tmpDB.Exec(`
		CREATE TABLE downloads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER NOT NULL,
			url TEXT NOT NULL,
			status INTEGER NOT NULL,
			format_downloaded TEXT,
			md5 TEXT,
			output_filename TEXT,
			last_attempt INTEGER,
			fail_message TEXT,
			attempt_count INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create downloads table: %v", err)
	}

	// Create a DownloadDB instance directly with the test database
	downloadDB := &DownloadDB{db: tmpDB}

	testMD5 := "test-md5-hash-123"

	// Test 1: No duplicate exists - should return false
	hasDuplicate, err := downloadDB.CheckForDuplicateInDownloads(testMD5)
	if err != nil {
		t.Errorf("Error checking for duplicates: %v", err)
	}
	if hasDuplicate {
		t.Error("Expected no duplicate, but found one")
	}

	// Test 2: Insert a failed download with MD5 - should NOT be considered a duplicate
	_, err = tmpDB.Exec(
		"INSERT INTO downloads (playlist_id, url, status, md5, attempt_count) VALUES (?, ?, ?, ?, ?)",
		1, "https://example.com/video1", StFailedAutoRetry, testMD5, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert failed download: %v", err)
	}

	hasDuplicate, err = downloadDB.CheckForDuplicateInDownloads(testMD5)
	if err != nil {
		t.Errorf("Error checking for duplicates: %v", err)
	}
	if hasDuplicate {
		t.Error("Failed download should NOT be considered a duplicate")
	}

	// Test 3: Insert another failed download with different status - should NOT be considered a duplicate
	_, err = tmpDB.Exec(
		"INSERT INTO downloads (playlist_id, url, status, md5, attempt_count) VALUES (?, ?, ?, ?, ?)",
		1, "https://example.com/video2", StFailedGiveUp, testMD5, 6,
	)
	if err != nil {
		t.Fatalf("Failed to insert failed download: %v", err)
	}

	hasDuplicate, err = downloadDB.CheckForDuplicateInDownloads(testMD5)
	if err != nil {
		t.Errorf("Error checking for duplicates: %v", err)
	}
	if hasDuplicate {
		t.Error("Failed download (give up) should NOT be considered a duplicate")
	}

	// Test 4: Insert a successful download - SHOULD be considered a duplicate
	_, err = tmpDB.Exec(
		"INSERT INTO downloads (playlist_id, url, status, md5, attempt_count) VALUES (?, ?, ?, ?, ?)",
		1, "https://example.com/video3", StSuccess, testMD5, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert successful download: %v", err)
	}

	hasDuplicate, err = downloadDB.CheckForDuplicateInDownloads(testMD5)
	if err != nil {
		t.Errorf("Error checking for duplicates: %v", err)
	}
	if !hasDuplicate {
		t.Error("Successful download SHOULD be considered a duplicate")
	}

	// Test 5: Test with StSuccessDuplicate status - SHOULD be considered a duplicate
	testMD5_2 := "test-md5-hash-456"
	_, err = tmpDB.Exec(
		"INSERT INTO downloads (playlist_id, url, status, md5, attempt_count) VALUES (?, ?, ?, ?, ?)",
		1, "https://example.com/video4", StSuccessDuplicate, testMD5_2, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert success duplicate download: %v", err)
	}

	hasDuplicate, err = downloadDB.CheckForDuplicateInDownloads(testMD5_2)
	if err != nil {
		t.Errorf("Error checking for duplicates: %v", err)
	}
	if !hasDuplicate {
		t.Error("Success duplicate download SHOULD be considered a duplicate")
	}
}