package config_test

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/cli/config"
	"github.com/stretchr/testify/assert"
)

// resetFlags properly resets flag state between tests
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

// TestDefaultValues verifies default config when no flags are provided
func TestDefaultValues(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	cfg := config.GetFlags()

	assert.Equal(t, ".", cfg.Directory) // Default should be current directory
	assert.Nil(t, cfg.Extensions)
	assert.Zero(t, cfg.MinSize)
	assert.Zero(t, cfg.MaxSize)
	assert.Nil(t, cfg.Exclude)
	assert.False(t, cfg.IncludeSubdirs)
	assert.False(t, cfg.ShowProgress)
	assert.False(t, cfg.IsCLIMode)
	assert.False(t, cfg.HaveProgress)
	assert.False(t, cfg.SkipConfirm)
	assert.False(t, cfg.DeleteEmptyFolders)
	assert.True(t, cfg.OlderThan.IsZero())
	assert.True(t, cfg.NewerThan.IsZero())
	assert.False(t, cfg.MoveFileToTrash)
	assert.False(t, cfg.UseRules)
}

// TestDirectoryFlag verifies -d flag parsing
func TestDirectoryFlag(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-d", "/test/path"}

	cfg := config.GetFlags()
	assert.Equal(t, "/test/path", cfg.Directory)
}

// TestExtensionsFlag verifies -e flag parsing
func TestExtensionsFlag(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-e", "txt,log"}

	cfg := config.GetFlags()
	assert.Equal(t, []string{".txt", ".log"}, cfg.Extensions)
}

// TestExcludeFlag verifies --exclude flag parsing
func TestExcludeFlag(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--exclude", "temp,backup"}

	cfg := config.GetFlags()
	assert.Equal(t, []string{"temp", "backup"}, cfg.Exclude)
}

// TestMinSizeFlag verifies --min-size flag parsing
func TestMinSizeFlag(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--min-size", "10MB"}

	cfg := config.GetFlags()
	assert.Equal(t, int64(10*1024*1024), cfg.MinSize)
}

// TestMaxSizeFlag verifies --max-size flag parsing
func TestMaxSizeFlag(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--max-size", "1GB"}

	cfg := config.GetFlags()
	assert.Equal(t, int64(1024*1024*1024), cfg.MaxSize)
}

// TestOlderFlag verifies --older flag parsing
func TestOlderFlag(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd", "--older", "1day"}

	cfg := config.GetFlags()

	expected := time.Now().Add(-24 * time.Hour)
	assert.WithinDuration(t, expected, cfg.OlderThan, 5*time.Second)
}

func TestNewerFlag(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--newer", "1hour"} // Use hours format

	cfg := config.GetFlags()

	// Calculate expected time
	expected := time.Now().Add(-1 * time.Hour)

	// Use more lenient time comparison
	assert.WithinDuration(t, expected, cfg.NewerThan, 5*time.Second)
}

// TestBooleanFlags verifies boolean flag parsing
func TestBooleanFlags(t *testing.T) {
	testCases := []struct {
		flag  string
		check func(*config.Config) bool
	}{
		{"--cli", func(c *config.Config) bool { return c.IsCLIMode }},
		{"--progress", func(c *config.Config) bool { return c.HaveProgress }}, // FIXED HERE
		{"--subdirs", func(c *config.Config) bool { return c.IncludeSubdirs }},
		{"--skip-confirm", func(c *config.Config) bool { return c.SkipConfirm }},
		{"--prune-empty", func(c *config.Config) bool { return c.DeleteEmptyFolders }},
	}

	for _, tc := range testCases {
		t.Run(tc.flag, func(t *testing.T) {
			resetFlags()
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = []string{"cmd", tc.flag}

			cfg := config.GetFlags()
			assert.True(t, tc.check(cfg), "Flag %s should be true", tc.flag)
		})
	}
}

// TestAllFlagsTogether verifies all flags work together
func TestAllFlagsTogether(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"cmd",
		"-d", "/full/path",
		"-e", "go,mod",
		"--exclude", "vendor,node_modules",
		"--min-size", "1MB",
		"--max-size", "10MB",
		"--older", "7day", // ✅ fixed from "7d"
		"--newer", "1hour", // also safer format
		"--cli",
		"--progress",
		"--subdirs",
		"--skip-confirm",
		"--prune-empty",
	}

	cfg := config.GetFlags()

	expectedOlder := time.Now().Add(-7 * 24 * time.Hour)
	expectedNewer := time.Now().Add(-1 * time.Hour)

	assert.Equal(t, "/full/path", cfg.Directory)
	assert.Equal(t, []string{".go", ".mod"}, cfg.Extensions)
	assert.Equal(t, []string{"vendor", "node_modules"}, cfg.Exclude)
	assert.Equal(t, int64(1024*1024), cfg.MinSize)
	assert.Equal(t, int64(10*1024*1024), cfg.MaxSize)

	assert.WithinDuration(t, expectedOlder, cfg.OlderThan, 5*time.Second)
	assert.WithinDuration(t, expectedNewer, cfg.NewerThan, 5*time.Second)

	assert.True(t, cfg.IsCLIMode)
	assert.True(t, cfg.HaveProgress)
	assert.True(t, cfg.IncludeSubdirs)
	assert.True(t, cfg.SkipConfirm)
	assert.True(t, cfg.DeleteEmptyFolders)
}
