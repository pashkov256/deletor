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
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single extension without dot",
			input:    "txt",
			expected: []string{".txt"},
		},
		{
			name:     "single extension with dot",
			input:    ".log",
			expected: []string{".log"},
		},
		{
			name:     "multiple extensions with spaces",
			input:    "txt, log, json",
			expected: []string{".txt", ".log", ".json"},
		},
		{
			name:     "multiple extensions without spaces",
			input:    "txt,json,log",
			expected: []string{".txt", ".json", ".log"},
		},
		{
			name:     "uppercase extensions",
			input:    "TXT,LOG,JSON",
			expected: []string{".txt", ".log", ".json"},
		},
		{
			name:     "mixed case extensions",
			input:    "Txt,Log,JsoN",
			expected: []string{".txt", ".log", ".json"},
		},
		{
			name:     "extensions with extra spaces",
			input:    "  txt  ,  log  ,  json  ",
			expected: []string{".txt", ".log", ".json"},
		},
		{
			name:     "empty entries",
			input:    "txt,,log,",
			expected: []string{".txt", ".log"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExtToSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseExtToSlice(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseExcludeToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single pattern",
			input:    "*.tmp",
			expected: []string{"*.tmp"},
		},
		{
			name:     "multiple patterns with spaces",
			input:    "*.tmp, *.log, *.bak",
			expected: []string{"*.tmp", "*.log", "*.bak"},
		},
		{
			name:     "multiple patterns without spaces",
			input:    "*.tmp,*.log,*.bak",
			expected: []string{"*.tmp", "*.log", "*.bak"},
		},
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