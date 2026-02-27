package path

import (
	"testing"
)

func TestPathDefaults(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"AppDirName", AppDirName, "deletor"},
		{"RuleFileName", RuleFileName, "rule.json"},
		{"LogFileName", LogFileName, "deletor.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q; want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}
