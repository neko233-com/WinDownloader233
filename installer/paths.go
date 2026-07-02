package installer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultCacheDir returns the app cache root or a child path.
func DefaultCacheDir(parts ...string) string {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	all := append([]string{base, "WinDownloader233"}, parts...)
	return filepath.Join(all...)
}

func safeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "task"
	}
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]+`)
	name = re.ReplaceAllString(name, "_")
	name = strings.Trim(name, ". ")
	if name == "" {
		return "task"
	}
	if len(name) > 120 {
		name = name[:120]
	}
	return name
}
