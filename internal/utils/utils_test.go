package utils

import (
	"reflect"
	"testing"
)

func TestParseExtToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		desc     string
	}{
		// Empty and whitespace cases
		{
			name:     "empty_string",
			input:    "",
			expected: []string{},
			desc:     "Empty string should return empty slice",
		},
		{
			name:     "only_whitespace",
			input:    "   ",
			expected: []string{},
			desc:     "Only whitespace should return empty slice",
		},
		{
			name:     "only_commas",
			input:    ",,,",
			expected: []string{},
			desc:     "Only commas should return empty slice",
		},

		// Single extension cases
		{
			name:     "single_ext_without_dot",
			input:    "txt",
			expected: []string{".txt"},
			desc:     "Single extension without dot should add dot prefix and lowercase",
		},
		{
			name:     "single_ext_with_dot",
			input:    ".txt",
			expected: []string{".txt"},
			desc:     "Single extension with dot should keep dot and lowercase",
		},
		{
			name:     "single_ext_uppercase",
			input:    "TXT",
			expected: []string{".txt"},
			desc:     "Uppercase extension should be converted to lowercase",
		},
		{
			name:     "single_ext_mixed_case",
			input:    "JpG",
			expected: []string{".jpg"},
			desc:     "Mixed case extension should be converted to lowercase",
		},
		{
			name:     "single_ext_with_spaces",
			input:    "  txt  ",
			expected: []string{".txt"},
			desc:     "Extension with surrounding spaces should be trimmed",
		},

		// Multiple extension cases
		{
			name:     "multiple_exts_without_dots",
			input:    "txt,jpg,pdf",
			expected: []string{".txt", ".jpg", ".pdf"},
			desc:     "Multiple extensions without dots should all get dot prefix",
		},
		{
			name:     "multiple_exts_with_dots",
			input:    ".txt,.jpg,.pdf",
			expected: []string{".txt", ".jpg", ".pdf"},
			desc:     "Multiple extensions with dots should keep dots",
		},
		{
			name:     "multiple_exts_mixed_dots",
			input:    "txt,.jpg,pdf",
			expected: []string{".txt", ".jpg", ".pdf"},
			desc:     "Mix of extensions with and without dots should normalize all",
		},
		{
			name:     "multiple_exts_with_spaces",
			input:    "txt, jpg, pdf",
			expected: []string{".txt", ".jpg", ".pdf"},
			desc:     "Extensions with spaces after commas should be trimmed",
		},
		{
			name:     "multiple_exts_mixed_case",
			input:    "TXT,JpG,PDF",
			expected: []string{".txt", ".jpg", ".pdf"},
			desc:     "Mixed case extensions should all be lowercased",
		},

		// Edge cases with empty items
		{
			name:     "trailing_comma",
			input:    "txt,jpg,",
			expected: []string{".txt", ".jpg"},
			desc:     "Trailing comma should be ignored",
		},
		{
			name:     "leading_comma",
			input:    ",txt,jpg",
			expected: []string{".txt", ".jpg"},
			desc:     "Leading comma should be ignored",
		},
		{
			name:     "multiple_consecutive_commas",
			input:    "txt,,jpg",
			expected: []string{".txt", ".jpg"},
			desc:     "Multiple consecutive commas should be treated as empty items and ignored",
		},
		{
			name:     "comma_with_spaces",
			input:    "txt, , jpg",
			expected: []string{".txt", ".jpg"},
			desc:     "Comma with only spaces should be ignored",
		},

		// Special characters and realistic scenarios
		{
			name:     "programming_languages",
			input:    "go,js,ts,py,java",
			expected: []string{".go", ".js", ".ts", ".py", ".java"},
			desc:     "Common programming language extensions",
		},
		{
			name:     "compressed_files",
			input:    "zip,tar,gz,7z,rar",
			expected: []string{".zip", ".tar", ".gz", ".7z", ".rar"},
			desc:     "Common compressed file extensions including numbers",
		},
		{
			name:     "media_files",
			input:    "mp4,mp3,avi,mkv,jpg,png",
			expected: []string{".mp4", ".mp3", ".avi", ".mkv", ".jpg", ".png"},
			desc:     "Common media file extensions",
		},
		{
			name:     "double_extensions",
			input:    "tar.gz,min.js",
			expected: []string{".tar.gz", ".min.js"},
			desc:     "Double extensions should be preserved",
		},
		{
			name:     "excessive_whitespace",
			input:    "  txt  ,  jpg  ,  pdf  ",
			expected: []string{".txt", ".jpg", ".pdf"},
			desc:     "Excessive whitespace around items should be trimmed",
		},
		{
			name:     "mixed_case_and_dots",
			input:    "TXT, .JpG, Pdf, .DOC",
			expected: []string{".txt", ".jpg", ".pdf", ".doc"},
			desc:     "Mixed case and dot prefixes should be normalized",
		},

		// Duplicate handling (note: function doesn't deduplicate, so duplicates are preserved)
		{
			name:     "duplicate_extensions",
			input:    "txt,jpg,txt,pdf,jpg",
			expected: []string{".txt", ".jpg", ".txt", ".pdf", ".jpg"},
			desc:     "Duplicate extensions are preserved (no deduplication)",
		},

		// Single character extensions
		{
			name:     "multiple patterns with spaces",
			input:    "*.tmp, *.log, *.bak",
			expected: []string{"*.tmp", "*.log", "*.bak"},
		},

		// Duplicate handling
		{
			name:     "multiple patterns without spaces",
			input:    "*.tmp,*.log,*.bak",
			expected: []string{"*.tmp", "*.log", "*.bak"},
		},

		// Excessive whitespace
		{
			name:     "patterns with extra spaces",
			input:    "  *.tmp  ,  *.log  ,  *.bak  ",
			expected: []string{"*.tmp", "*.log", "*.bak"},
		},
		{
			name:     "empty entries",
			input:    "*.tmp,,*.log,",
			expected: []string{"*.tmp", "*.log"},
		},

		// Real-world exclusion scenarios
		{
			name:     "mixed patterns",
			input:    "*.tmp, folder/, .git/",
			expected: []string{"*.tmp", "folder/", ".git/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExcludeToSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseExcludeToSlice(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
