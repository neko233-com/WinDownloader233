package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neko233/WinDownloader233/i18n"
	"github.com/neko233/WinDownloader233/installer"
	"github.com/neko233/WinDownloader233/mirror"
	"github.com/neko233/WinDownloader233/registry"
)

// AppService is the main Wails service exposing backend functionality
// to the frontend via auto-generated bindings.
type AppService struct {
	registryMgr  *registry.Manager
	installerMgr *installer.Manager
	mirrorMgr    *mirror.Manager
	downloader   *installer.Downloader
}

// NewAppService initializes all backend managers.
func NewAppService() *AppService {
	downloadDir := filepath.Join(os.TempDir(), "WinDownloader233_downloads")

	mirrorMgr := mirror.NewManager(mirror.DefaultConfig())
	registryMgr := registry.NewManager(
		"https://raw.githubusercontent.com/neko233/WinDownloader233-registry/main/tools.json",
	)
	installerMgr := installer.NewManager()
	downloader := installer.NewDownloader(installerMgr, mirrorMgr, downloadDir)

	return &AppService{
		registryMgr:  registryMgr,
		installerMgr: installerMgr,
		mirrorMgr:    mirrorMgr,
		downloader:   downloader,
	}
}

// --- Registry ---

// InitRegistry loads embedded registry and attempts remote sync.
func (s *AppService) InitRegistry() (string, error) {
	if err := s.registryMgr.LoadEmbedded(); err != nil {
		return "", fmt.Errorf("load embedded: %w", err)
	}

	updated, err := s.registryMgr.SyncRemote()
	if err != nil {
		return i18n.T("使用本地数据（远程同步失败）", "Using local data (remote sync failed)"), nil
	}
	if updated {
		return i18n.T("工具列表已从远程更新", "Registry updated from remote"), nil
	}
	return i18n.T("工具列表已加载", "Registry loaded"), nil
}

// ToolDTO is the data transfer object for tools sent to the frontend.
type ToolDTO struct {
	ID          string `json:"id"`
	NameZH      string `json:"nameZh"`
	NameEN      string `json:"nameEn"`
	DescZH      string `json:"descZh"`
	DescEN      string `json:"descEn"`
	Category    string `json:"category"`
	Tags        []string `json:"tags"`
	Icon        string `json:"icon"`
	Version     string `json:"version"`
	Size        string `json:"size"`
	WingetID    string `json:"wingetId"`
	DownloadURL string `json:"downloadUrl"`
	MirrorURL   string `json:"mirrorUrl"`
	Homepage    string `json:"homepage"`
	IsFree      bool   `json:"isFree"`
	InstallType string `json:"installType"`
	Installed   bool   `json:"installed"`
}

func toolToDTO(t registry.Tool) ToolDTO {
	return ToolDTO{
		ID:          t.ID,
		NameZH:      t.Name.ZH,
		NameEN:      t.Name.EN,
		DescZH:      t.Description.ZH,
		DescEN:      t.Description.EN,
		Category:    t.Category,
		Tags:        t.Tags,
		Icon:        t.Icon,
		Version:     t.Version,
		Size:        t.Size,
		WingetID:    t.WingetID,
		DownloadURL: t.DownloadURL,
		MirrorURL:   t.MirrorURL,
		Homepage:    t.Homepage,
		IsFree:      t.IsFree,
		InstallType: t.InstallType,
		Installed:   installer.IsWingetInstalled(t.WingetID),
	}
}

// GetToolsByCategory returns tools for a given category.
func (s *AppService) GetToolsByCategory(category string) []ToolDTO {
	tools := s.registryMgr.GetByCategory(category)
	out := make([]ToolDTO, len(tools))
	for i, t := range tools {
		out[i] = toolToDTO(t)
	}
	return out
}

// GetAllTools returns all tools in the registry.
func (s *AppService) GetAllTools() []ToolDTO {
	tools := s.registryMgr.GetAll()
	out := make([]ToolDTO, len(tools))
	for i, t := range tools {
		out[i] = toolToDTO(t)
	}
	return out
}

// SearchTools searches tools by query string.
func (s *AppService) SearchTools(query string) []ToolDTO {
	tools := s.registryMgr.Search(query)
	out := make([]ToolDTO, len(tools))
	for i, t := range tools {
		out[i] = toolToDTO(t)
	}
	return out
}

// GetToolsByTags returns tools matching all given tags.
func (s *AppService) GetToolsByTags(tags []string) []ToolDTO {
	tools := s.registryMgr.GetByTags(tags)
	out := make([]ToolDTO, len(tools))
	for i, t := range tools {
		out[i] = toolToDTO(t)
	}
	return out
}

