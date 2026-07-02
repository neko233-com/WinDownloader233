package registry

import "time"

// Category constants for tool classification.
const (
	CategoryProgramming = "programming"
	CategoryArt         = "art"
	CategoryPlanning    = "planning"
	CategoryAudio       = "audio"
	CategoryQA          = "qa"
	CategoryPM          = "pm"
	CategoryAI          = "ai"
)

// AllCategories returns all category IDs in display order.
func AllCategories() []string {
	return []string{
		CategoryProgramming, CategoryArt, CategoryPlanning,
		CategoryAudio, CategoryQA, CategoryPM, CategoryAI,
	}
}

// Tool represents a single installable tool in the registry.
type Tool struct {
	ID          string           `json:"id"`
	Name        LocalizedString  `json:"name"`
	Description LocalizedString  `json:"description"`
	Category    string           `json:"category"`
	Tags        []string         `json:"tags"`
	Icon        string           `json:"icon"`
	Version     string           `json:"version"`
	Size        string           `json:"size"`
	WingetID    string           `json:"wingetId,omitempty"`
	DownloadURL string           `json:"downloadUrl,omitempty"`
	MirrorURL   string           `json:"mirrorUrl,omitempty"`
	Homepage    string           `json:"homepage,omitempty"`
	IsFree      bool             `json:"isFree"`
	InstallType string           `json:"installType"` // "winget" | "download" | "manual"
}

// LocalizedString holds bilingual text.
type LocalizedString struct {
	ZH string `json:"zh"`
	EN string `json:"en"`
}

// ToolRegistry is the top-level registry file structure.
// The Timestamp field (format: "2006-01-02-15-04-05") is used to compare
// local vs remote registries — the newer one wins.
type ToolRegistry struct {
	Timestamp string `json:"timestamp"`
	Version   int    `json:"version"`
	Tools     []Tool `json:"tools"`
}

// IsNewerThan returns true if r has a newer timestamp than other.
func (r *ToolRegistry) IsNewerThan(other *ToolRegistry) bool {
	const layout = "2006-01-02-15-04-05"
	t1, e1 := time.Parse(layout, r.Timestamp)
	t2, e2 := time.Parse(layout, other.Timestamp)
	if e1 != nil || e2 != nil {
		return false
	}
	return t1.After(t2)
}
