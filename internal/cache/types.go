package cache

type CacheType string
type OS string

const (
	Windows OS = "windows"
	Linux   OS = "linux"
)
const (
	SystemCache CacheType = "system"
	AppCache    CacheType = "app"
)

type CacheLocation struct {
	Path string
	Type string
}

type ScanResult struct {
	FileCount int64
	Path      string
	Size      int64
	Error     error
}
