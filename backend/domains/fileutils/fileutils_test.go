package fileutils

import (
	"os"
	"path/filepath"
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
			expected: "Mix_of_many_bad_chars__like_this_and_that",
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
			name:     "Leading and trailing spaces/underscores as is",
			input:    " __Test String__ ",
			expected: " __Test String__ ",
		},
		{
			name:     "Invisible Unicode characters removed",
			input:    "Test\u200BZero\u200CWidth\u200DChars\uFEFF",
			expected: "Test_Zero_Width_Chars_",
		},
		{
			name:     "Windows reserved characters",
			input:    "CON.txt becomes safe",
			expected: "CON.txt becomes safe",
		},
		{
			name:     "Too Long and Complex mixed example",
			input:    "Long  _/\\Mix<of>:many\"bad|chars?and*invisible\u200Bchars\x00\x1f  _",
			expected: "Long  ___Mix_of__many_bad_chars_and_invisible_", // cuts off at 48 chars
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "_",
		},
		{
			name:     "Only invalid characters",
			input:    "/\\<>:\"|?*",
			expected: "_________",
		},
		{
			name:     "Only control characters",
			input:    "\x00\x01\x1f",
			expected: "_",
		},
		{
			name:     "Only spaces and underscores",
			input:    "   ___   ",
			expected: "   ___   ",
		},
		{
			name:     "Only invisible characters",
			input:    "\u200B\u200C\u200D\uFEFF",
			expected: "____",
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

	tmpDir, err := os.MkdirTemp("", "sanitize_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check expected output
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}

			// Test actual file creation
			filePath := filepath.Join(tmpDir, result+".tmp")
			file, err := os.Create(filePath)
			if err != nil {
				t.Errorf("Failed to create file with sanitized name %q: %v", tt.input, err)
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
		})
	}
}
