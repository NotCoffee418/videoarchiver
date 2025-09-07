package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckFileCorruption(t *testing.T) {
	// Create a temporary test file with some content
	tmpDir, err := os.MkdirTemp("", "corruption_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with a simple text file (this should fail since it's not a valid media file)
	textFilePath := filepath.Join(tmpDir, "test.mp4")
	err = os.WriteFile(textFilePath, []byte("This is not a valid video file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test corruption check on invalid file - should return error
	err = CheckFileCorruption(textFilePath)
	if err == nil {
		t.Error("Expected corruption check to fail on invalid file, but it passed")
	}

	// Test with non-existent file - should return error
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.mp4")
	err = CheckFileCorruption(nonExistentPath)
	if err == nil {
		t.Error("Expected corruption check to fail on non-existent file, but it passed")
	}
}
