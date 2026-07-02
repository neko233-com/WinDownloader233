package mirror

import (
	"strings"
	"sync"
)

// MirrorConfig holds the current mirror configuration.
type MirrorConfig struct {
	GitHubProxy  string `json:"githubProxy"`  // e.g. https://ghfast.top/
	DownloadCDN  string `json:"downloadCdn"`  // e.g. https://mirror.ghproxy.com/
	UseProxy     bool   `json:"useProxy"`
	HTTPProxy    string `json:"httpProxy"`    // e.g. http://127.0.0.1:7890
}

// Manager handles mirror/CDN configuration for China users.
type Manager struct {
	mu     sync.RWMutex
	config MirrorConfig
}

// DefaultConfig returns sensible defaults for China users.
func DefaultConfig() MirrorConfig {
	return MirrorConfig{
		GitHubProxy: "https://ghfast.top/",
		DownloadCDN: "https://ghfast.top/",
		UseProxy:    false,
		HTTPProxy:   "",
	}
}

// NewManager creates a mirror manager with the given config.
func NewManager(cfg MirrorConfig) *Manager {
	return &Manager{config: cfg}
}

// GetConfig returns the current mirror configuration.
func (m *Manager) GetConfig() MirrorConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// SetConfig replaces the mirror configuration.
func (m *Manager) SetConfig(cfg MirrorConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
}

// ResolveGitHubURL converts a raw GitHub URL to use the configured proxy.
// Returns the original URL if proxy is not configured or URL is not a GitHub URL.
func (m *Manager) ResolveGitHubURL(rawURL string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.config.UseProxy || m.config.GitHubProxy == "" {
		return rawURL
	}
	if !isGitHubURL(rawURL) {
		return rawURL
	}
	return strings.TrimRight(m.config.GitHubProxy, "/") + "/" + rawURL
}

// ResolveDownloadURL applies CDN mirror to a download URL if enabled.
func (m *Manager) ResolveDownloadURL(rawURL string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.config.UseProxy || m.config.DownloadCDN == "" {
		return rawURL
	}
	if !isGitHubURL(rawURL) {
		return rawURL
	}
	return strings.TrimRight(m.config.DownloadCDN, "/") + "/" + rawURL
}

func isGitHubURL(url string) bool {
	return strings.Contains(url, "github.com") ||
		strings.Contains(url, "raw.githubusercontent.com") ||
		strings.Contains(url, "githubusercontent.com")
}

// AvailableProxies returns a list of known GitHub proxy services for China.
func AvailableProxies() []ProxyOption {
	return []ProxyOption{
		{Name: "ghfast.top", URL: "https://ghfast.top/", Description: "高速稳定"},
		{Name: "gh-proxy.com", URL: "https://gh-proxy.com/", Description: "免费公共代理"},
		{Name: "mirror.ghproxy.com", URL: "https://mirror.ghproxy.com/", Description: "GitHub 镜像"},
		{Name: "自定义", URL: "", Description: "输入自定义代理地址"},
	}
}

// ProxyOption represents a selectable GitHub proxy.
type ProxyOption struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}
