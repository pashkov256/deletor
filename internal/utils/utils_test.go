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
			name:     "single extension with dot",
			input:    ".go",
			expected: []string{".go"},
		},
		{
			name:     "single extension without dot",
			input:    "go",
			expected: []string{".go"},
		},
		{
			name:     "multiple extensions with dots",
			input:    ".go,.js,.py",
			expected: []string{".go", ".js", ".py"},
		},
		{
			name:     "multiple extensions without dots",
			input:    "go,js,py",
			expected: []string{".go", ".js", ".py"},
		},
		{
			name:     "mixed extensions with and without dots",
			input:    ".go,js,.py,ts",
			expected: []string{".go", ".js", ".py", ".ts"},
		},
		{
			name:     "extensions with spaces",
			input:    " .go , js , .py ",
			expected: []string{".go", ".js", ".py"},
		},
		{
			name:     "uppercase extensions should be lowercased",
			input:    "GO,JS,PY",
			expected: []string{".go", ".js", ".py"},
		},
		{
			name:     "mixed case extensions",
			input:    ".Go,.JS,pY",
			expected: []string{".go", ".js", ".py"},
		},
		{
			name:     "empty items in list",
			input:    "go,,js,,py",
			expected: []string{".go", ".js", ".py"},
		},
		{
			name:     "single space",
			input:    " ",
			expected: []string{},
		},
		{
			name:     "comma only",
			input:    ",",
			expected: []string{},
		},
		{
			name:     "multiple commas",
			input:    ",,,",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExtToSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseExtToSlice(%q) = %v, expected %v", tt.input, result, tt.expected)
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
			input:    "node_modules",
			expected: []string{"node_modules"},
		},
		{
			name:     "multiple patterns",
			input:    "node_modules,.git,dist",
			expected: []string{"node_modules", ".git", "dist"},
		},
		{
			name:     "patterns with spaces",
			input:    " node_modules , .git , dist ",
			expected: []string{"node_modules", ".git", "dist"},
		},
		{
			name:     "patterns with wildcard",
			input:    "*.log,temp*,*cache",
			expected: []string{"*.log", "temp*", "*cache"},
		},
		{
			name:     "empty items in list",
			input:    "node_modules,,dist,,build",
			expected: []string{"node_modules", "dist", "build"},
		},
		{
			name:     "single space",
			input:    " ",
			expected: []string{},
		},
		{
			name:     "comma only",
			input:    ",",
			expected: []string{},
		},
		{
			name:     "multiple commas",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "paths with slashes",
			input:    "vendor/,build/output,tmp/cache",
			expected: []string{"vendor/", "build/output", "tmp/cache"},
		},
		{
			name:     "mixed spacing",
			input:    "  node_modules  ,  .git  ,  dist  ",
			expected: []string{"node_modules", ".git", "dist"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExcludeToSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseExcludeToSlice(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
