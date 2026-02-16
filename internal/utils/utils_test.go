package utils

import (
	"reflect"
	"testing"
)

func TestParseExtToSlice(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "single extension without dot",
			input: "txt",
			want:  []string{".txt"},
		},
		{
			name:  "single extension with dot",
			input: ".txt",
			want:  []string{".txt"},
		},
		{
			name:  "multiple extensions without dots",
			input: "txt,jpg,png",
			want:  []string{".txt", ".jpg", ".png"},
		},
		{
			name:  "multiple extensions with dots",
			input: ".txt,.jpg,.png",
			want:  []string{".txt", ".jpg", ".png"},
		},
		{
			name:  "mixed extensions with and without dots",
			input: "txt,.jpg,png",
			want:  []string{".txt", ".jpg", ".png"},
		},
		{
			name:  "extensions with spaces",
			input: "  txt  ,  jpg  ,  png  ",
			want:  []string{".txt", ".jpg", ".png"},
		},
		{
			name:  "uppercase extensions",
			input: "TXT,JPG,PNG",
			want:  []string{".txt", ".jpg", ".png"},
		},
		{
			name:  "mixed case with spaces and dots",
			input: " .TXT , Jpg , .pNg ",
			want:  []string{".txt", ".jpg", ".png"},
		},
		{
			name:  "empty values in list",
			input: "txt,,jpg,,,png",
			want:  []string{".txt", ".jpg", ".png"},
		},
		{
			name:  "only commas",
			input: ",,,",
			want:  []string{},
		},
		{
			name:  "only spaces",
			input: "   ",
			want:  []string{},
		},
		{
			name:  "single space delimited values",
			input: "txt, jpg, png",
			want:  []string{".txt", ".jpg", ".png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseExtToSlice(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExtToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseExcludeToSlice(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "single pattern",
			input: "node_modules",
			want:  []string{"node_modules"},
		},
		{
			name:  "multiple patterns",
			input: "node_modules,.git,dist",
			want:  []string{"node_modules", ".git", "dist"},
		},
		{
			name:  "patterns with spaces",
			input: "  node_modules  ,  .git  ,  dist  ",
			want:  []string{"node_modules", ".git", "dist"},
		},
		{
			name:  "patterns with wildcards",
			input: "*.log,*.tmp,test_*",
			want:  []string{"*.log", "*.tmp", "test_*"},
		},
		{
			name:  "empty values in list",
			input: "node_modules,,.git,,,dist",
			want:  []string{"node_modules", ".git", "dist"},
		},
		{
			name:  "only commas",
			input: ",,,",
			want:  []string{},
		},
		{
			name:  "only spaces",
			input: "   ",
			want:  []string{},
		},
		{
			name:  "patterns with paths",
			input: "tmp/cache,logs/debug,build/output",
			want:  []string{"tmp/cache", "logs/debug", "build/output"},
		},
		{
			name:  "single character patterns",
			input: "a,b,c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "mixed spacing",
			input: "node_modules, .git,  dist   , tmp",
			want:  []string{"node_modules", ".git", "dist", "tmp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseExcludeToSlice(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExcludeToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
