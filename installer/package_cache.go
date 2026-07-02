package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type packageRowsCacheEntry struct {
	UpdatedAt time.Time     `json:"updatedAt"`
	Rows      []PackageInfo `json:"rows"`
}

var packageCache = struct {
	mu          sync.Mutex
	managersAt  time.Time
	managers    []PackageManagerInfo
	rows        map[string]packageRowsCacheEntry
	refreshing  map[string]bool
	loadedFiles map[string]bool
}{
	rows:        make(map[string]packageRowsCacheEntry),
	refreshing:  make(map[string]bool),
	loadedFiles: make(map[string]bool),
}

const (
	managerCacheTTL = 2 * time.Minute
	rowCacheTTL     = 10 * time.Minute
)

func cachedPackageManagers(force bool) []PackageManagerInfo {
	now := time.Now()
	packageCache.mu.Lock()
	if !force && len(packageCache.managers) > 0 && now.Sub(packageCache.managersAt) < managerCacheTTL {
		out := cloneManagers(packageCache.managers)
		packageCache.mu.Unlock()
		return out
	}
	packageCache.mu.Unlock()

	rows := listPackageManagersLive()

	packageCache.mu.Lock()
	packageCache.managers = cloneManagers(rows)
	packageCache.managersAt = now
	packageCache.mu.Unlock()
	return rows
}

func cachedPackageRows(kind, manager string, force bool) []PackageInfo {
	key := packageCacheKey(kind, manager)
	if !force {
		if rows, ok := cachedRows(key); ok {
			startPackageRefresh(kind, manager, key)
			return rows
		}
		if rows, ok := loadPackageRowsFromDisk(key); ok {
			startPackageRefresh(kind, manager, key)
			return rows
		}
	}
	return refreshPackageRows(kind, manager, key)
}

func InvalidatePackageCache(manager string) {
	packageCache.mu.Lock()
	defer packageCache.mu.Unlock()
	for _, kind := range []string{"installed", "updates"} {
		delete(packageCache.rows, packageCacheKey(kind, manager))
		delete(packageCache.loadedFiles, packageCacheKey(kind, manager))
		delete(packageCache.rows, packageCacheKey(kind, "all"))
		delete(packageCache.loadedFiles, packageCacheKey(kind, "all"))
	}
}

func CachedInstalledPackageID(manager, id string) bool {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return false
	}
	key := packageCacheKey("installed", manager)
	rows, ok := cachedRows(key)
	if !ok {
		rows, ok = loadPackageRowsFromDisk(key)
	}
	if !ok {
		return false
	}
	for _, row := range rows {
		if strings.ToLower(row.ID) == id || strings.ToLower(row.Name) == id {
			return true
		}
	}
	return false
}

func cachedRows(key string) ([]PackageInfo, bool) {
	packageCache.mu.Lock()
	defer packageCache.mu.Unlock()
	entry, ok := packageCache.rows[key]
	if !ok {
		return nil, false
	}
	return clonePackageRows(entry.Rows), true
}

func loadPackageRowsFromDisk(key string) ([]PackageInfo, bool) {
	packageCache.mu.Lock()
	if packageCache.loadedFiles[key] {
		packageCache.mu.Unlock()
		return nil, false
	}
	packageCache.loadedFiles[key] = true
	packageCache.mu.Unlock()

	data, err := os.ReadFile(packageCachePath(key))
	if err != nil {
		return nil, false
	}
	var entry packageRowsCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil || entry.UpdatedAt.IsZero() {
		return nil, false
	}
	if time.Since(entry.UpdatedAt) > 24*time.Hour {
		return nil, false
	}

	packageCache.mu.Lock()
	packageCache.rows[key] = entry
	packageCache.mu.Unlock()
	return clonePackageRows(entry.Rows), true
}

func startPackageRefresh(kind, manager, key string) {
	packageCache.mu.Lock()
	entry := packageCache.rows[key]
	if packageCache.refreshing[key] || time.Since(entry.UpdatedAt) < rowCacheTTL {
		packageCache.mu.Unlock()
		return
	}
	packageCache.refreshing[key] = true
	packageCache.mu.Unlock()

	go func() {
		_ = refreshPackageRows(kind, manager, key)
		packageCache.mu.Lock()
		packageCache.refreshing[key] = false
		packageCache.mu.Unlock()
	}()
}

func refreshPackageRows(kind, manager, key string) []PackageInfo {
	var rows []PackageInfo
	switch kind {
	case "installed":
		rows = listInstalledPackagesLive(manager)
	case "updates":
		rows = listPackageUpdatesLive(manager)
	default:
		rows = nil
	}
	entry := packageRowsCacheEntry{UpdatedAt: time.Now(), Rows: clonePackageRows(rows)}

	packageCache.mu.Lock()
	packageCache.rows[key] = entry
	packageCache.loadedFiles[key] = true
	packageCache.mu.Unlock()

	_ = os.MkdirAll(packageCacheDir(), 0755)
	if data, err := json.Marshal(entry); err == nil {
		_ = os.WriteFile(packageCachePath(key), data, 0644)
	}
	return rows
}

func packageCacheKey(kind, manager string) string {
	if manager == "" {
		manager = "all"
	}
	return kind + "_" + manager
}

func packageCacheDir() string {
	return DefaultCacheDir("package-cache")
}

func packageCachePath(key string) string {
	return filepath.Join(packageCacheDir(), safeFilename(key)+".json")
}

func clonePackageRows(rows []PackageInfo) []PackageInfo {
	out := make([]PackageInfo, len(rows))
	copy(out, rows)
	return out
}

func cloneManagers(rows []PackageManagerInfo) []PackageManagerInfo {
	out := make([]PackageManagerInfo, len(rows))
	copy(out, rows)
	return out
}
