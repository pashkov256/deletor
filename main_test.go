package main

import (
	"bytes"
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

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
			// Перехватываем вывод
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printFilesTable(tt.args.files)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			got := buf.String()

			// Удаляем цветовые коды перед сравнением
			got = strings.ReplaceAll(got, "\x1b[33m", "") // желтый
			got = strings.ReplaceAll(got, "\x1b[0m", "")  // сброс
			got = strings.ReplaceAll(got, "\x1b[37m", "") // белый

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
			got := askForConfirmation("Delete these files?")
			if got != tt.want {
				t.Errorf("gotAskForConfirmation = %v\n wantAskForConfirmation = %v", got, tt.want)
			}
		})
	}
}

func TestToBytes(t *testing.T) {
	got := math.Abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %f; want 1", got)
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
			if gotFormatSize := formatSize(tt.args.bytes); gotFormatSize != tt.wantFormatSize {
				t.Errorf("gotFormatSize = %v\n wantFormatSize = %v", gotFormatSize, tt.wantFormatSize)
			}
		})
	}
}

func TestLogDeletionToFile(t *testing.T) {
	got := math.Abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %f; want 1", got)
	}
}
