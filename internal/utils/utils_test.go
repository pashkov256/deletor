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
			name:     "single_char_extension",
			input:    "c,h,o",
			expected: []string{".c", ".h", ".o"},
			desc:     "Single character extensions should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExtToSlice(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("\nTest: %s\nDescription: %s\nInput: %q\nExpected: %v\nGot: %v",
					tt.name, tt.desc, tt.input, tt.expected, result)
			}
		})
	}
}

func TestParseExcludeToSlice(t *testing.T) {
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

		// Single pattern cases
		{
			name:     "single_pattern",
			input:    "data",
			expected: []string{"data"},
			desc:     "Single pattern should be returned as-is",
		},
		{
			name:     "single_pattern_with_spaces",
			input:    "  data  ",
			expected: []string{"data"},
			desc:     "Pattern with surrounding spaces should be trimmed",
		},
		{
			name:     "single_path_pattern",
			input:    "node_modules",
			expected: []string{"node_modules"},
			desc:     "Path-like pattern should be preserved",
		},

		// Multiple pattern cases
		{
			name:     "multiple_patterns",
			input:    "data,backup,temp",
			expected: []string{"data", "backup", "temp"},
			desc:     "Multiple patterns should be split correctly",
		},
		{
			name:     "multiple_patterns_with_spaces",
			input:    "data, backup, temp",
			expected: []string{"data", "backup", "temp"},
			desc:     "Patterns with spaces after commas should be trimmed",
		},
		{
			name:     "patterns_with_varied_spacing",
			input:    "  data  ,backup,  temp  ",
			expected: []string{"data", "backup", "temp"},
			desc:     "Varied spacing around patterns should all be trimmed",
		},

		// Edge cases with empty items
		{
			name:     "trailing_comma",
			input:    "data,backup,",
			expected: []string{"data", "backup"},
			desc:     "Trailing comma should be ignored",
		},
		{
			name:     "leading_comma",
			input:    ",data,backup",
			expected: []string{"data", "backup"},
			desc:     "Leading comma should be ignored",
		},
		{
			name:     "multiple_consecutive_commas",
			input:    "data,,backup",
			expected: []string{"data", "backup"},
			desc:     "Multiple consecutive commas should be ignored",
		},
		{
			name:     "comma_with_spaces_only",
			input:    "data, , backup",
			expected: []string{"data", "backup"},
			desc:     "Comma-separated whitespace should be ignored",
		},

		// Realistic patterns
		{
			name:     "directory_names",
			input:    "node_modules,.git,dist,build",
			expected: []string{"node_modules", ".git", "dist", "build"},
			desc:     "Common directory names to exclude",
		},
		{
			name:     "file_patterns",
			input:    "*.log,*.tmp,*.cache",
			expected: []string{"*.log", "*.tmp", "*.cache"},
			desc:     "Wildcard patterns should be preserved",
		},
		{
			name:     "path_patterns",
			input:    "src/test,build/output,temp/cache",
			expected: []string{"src/test", "build/output", "temp/cache"},
			desc:     "Path patterns with slashes should be preserved",
		},
		{
			name:     "windows_path_patterns",
			input:    "C:\\\\temp,D:\\\\backup",
			expected: []string{"C:\\\\temp", "D:\\\\backup"},
			desc:     "Windows-style paths should be preserved",
		},
		{
			name:     "hidden_files_and_folders",
			input:    ".env,.DS_Store,.vscode",
			expected: []string{".env", ".DS_Store", ".vscode"},
			desc:     "Hidden files and folders (starting with dot) should be preserved",
		},

		// Special characters
		{
			name:     "patterns_with_underscores",
			input:    "__pycache__,node_modules,test_data",
			expected: []string{"__pycache__", "node_modules", "test_data"},
			desc:     "Patterns with underscores should be preserved",
		},
		{
			name:     "patterns_with_hyphens",
			input:    "test-data,my-backup,old-files",
			expected: []string{"test-data", "my-backup", "old-files"},
			desc:     "Patterns with hyphens should be preserved",
		},
		{
			name:     "patterns_with_numbers",
			input:    "temp1,backup2,old3",
			expected: []string{"temp1", "backup2", "old3"},
			desc:     "Patterns with numbers should be preserved",
		},
		{
			name:     "patterns_with_mixed_special_chars",
			input:    "test_data-v1,backup.old,temp@2024",
			expected: []string{"test_data-v1", "backup.old", "temp@2024"},
			desc:     "Patterns with various special characters should be preserved",
		},

		// Case sensitivity (note: function preserves case)
		{
			name:     "mixed_case_patterns",
			input:    "Data,BACKUP,TeMp",
			expected: []string{"Data", "BACKUP", "TeMp"},
			desc:     "Case should be preserved exactly as input",
		},

		// Duplicate handling
		{
			name:     "duplicate_patterns",
			input:    "data,backup,data,temp,backup",
			expected: []string{"data", "backup", "data", "temp", "backup"},
			desc:     "Duplicate patterns are preserved (no deduplication)",
		},

		// Excessive whitespace
		{
			name:     "excessive_whitespace_between_items",
			input:    "data    ,    backup    ,    temp",
			expected: []string{"data", "backup", "temp"},
			desc:     "Excessive whitespace should be trimmed from each item",
		},
		{
			name:     "tabs_and_spaces",
			input:    "data\t,\tbackup\t,\ttemp",
			expected: []string{"data", "backup", "temp"},
			desc:     "Tabs should be treated as whitespace and trimmed",
		},

		// Real-world exclusion scenarios
		{
			name:     "common_dev_directories",
			input:    "node_modules,vendor,target,.gradle,build",
			expected: []string{"node_modules", "vendor", "target", ".gradle", "build"},
			desc:     "Common development directories to exclude",
		},
		{
			name:     "temp_and_cache_patterns",
			input:    "tmp,temp,cache,.cache,*.tmp,*.log",
			expected: []string{"tmp", "temp", "cache", ".cache", "*.tmp", "*.log"},
			desc:     "Common temporary and cache patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExcludeToSlice(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("\nTest: %s\nDescription: %s\nInput: %q\nExpected: %v\nGot: %v",
					tt.name, tt.desc, tt.input, tt.expected, result)
			}
		})
	}
}

// Benchmark tests to measure performance

func BenchmarkParseExtToSlice(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"single", "txt"},
		{"multiple_small", "txt,jpg,pdf"},
		{"multiple_large", "go,js,ts,py,java,cpp,rs,php,rb,swift,kt,scala,hs,lua"},
		{"with_spaces", "txt, jpg, pdf, doc, docx, xls, xlsx, ppt, pptx"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ParseExtToSlice(tc.input)
			}
		})
	}
}

func BenchmarkParseExcludeToSlice(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"single", "data"},
		{"multiple_small", "data,backup,temp"},
		{"multiple_large", "node_modules,.git,dist,build,target,vendor,.gradle,.idea,.vscode"},
		{"with_paths", "src/test,build/output,temp/cache,data/backup,logs/old"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ParseExcludeToSlice(tc.input)
			}
		})
	}
}
