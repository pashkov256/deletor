package cache

// CacheType represents the type of cache (system or application)
type CacheType string

// OS represents supported operating systems
type OS string

const (
	Windows OS = "windows"
	Linux   OS = "linux"
)

const (
	SystemCache CacheType = "system" // System-wide cache
	AppCache    CacheType = "app"    // Application-specific cache
)

// CacheLocation represents a cache directory location with its path and type
type CacheLocation struct {
	Path string
	Type string
}

// ScanResult contains information about a cache scan operation
type ScanResult struct {
	FileCount int64  // Number of files found
	Path      string // Path that was scanned
	Size      int64  // Total size of cache in bytes
	Error     error  // Any error that occurred during scanning
}
