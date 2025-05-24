package utils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

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
