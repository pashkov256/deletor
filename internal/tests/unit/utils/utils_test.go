package utils

import (
	"bytes"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/cli/output"
	"github.com/pashkov256/deletor/internal/utils"
)

var printer = output.NewPrinter()

func TestPrintFilesTable(t *testing.T) {
	type args struct {
		files map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "BytePrint",
			args: args{map[string]string{
				"/Users/test/Documents/deletor/main.go": "8.04 KB",
			}},
			want: "8.04 KB  /Users/test/Documents/deletor/main.go\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printer.PrintFilesTable(tt.args.files)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			got := buf.String()

			// Remove color codes before comparison
			got = strings.ReplaceAll(got, "\x1b[33m", "")
			got = strings.ReplaceAll(got, "\x1b[0m", "")
			got = strings.ReplaceAll(got, "\x1b[37m", "")

			if got != tt.want {
				t.Errorf("\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestAskForConfirmation(t *testing.T) {
	type args struct {
		userInput string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"yLowerCase", args{"y\n"}, true},
		{"yUpperCase", args{"Y\n"}, true},
		{"yesLowerCase", args{"YES\n"}, true},
		{"yesUpperCase", args{"yes\n"}, true},
		{"nLowerCase", args{"n\n"}, false},
		{"nUpperCase", args{"N\n"}, false},
		{"noLowerCase", args{"no\n"}, false},
		{"noUpperCase", args{"NO\n"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStdin := os.Stdin
			defer func() { os.Stdin = originalStdin }()

			r, w, _ := os.Pipe()
			os.Stdin = r
			go func() {
				w.Write([]byte(tt.args.userInput))
				w.Close()
			}()
			got := printer.AskForConfirmation("Delete these files?")
			if got != tt.want {
				t.Errorf("gotAskForConfirmation = %v\n wantAskForConfirmation = %v", got, tt.want)
			}
		})
	}
}

func TestToBytes(t *testing.T) {
	type args struct {
		sizeStr string
	}

	tests := []struct {
		name       string
		args       args
		wantToByte int64
	}{
		{"minB", args{"0b"}, 0},
		{"100B", args{"100b"}, 100},
		{"minKB", args{"0k"}, 0},
		{"100KB", args{"100kb"}, 102400},
		{"minMB", args{"0mb"}, 0},
		{"100MB", args{"100mb"}, 104857600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := utils.ToBytes(tt.args.sizeStr); got != tt.wantToByte {
				t.Errorf("gotToBytes = %v\n wantToBytes = %v", got, tt.wantToByte)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	type args struct {
		bytes int64
	}

	tests := []struct {
		name           string
		args           args
		wantFormatSize string
	}{
		{"MinB", args{0}, "0 B"},
		{"MaxB", args{1023}, "1023 B"},
		{"MinKB", args{1024}, "1.00 KB"},
		{"MaxKB", args{1048575}, "1024.00 KB"},
		{"MinMB", args{1048576}, "1.00 MB"},
		{"MaxMB", args{1073741823}, "1024.00 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFormatSize := utils.FormatSize(tt.args.bytes); gotFormatSize != tt.wantFormatSize {
				t.Errorf("gotFormatSize = %v\n wantFormatSize = %v", gotFormatSize, tt.wantFormatSize)
			}
		})
	}
}

func TestParseTimeDuration(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDuration time.Duration
		wantErr      bool
		wantZero     bool // expect zero time.Time (invalid input with no leading digits)
	}{
		// Standard units
		{"1sec", "1sec", 1 * time.Second, false, false},
		{"2min", "2min", 2 * time.Minute, false, false},
		{"3hour", "3hour", 3 * time.Hour, false, false},
		{"4day", "4day", 4 * 24 * time.Hour, false, false},
		{"5week", "5week", 5 * 7 * 24 * time.Hour, false, false},
		{"6month", "6month", 6 * 30 * 24 * time.Hour, false, false},
		{"7year", "7year", 7 * 365 * 24 * time.Hour, false, false},

		// Plural / longer unit names
		{"10seconds", "10seconds", 10 * time.Second, false, false},
		{"30minutes", "30minutes", 30 * time.Minute, false, false},
		{"24hours", "24hours", 24 * time.Hour, false, false},
		{"2days", "2days", 2 * 24 * time.Hour, false, false},
		{"2weeks", "2weeks", 2 * 7 * 24 * time.Hour, false, false},
		{"3months", "3months", 3 * 30 * 24 * time.Hour, false, false},
		{"1years", "1years", 365 * 24 * time.Hour, false, false},

		// Input with space between number and unit
		{"space 24 hours", "24 hours", 24 * time.Hour, false, false},
		{"space 7 day", "7 day", 7 * 24 * time.Hour, false, false},

		// Leading/trailing whitespace
		{"whitespace", "  12hour  ", 12 * time.Hour, false, false},

		// Upper case input (function lowercases)
		{"uppercase", "5DAY", 5 * 24 * time.Hour, false, false},
		{"short day", "10d", 10 * 24 * time.Hour, false, false},
		{"short hour", "3h", 3 * time.Hour, false, false},
		{"short month", "2mo", 2 * 30 * 24 * time.Hour, false, false},

		// Edge: empty string → unitIndex==0 → returns zero time
		{"empty string", "", 0, false, true},

		// Edge: no number, just unit → unitIndex==0 → returns zero time
		{"no number", "day", 0, false, true},

		// Unknown unit
		{"unknown unit", "5xyz", 0, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			got, err := utils.ParseTimeDuration(tt.input)
			after := time.Now()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
				return
			}

			if tt.wantZero {
				if !got.IsZero() {
					t.Errorf("expected zero time for input %q, got %v", tt.input, got)
				}
				return
			}

			// The function returns time.Now().Add(-duration).
			// Verify the result is within a reasonable tolerance (2 seconds).
			expectedEarliest := before.Add(-tt.wantDuration)
			expectedLatest := after.Add(-tt.wantDuration)

			diffEarliest := math.Abs(float64(got.Sub(expectedEarliest)))
			diffLatest := math.Abs(float64(got.Sub(expectedLatest)))
			minDiff := math.Min(diffEarliest, diffLatest)

			if minDiff > float64(2*time.Second) {
				t.Errorf("ParseTimeDuration(%q) = %v, expected ~%v ago from now",
					tt.input, got, tt.wantDuration)
			}
		})
	}
}
