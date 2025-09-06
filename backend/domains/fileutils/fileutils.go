package fileutils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// CalculateMD5 calculates the MD5 hash of a file
func CalculateMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// sanitizeFilename replaces filesystem-unsafe characters with underscores
// while preserving Unicode characters like emojis and foreign text.
// Handles control characters, invisible characters, and caps filename length to 48 characters.
func SanitizeFilename(filename string) string {
	// Limit string length to 48
	if len(filename) > 48 {
		filename = filename[:48]
	}

	// Remove control characters (ASCII 0-31) including NULL byte
	re := regexp.MustCompile(`[\x00-\x1f]`)
	filename = re.ReplaceAllString(filename, "")

	// Remove invisible Unicode characters
	// Zero-width spaces and other invisible characters
	invalidChars := []string{
		// Invisible Characters
		"\u200B", // Zero Width Space
		"\u200C", // Zero Width Non-Joiner
		"\u200D", // Zero Width Joiner
		"\u200E", // Left-to-Right Mark
		"\u200F", // Right-to-Left Mark
		"\u202A", // Left-to-Right Embedding
		"\u202B", // Right-to-Left Embedding
		"\u202C", // Pop Directional Formatting
		"\u202D", // Left-to-Right Override
		"\u202E", // Right-to-Left Override
		"\u2060", // Word Joiner
		"\u2061", // Function Application
		"\u2062", // Invisible Times
		"\u2063", // Invisible Separator
		"\u2064", // Invisible Plus
		"\uFEFF", // Zero Width No-Break Space (BOM)

		// Windows reserved characters
		"<", ">", ":", "\"", "/", "\\", "|", "?", "*",
	}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// Make filename _ if we have no characters left
	if filename == "" {
		filename = "_"
	}

	return filename
}
