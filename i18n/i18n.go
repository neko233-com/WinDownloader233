package i18n

import "sync"

// Lang represents a supported language.
type Lang string

const (
	ZH Lang = "zh"
	EN Lang = "en"
)

var (
	mu      sync.RWMutex
	current Lang = ZH
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
		"search":              {ZH: "搜索工具...", EN: "Search tools..."},
		"app_subtitle":        {ZH: "软件、驱动、工具集、备份迁移一体化管理", EN: "Unified software, driver, toolkit and backup management"},
		"view_discover":       {ZH: "搜索软件", EN: "Search"},
		"view_installed":      {ZH: "已安装", EN: "Installed"},
		"view_updates":        {ZH: "可更新", EN: "Updates"},
		"view_drivers":        {ZH: "驱动下载", EN: "Drivers"},
		"view_bundle":         {ZH: "备份迁移", EN: "Backup"},
		"view_toolkit":        {ZH: "工具集", EN: "Toolkit"},
		"title_toolkit":       {ZH: "游戏开发工具集", EN: "Game Dev Toolkit"},
		"title_drivers":       {ZH: "驱动下载", EN: "Driver Downloads"},
		"title_packages":      {ZH: "软件包管理", EN: "Package Manager"},
		"desc_toolkit":        {ZH: "按分类安装常用开发工具", EN: "Install common game-dev tools by category"},
		"desc_drivers":        {ZH: "主板、显卡、鼠标、键盘、音频、网卡等官方驱动入口", EN: "Official driver links for boards, GPU, peripherals, audio and network"},
		"desc_packages":       {ZH: "搜索、安装、更新、卸载本机软件包", EN: "Search, install, update and remove local packages"},
		"install":             {ZH: "安装", EN: "Install"},
		"installed":           {ZH: "已安装", EN: "Installed"},
		"installing":          {ZH: "安装中...", EN: "Installing..."},
		"download":            {ZH: "下载", EN: "Download"},
		"downloading":         {ZH: "下载中...", EN: "Downloading..."},
		"uninstall":           {ZH: "卸载", EN: "Uninstall"},
		"settings":            {ZH: "设置", EN: "Settings"},
		"language":            {ZH: "语言", EN: "Language"},
		"mirror":              {ZH: "镜像加速", EN: "Mirror Proxy"},
		"refresh":             {ZH: "刷新列表", EN: "Refresh"},
		"all_tags":            {ZH: "全部标签", EN: "All Tags"},
		"free":                {ZH: "免费", EN: "Free"},
		"paid":                {ZH: "付费", EN: "Paid"},
		"version":             {ZH: "版本", EN: "Version"},
		"size":                {ZH: "大小", EN: "Size"},
		"homepage":            {ZH: "官网", EN: "Homepage"},
		"no_results":          {ZH: "没有找到匹配的工具", EN: "No matching tools found"},
		"loading":             {ZH: "加载中", EN: "Loading"},
		"all_sources":         {ZH: "全部来源", EN: "All sources"},
		"all":                 {ZH: "全部", EN: "All"},
		"manager":             {ZH: "来源", EN: "Manager"},
		"name":                {ZH: "名称", EN: "Name"},
		"available":           {ZH: "可用版本", EN: "Available"},
		"action":              {ZH: "操作", EN: "Action"},
		"running":             {ZH: "运行中", EN: "Running"},
		"update":              {ZH: "更新", EN: "Update"},
		"search_btn":          {ZH: "搜索", EN: "Search"},
		"install_selected":    {ZH: "安装选中", EN: "Install selected"},
		"update_selected":     {ZH: "更新选中", EN: "Update selected"},
		"uninstall_selected":  {ZH: "卸载选中", EN: "Uninstall selected"},
		"reload":              {ZH: "刷新", EN: "Reload"},
		"force_refresh":       {ZH: "强制刷新", EN: "Force refresh"},
		"check_updates":       {ZH: "检查更新", EN: "Check updates"},
		"export_backup":       {ZH: "导出备份", EN: "Export backup"},
		"export_installed":    {ZH: "导出已安装", EN: "Export installed"},
		"parse_import":        {ZH: "解析导入", EN: "Parse import"},
		"install_imported":    {ZH: "安装导入项", EN: "Install imported"},
		"package_placeholder": {ZH: "输入软件名，例如 vscode、git、python", EN: "Type a package name, e.g. vscode, git, python"},
		"package_options":     {ZH: "可选安装参数", EN: "optional install options"},
		"package_empty":       {ZH: "输入关键词后搜索包", EN: "Enter a query to search packages"},
		"no_package_results":  {ZH: "没有结果", EN: "No results"},
		"bundle_placeholder":  {ZH: "粘贴 WinDownloader233 备份 JSON", EN: "Paste WinDownloader233 package backup JSON"},
		"driver_placeholder":  {ZH: "搜索厂商 / 型号 / 别名，例如 RTX、ROG、G502、Realtek、声卡", EN: "Search vendor / model / alias, e.g. RTX, ROG, G502, Realtek, audio"},
		"driver_empty":        {ZH: "没有匹配驱动。试试厂商名、型号、设备类型或中文别名。", EN: "No matching drivers. Try vendor, model, device type or alias."},
		"official_download":   {ZH: "官方下载", EN: "Official download"},
		"support_page":        {ZH: "支持页", EN: "Support"},
		"tasks":               {ZH: "任务日志", EN: "Tasks"},
		"background_tasks":    {ZH: "后台任务", EN: "Background tasks"},
		"task_count":          {ZH: "个任务", EN: "tasks"},
		"running_count":       {ZH: "个运行中", EN: "running"},
		"no_tasks":            {ZH: "暂无后台任务", EN: "No background tasks"},
		"logs":                {ZH: "日志", EN: "Logs"},
		"refresh_logs":        {ZH: "刷新", EN: "Refresh"},
		"copy_logs":           {ZH: "复制日志", EN: "Copy logs"},
		"export_logs":         {ZH: "导出日志", EN: "Export logs"},
		"no_logs":             {ZH: "暂无日志；任务启动后会实时追加。", EN: "No logs yet. Logs stream here after task starts."},
		"logs_copied":         {ZH: "日志已复制", EN: "Logs copied"},
		"global_search_hint":  {ZH: "按 Ctrl+K 全局搜索", EN: "Press Ctrl+K for global search"},
		"registry_updated":    {ZH: "工具列表已更新", EN: "Registry updated"},
		"sync_failed":         {ZH: "同步失败，使用本地数据", EN: "Sync failed, using local data"},
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
