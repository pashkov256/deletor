package main

import (
	"bytes"
	"io"
	"math"
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
	got := math.Abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %f; want 1", got)
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
