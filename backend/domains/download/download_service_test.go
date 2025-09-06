package download

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Foreign characters preserved",
			input:    "正常中文字符",
			expected: "正常中文字符",
		},
		{
			name:     "Emojis preserved",
			input:    "Emoji test 🎥📹💿",
			expected: "Emoji test 🎥📹💿",
		},
		{
			name:     "Problematic chars replaced",
			input:    "Mix/of\\many:bad<chars>\"like|this?and*that",
			expected: "Mix_of_many_bad_chars_like_this_and_that",
		},
		{
			name:     "Path separators handled",
			input:    "Part 1/2: My Video",
			expected: "Part 1_2_ My Video",
		},
		{
			name:     "Control characters removed",
			input:    "Test\x00\x01\x1fString",
			expected: "TestString",
		},
		{
			name:     "Multiple underscores cleaned",
			input:    "Test___Multiple__Underscores_",
			expected: "Test_Multiple_Underscores",
		},
		{
			name:     "Leading and trailing spaces/underscores trimmed",
			input:    " __Test String__ ",
			expected: "Test String",
		},
		{
			name:     "Invisible Unicode characters removed",
			input:    "Test\u200BZero\u200CWidth\u200DChars\uFEFF",
			expected: "TestZeroWidthChars",
		},
		{
			name:     "Windows reserved characters",
			input:    "CON.txt becomes safe",
			expected: "CON.txt becomes safe",
		},
		{
			name:     "Complex mixed example",
			input:    "  _/\\Mix<of>:many\"bad|chars?and*invisible\u200Bchars\x00\x1f  _",
			expected: "Mix_of_many_bad_chars_and_invisiblechars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilenameWithLength(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		extension string
		maxLength int
		expected  string
	}{
		{
			name:      "Normal filename within limit",
			filename:  "Short Video Title",
			extension: "mp4",
			maxLength: 48,
			expected:  "Short Video Title",
		},
		{
			name:      "Long filename gets truncated",
			filename:  "This is a very long video title that exceeds the maximum allowed length",
			extension: "mp4",
			maxLength: 48,
			expected:  "This is a very long video title that exceeds", // 48 - 4 (.mp4) = 44 chars
		},
		{
			name:      "Extension with dot already",
			filename:  "Test Video",
			extension: ".avi",
			maxLength: 20,
			expected:  "Test Video", // 20 - 4 (.avi) = 16 chars, fits
		},
		{
			name:      "Very short max length",
			filename:  "Long Title",
			extension: "mp4",
			maxLength: 8,
			expected:  "Long", // 8 - 4 (.mp4) = 4 chars
		},
		{
			name:      "Filename with problematic chars and length limit",
			filename:  "Video/with\\bad:chars<and>very|long?title*here",
			extension: "mp4",
			maxLength: 25,
			expected:  "Video_with_bad_chars", // Sanitized and truncated to 25 - 4 = 21 chars, then trimmed
		},
		{
			name:      "Extension too long fallback",
			filename:  "Test",
			extension: "verylongextension",
			maxLength: 10,
			expected:  "Test", // Uses .tmp fallback, so 10 - 4 = 6 chars available
		},
		{
			name:      "Truncation ends with underscore cleanup",
			filename:  "Video_title_with_underscores_at_end",
			extension: "mp4",
			maxLength: 20,
			expected:  "Video_title_with", // Truncated and cleaned trailing underscore
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilenameWithLength(tt.filename, tt.extension, tt.maxLength)
			if result != tt.expected {
				t.Errorf("sanitizeFilenameWithLength(%q, %q, %d) = %q, want %q", 
					tt.filename, tt.extension, tt.maxLength, result, tt.expected)
			}
		})
	}
}

