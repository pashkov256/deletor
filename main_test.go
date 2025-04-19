package main

import (
	"bytes"
	"io"
	"os"
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
		{"BytePrint", args{map[string]string{
			"/Users/test/Documents/deletor/main.go": "8.04 KB",
		}}, "8.04 KB  /Users/test/Documents/deletor/main.go\n"},
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printFilesTable(tt.args.files)
			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			got := buf.String()
			if got != tt.want {
				t.Errorf("gotFormatSize = %v\n wantFormatSize = %v", got, tt.want)
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
			if got, _ := toBytes(tt.args.sizeStr); got != tt.wantToByte {
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
			if gotFormatSize := formatSize(tt.args.bytes); gotFormatSize != tt.wantFormatSize {
				t.Errorf("gotFormatSize = %v\n wantFormatSize = %v", gotFormatSize, tt.wantFormatSize)
			}
		})
	}
}
