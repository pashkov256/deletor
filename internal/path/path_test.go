package path

import "testing"

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"AppDirName is set", AppDirName, "deletor"},
		{"RuleFileName is set", RuleFileName, "rule.json"},
		{"LogFileName is set", LogFileName, "deletor.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant == "" {
				t.Errorf("%s is empty", tt.name)
			}
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
