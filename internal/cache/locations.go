package cache

import (
	"os"
	"path/filepath"
)

// getLocationsForOS returns a list of cache locations specific to the given operating system.
// Returns nil for unsupported operating systems.
func getLocationsForOS(osName OS) []CacheLocation {
	switch osName {
	case Windows:
		return []CacheLocation{
			{
				Path: filepath.Join(os.Getenv("LOCALAPPDATA"), "Temp"),
			},

			{
				Path: filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "Explorer"),
			},
		}
	case Linux:
		home, _ := os.UserHomeDir()
		return []CacheLocation{
			{
				Path: "/tmp",
			},
			{
				Path: "/var/tmp",
			},
			{
				Path: filepath.Join(home, ".cache"),
			},
		}
	default:
		return nil
	}
}
