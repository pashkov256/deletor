package options

import "testing"

func TestGetNextOption(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		prefix   string
		max      int
		forward  bool
		expected string
	}{
		// 1. Basic Forward Navigation
		{
			name:     "Forward from 1 to 2",
			current:  "Option1",
			prefix:   "Option",
			max:      3,
			forward:  true,
			expected: "Option2",
		},
		{
			name:     "Forward from 2 to 3",
			current:  "Test2",
			prefix:   "Test",
			max:      3,
			forward:  true,
			expected: "Test3",
		},

		// 2. Forward Wrapping (End -> Start)
		{
			name:     "Forward wrap around (Max to 1)",
			current:  "Page3",
			prefix:   "Page",
			max:      3,
			forward:  true,
			expected: "Page1",
		},

		// 3. Basic Backward Navigation
		{
			name:     "Backward from 2 to 1",
			current:  "Option2",
			prefix:   "Option",
			max:      3,
			forward:  false, // Backward
			expected: "Option1",
		},

		// 4. Backward Wrapping (Start -> End)
		{
			name:     "Backward wrap around (1 to Max)",
			current:  "Option1",
			prefix:   "Option",
			max:      5,
			forward:  false, // Backward
			expected: "Option5",
		},

		// 5. Edge Case: Invalid or Empty Current Option
		// The code defaults 'currentNum' to 1 if parsing fails.
		// So if we go forward, 1 + 1 = 2.
		{
			name:     "Invalid current string (defaults to 1, goes forward to 2)",
			current:  "InvalidString",
			prefix:   "Option",
			max:      5,
			forward:  true,
			expected: "Option2",
		},
		// If parsing fails (defaults to 1) and we go backward:
		// 1 - 1 = 0 -> wraps to Max.
		{
			name:     "Invalid current string (defaults to 1, goes backward to Max)",
			current:  "InvalidString",
			prefix:   "Option",
			max:      5,
			forward:  false,
			expected: "Option5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNextOption(tt.current, tt.prefix, tt.max, tt.forward)
			if got != tt.expected {
				t.Errorf("GetNextOption(%q, %q, %d, %v) = %q; want %q",
					tt.current, tt.prefix, tt.max, tt.forward, got, tt.expected)
			}
		})
	}
}
