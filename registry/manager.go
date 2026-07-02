package registry

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed data
var embeddedData embed.FS

// Manager handles loading, caching, and merging of tool registries.
type Manager struct {
	mu            sync.RWMutex
	current       *ToolRegistry
	remoteBaseURL string // GitHub raw URL or CDN proxy
}

// NewManager creates a registry manager with the given remote URL.
func NewManager(remoteBaseURL string) *Manager {
	return &Manager{remoteBaseURL: remoteBaseURL}
}

// LoadEmbedded loads the built-in registry from the embedded JSON file.
func (m *Manager) LoadEmbedded() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := embeddedData.ReadFile("data/tools.json")
	if err != nil {
		return fmt.Errorf("read embedded registry: %w", err)
	}

	var reg ToolRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return fmt.Errorf("parse embedded registry: %w", err)
	}

	m.current = &reg
	return nil
}

// SyncRemote fetches the remote registry and uses it if it's newer.
// Returns true if the remote version was adopted.
func (m *Manager) SyncRemote() (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.remoteBaseURL == "" || m.current == nil {
		return false, nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(m.remoteBaseURL)
	if err != nil {
		return false, fmt.Errorf("fetch remote registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("remote returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("read remote body: %w", err)
	}

	var remote ToolRegistry
	if err := json.Unmarshal(body, &remote); err != nil {
		return false, fmt.Errorf("parse remote registry: %w", err)
	}

	if remote.IsNewerThan(m.current) {
		m.current = &remote
		return true, nil
	}

	return false, nil
}

// SetRemoteURL updates the remote URL (e.g. switching to a CDN mirror).
func (m *Manager) SetRemoteURL(url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.remoteBaseURL = url
}

// GetAll returns all tools in the current registry.
func (m *Manager) GetAll() []Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil {
		return nil
	}
	out := make([]Tool, len(m.current.Tools))
	copy(out, m.current.Tools)
	return out
}

// GetByCategory returns tools filtered by category, sorted by name.
func (m *Manager) GetByCategory(category string) []Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil {
		return nil
	}

	var tools []Tool
	for _, t := range m.current.Tools {
		if t.Category == category {
			tools = append(tools, t)
		}
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name.EN < tools[j].Name.EN })
	return tools
}

// GetByTags returns tools that contain ALL of the specified tags.
func (m *Manager) GetByTags(tags []string) []Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil || len(tags) == 0 {
		return nil
	}

	tagSet := make(map[string]bool, len(tags))
	for _, t := range tags {
		tagSet[strings.ToLower(t)] = true
	}

	var tools []Tool
	for _, t := range m.current.Tools {
		if toolHasAllTags(t, tagSet) {
			tools = append(tools, t)
		}
	}
	return tools
}

func toolHasAllTags(tool Tool, required map[string]bool) bool {
	have := make(map[string]bool, len(tool.Tags))
	for _, t := range tool.Tags {
		have[strings.ToLower(t)] = true
	}
	for tag := range required {
		if !have[tag] {
			return false
		}
	}
	return true
}

// Search performs a case-insensitive search across name, description, tags.
func (m *Manager) Search(query string) []Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil || query == "" {
		return m.getAllUnsafe()
	}

	q := strings.ToLower(query)
	var tools []Tool
	for _, t := range m.current.Tools {
		if matchesQuery(t, q) {
			tools = append(tools, t)
		}
	}
	return tools
}

func (m *Manager) getAllUnsafe() []Tool {
	if m.current == nil {
		return nil
	}
	out := make([]Tool, len(m.current.Tools))
	copy(out, m.current.Tools)
	return out
}

func matchesQuery(t Tool, q string) bool {
	if strings.Contains(strings.ToLower(t.Name.ZH), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Name.EN), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Description.ZH), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Description.EN), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.ID), q) {
		return true
	}
	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

// GetAllTags returns all unique tags across all tools.
func (m *Manager) GetAllTags() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil {
		return nil
	}

	tagMap := make(map[string]bool)
	for _, t := range m.current.Tools {
		for _, tag := range t.Tags {
			tagMap[tag] = true
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// GetByID returns a tool by its ID.
func (m *Manager) GetByID(id string) *Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil {
		return nil
	}
	for i := range m.current.Tools {
		if m.current.Tools[i].ID == id {
			t := m.current.Tools[i]
			return &t
		}
	}
	return nil
}

// GetTimestamp returns the current registry timestamp.
func (m *Manager) GetTimestamp() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil {
		return ""
	}
	return m.current.Timestamp
}
