package i18n

import "sync"

// Lang represents a supported language.
type Lang string

const (
	ZH Lang = "zh"
	EN Lang = "en"
)

var (
	mu       sync.RWMutex
	current  Lang = ZH
)

// SetLang sets the current display language.
func SetLang(l Lang) {
	mu.Lock()
	defer mu.Unlock()
	current = l
}

// GetLang returns the current display language.
func GetLang() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// T returns the string for the current language.
func T(zh, en string) string {
	if GetLang() == ZH {
		return zh
	}
	return en
}

// CategoryNames returns display names for all categories.
func CategoryNames() map[string]LocalizedString {
	return map[string]LocalizedString{
		"programming": {ZH: "程序开发", EN: "Programming"},
		"art":         {ZH: "美术设计", EN: "Art & Design"},
		"planning":    {ZH: "策划文档", EN: "Planning"},
		"audio":       {ZH: "音频制作", EN: "Audio"},
		"qa":          {ZH: "测试 QA", EN: "QA & Testing"},
		"pm":          {ZH: "项目管理", EN: "Project Mgmt"},
		"ai":          {ZH: "AI 环境", EN: "AI Environment"},
	}
}

// UI returns common UI strings.
func UI() map[string]LocalizedString {
	return map[string]LocalizedString{
		"search":         {ZH: "搜索工具...", EN: "Search tools..."},
		"install":        {ZH: "安装", EN: "Install"},
		"installed":      {ZH: "已安装", EN: "Installed"},
		"installing":     {ZH: "安装中...", EN: "Installing..."},
		"download":       {ZH: "下载", EN: "Download"},
		"downloading":    {ZH: "下载中...", EN: "Downloading..."},
		"uninstall":      {ZH: "卸载", EN: "Uninstall"},
		"settings":       {ZH: "设置", EN: "Settings"},
		"language":       {ZH: "语言", EN: "Language"},
		"mirror":         {ZH: "镜像加速", EN: "Mirror Proxy"},
		"refresh":        {ZH: "刷新列表", EN: "Refresh"},
		"all_tags":       {ZH: "全部标签", EN: "All Tags"},
		"free":           {ZH: "免费", EN: "Free"},
		"paid":           {ZH: "付费", EN: "Paid"},
		"version":        {ZH: "版本", EN: "Version"},
		"size":           {ZH: "大小", EN: "Size"},
		"homepage":       {ZH: "官网", EN: "Homepage"},
		"no_results":     {ZH: "没有找到匹配的工具", EN: "No matching tools found"},
		"registry_updated": {ZH: "工具列表已更新", EN: "Registry updated"},
		"sync_failed":    {ZH: "同步失败，使用本地数据", EN: "Sync failed, using local data"},
	}
}

// LocalizedString holds bilingual text.
type LocalizedString struct {
	ZH string `json:"zh"`
	EN string `json:"en"`
}

// Localize returns the string for the current language.
func (ls LocalizedString) Localize() string {
	return T(ls.ZH, ls.EN)
}

// LocalizeToolString returns the string from registry's LocalizedString for the current language.
func LocalizeToolString(zh, en string) string {
	return T(zh, en)
}