// GetAllTags returns all unique tags.
func (s *AppService) GetAllTags() []string {
	return s.registryMgr.GetAllTags()
}

// RefreshRegistry re-fetches the remote registry.
func (s *AppService) RefreshRegistry() (string, error) {
	updated, err := s.registryMgr.SyncRemote()
	if err != nil {
		return i18n.T("同步失败", "Sync failed") + ": " + err.Error(), err
	}
	if updated {
		return i18n.T("工具列表已更新", "Registry updated"), nil
	}
	return i18n.T("已是最新版本", "Already up to date"), nil
}

// --- Installation ---

// InstallTool installs a tool by its ID (uses winget or download).
func (s *AppService) InstallTool(toolID string) (string, error) {
	tool := s.registryMgr.GetByID(toolID)
	if tool == nil {
		return "", fmt.Errorf("tool not found: %s", toolID)
	}

	switch tool.InstallType {
	case "winget":
		go func() {
			_ = s.installerMgr.InstallWinget(toolID, tool.WingetID)
		}()
		return i18n.T("开始安装", "Starting installation"), nil

	case "download":
		if tool.DownloadURL == "" {
			return "", fmt.Errorf("no download URL for %s", toolID)
		}
		go func() {
			filename := filepath.Base(tool.DownloadURL)
			_, _ = s.downloader.Download(toolID, tool.DownloadURL, filename)
		}()
		return i18n.T("开始下载", "Starting download"), nil

	default:
		return i18n.T("请手动安装", "Please install manually") + ": " + tool.Homepage, nil
	}
}

// UninstallTool removes a tool via winget.
func (s *AppService) UninstallTool(toolID string) (string, error) {
	tool := s.registryMgr.GetByID(toolID)
	if tool == nil {
		return "", fmt.Errorf("tool not found: %s", toolID)
	}
	if tool.WingetID == "" {
		return i18n.T("该工具不支持自动卸载", "This tool doesn't support auto-uninstall"), nil
	}

	go func() {
		_ = s.installerMgr.UninstallWinget(toolID, tool.WingetID)
	}()
	return i18n.T("开始卸载", "Starting uninstall"), nil
}

// GetInstallProgress returns the current progress for a tool.
func (s *AppService) GetInstallProgress(toolID string) installer.Progress {
	return s.installerMgr.GetProgress(toolID)
}

// GetAllInstallProgress returns progress for all tracked tools.
func (s *AppService) GetAllInstallProgress() []installer.Progress {
	return s.installerMgr.GetAllProgress()
}

// IsToolInstalled checks if a tool is installed via winget.
func (s *AppService) IsToolInstalled(toolID string) bool {
	tool := s.registryMgr.GetByID(toolID)
	if tool == nil {
		return false
	}
	return installer.IsWingetInstalled(tool.WingetID)
}

// --- Settings ---

// SetLanguage sets the display language ("zh" or "en").
func (s *AppService) SetLanguage(lang string) {
	i18n.SetLang(i18n.Lang(lang))
}

// GetLanguage returns the current language.
func (s *AppService) GetLanguage() string {
	return string(i18n.GetLang())
}

// GetMirrorConfig returns the current mirror configuration.
func (s *AppService) GetMirrorConfig() mirror.MirrorConfig {
	return s.mirrorMgr.GetConfig()
}

// SetMirrorConfig updates the mirror configuration.
func (s *AppService) SetMirrorConfig(cfg mirror.MirrorConfig) {
	s.mirrorMgr.SetConfig(cfg)
}

// GetAvailableProxies returns known GitHub proxy services.
func (s *AppService) GetAvailableProxies() []mirror.ProxyOption {
	return mirror.AvailableProxies()
}

// SetRemoteRegistryURL sets a custom remote registry URL.
func (s *AppService) SetRemoteRegistryURL(url string) {
	s.registryMgr.SetRemoteURL(url)
}

// GetRegistryTimestamp returns the current registry timestamp.
func (s *AppService) GetRegistryTimestamp() string {
	return s.registryMgr.GetTimestamp()
}

// GetCategoryNames returns localized category names.
func (s *AppService) GetCategoryNames() map[string]string {
	names := i18n.CategoryNames()
	result := make(map[string]string, len(names))
	lang := i18n.GetLang()
	for k, v := range names {
		if lang == i18n.ZH {
			result[k] = v.ZH
		} else {
			result[k] = v.EN
		}
	}
	return result
}

// GetUIStrings returns localized UI strings.
func (s *AppService) GetUIStrings() map[string]string {
	strs := i18n.UI()
	result := make(map[string]string, len(strs))
	lang := i18n.GetLang()
	for k, v := range strs {
		if lang == i18n.ZH {
			result[k] = v.ZH
		} else {
			result[k] = v.EN
		}
	}
	return result
}