func TestFileCreationWithSanitizedNames(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sanitize_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		filename     string
		extension    string
		shouldCreate bool
	}{
		{
			name:         "Safe filename",
			filename:     "Normal Video Title",
			extension:    "mp4",
			shouldCreate: true,
		},
		{
			name:         "Filename with problematic chars",
			filename:     "Video/with\\bad:chars<and>problems|here?and*there",
			extension:    "avi",
			shouldCreate: true,
		},
		{
			name:         "Foreign characters",
			filename:     "正常中文字符 Video 🎥",
			extension:    "mkv",
			shouldCreate: true,
		},
		{
			name:         "Long filename",
			filename:     "This is a very long video title that should be truncated appropriately",
			extension:    "mp4",
			shouldCreate: true,
		},
		{
			name:         "Invisible characters",
			filename:     "Video\u200BWith\u200CInvisible\u200DChars\uFEFF",
			extension:    "webm",
			shouldCreate: true,
		},
		{
			name:         "Control characters",
			filename:     "Video\x00With\x01Control\x1fChars",
			extension:    "mov",
			shouldCreate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sanitize the filename with length limit
			sanitizedBase := sanitizeFilenameWithLength(tt.filename, tt.extension, 48)
			fullFilename := sanitizedBase + "." + tt.extension
			
			// Ensure the full filename is within the 48 character limit
			if len(fullFilename) > 48 {
				t.Errorf("Full filename %q exceeds 48 character limit (length: %d)", fullFilename, len(fullFilename))
			}
			
			// Create the full file path
			filePath := filepath.Join(tmpDir, fullFilename)
			
			// Try to create the file
			file, err := os.Create(filePath)
			if tt.shouldCreate && err != nil {
				t.Errorf("Failed to create file with sanitized name %q: %v", fullFilename, err)
				return
			}
			if !tt.shouldCreate && err == nil {
				t.Errorf("Expected file creation to fail for %q, but it succeeded", fullFilename)
				file.Close()
				return
			}
			
			if file != nil {
				file.Close()
				
				// Verify the file was created successfully
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("File was not created successfully: %s", filePath)
				}
				
				// Clean up individual test file
				os.Remove(filePath)
			}
			
			// Additional validation for the sanitized filename
			if strings.ContainsAny(sanitizedBase, "<>:\"/\\|?*") {
				t.Errorf("Sanitized filename %q still contains invalid characters", sanitizedBase)
			}
			
			// Check for control characters (ASCII 0-31)
			for i := 0; i < 32; i++ {
				if strings.Contains(sanitizedBase, string(rune(i))) {
					t.Errorf("Sanitized filename %q still contains control character %d", sanitizedBase, i)
				}
			}
			
			// Check for invisible Unicode characters
			invisibleChars := []string{"\u200B", "\u200C", "\u200D", "\uFEFF"}
			for _, char := range invisibleChars {
				if strings.Contains(sanitizedBase, char) {
					t.Errorf("Sanitized filename %q still contains invisible character %U", sanitizedBase, []rune(char)[0])
				}
			}
		})
	}
}

func TestSanitizeFilenameEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only invalid characters",
			input:    "/\\<>:\"|?*",
			expected: "",
		},
		{
			name:     "Only control characters",
			input:    "\x00\x01\x1f",
			expected: "",
		},
		{
			name:     "Only spaces and underscores",
			input:    "   ___   ",
			expected: "",
		},
		{
			name:     "Only invisible characters",
			input:    "\u200B\u200C\u200D\uFEFF",
			expected: "",
		},
		{
			name:     "Single valid character",
			input:    "a",
			expected: "a",
		},
		{
			name:     "Unicode emoji only",
			input:    "🎥",
			expected: "🎥",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilenameWithLengthEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		extension string
		maxLength int
		expected  string
	}{
		{
			name:      "Zero max length",
			filename:  "Test",
			extension: "mp4",
			maxLength: 0,
			expected:  "T", // At least 1 char for base name
		},
		{
			name:      "Max length smaller than extension",
			filename:  "Test",
			extension: "mp4",
			maxLength: 2,
			expected:  "T", // Falls back to .tmp, leaves minimal space
		},
		{
			name:      "Extension longer than max length",
			filename:  "Test",
			extension: "verylongextension",
			maxLength: 5,
			expected:  "T", // Uses .tmp fallback
		},
		{
			name:      "Empty filename",
			filename:  "",
			extension: "mp4",
			maxLength: 10,
			expected:  "",
		},
		{
			name:      "Filename becomes empty after sanitization",
			filename:  "/\\<>:\"|?*",
			extension: "mp4",
			maxLength: 10,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilenameWithLength(tt.filename, tt.extension, tt.maxLength)
			if result != tt.expected {
				t.Errorf("sanitizeFilenameWithLength(%q, %q, %d) = %q, want %q", 
					tt.filename, tt.extension, tt.maxLength, result, tt.expected)
			}
		})
	}
}