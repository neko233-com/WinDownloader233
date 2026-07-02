package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/neko233/unigetui233/mirror"
)

// Downloader handles direct file downloads with progress tracking.
type Downloader struct {
	manager    *Manager
	mirror     *mirror.Manager
	httpClient *http.Client
	downloadDir string
}

// NewDownloader creates a file downloader.
func NewDownloader(mgr *Manager, mirrorMgr *mirror.Manager, downloadDir string) *Downloader {
	return &Downloader{
		manager:     mgr,
		mirror:      mirrorMgr,
		httpClient:  &http.Client{Timeout: 30 * time.Minute},
		downloadDir: downloadDir,
	}
}

// Download downloads a file from the given URL with progress tracking.
// Returns the local file path on success.
func (d *Downloader) Download(toolID, url, filename string) (string, error) {
	resolvedURL := d.mirror.ResolveDownloadURL(url)
	d.manager.UpdateProgress(toolID, "downloading", 0, fmt.Sprintf("Starting download..."))

	req, err := http.NewRequest("GET", resolvedURL, nil)
	if err != nil {
		d.manager.UpdateProgress(toolID, "error", 0, err.Error())
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "UniGetUI233/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.manager.UpdateProgress(toolID, "error", 0, err.Error())
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d.manager.UpdateProgress(toolID, "error", 0, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Ensure download directory exists
	if err := os.MkdirAll(d.downloadDir, 0755); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}

	destPath := filepath.Join(d.downloadDir, filename)
	file, err := os.Create(destPath)
	if err != nil {
		d.manager.UpdateProgress(toolID, "error", 0, err.Error())
		return "", fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				d.manager.UpdateProgress(toolID, "error", 0, writeErr.Error())
				return "", fmt.Errorf("write: %w", writeErr)
			}
			downloaded += int64(n)

			if total > 0 {
				pct := float64(downloaded) / float64(total) * 100
				d.manager.UpdateProgress(toolID, "downloading", pct,
					fmt.Sprintf("%.1f MB / %.1f MB", float64(downloaded)/1e6, float64(total)/1e6))
			} else {
				d.manager.UpdateProgress(toolID, "downloading", -1,
					fmt.Sprintf("%.1f MB downloaded", float64(downloaded)/1e6))
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				d.manager.UpdateProgress(toolID, "error", 0, readErr.Error())
				return "", fmt.Errorf("read: %w", readErr)
			}
			break
		}
	}

	d.manager.UpdateProgress(toolID, "done", 100, fmt.Sprintf("Downloaded to %s", destPath))
	return destPath, nil
}
