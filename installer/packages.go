package installer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

// PackageManagerInfo describes one supported CLI package manager.
type PackageManagerInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Version     string `json:"version"`
	InstallHint string `json:"installHint"`
}

// PackageInfo is a normalized package row across supported managers.
type PackageInfo struct {
	Manager     string `json:"manager"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Available   string `json:"available"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
	Update      bool   `json:"update"`
}

// PackageAction is used by bulk import/install/update/uninstall flows.
type PackageAction struct {
	Manager string `json:"manager"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Options string `json:"options"`
}

// PackageBundle is the portable export format.
type PackageBundle struct {
	App      string          `json:"app"`
	Exported string          `json:"exported"`
	Packages []PackageAction `json:"packages"`
}

// ListPackageManagers reports installed package manager CLIs.
func ListPackageManagers() []PackageManagerInfo {
	return cachedPackageManagers(false)
}

func listPackageManagersLive() []PackageManagerInfo {
	managers := []PackageManagerInfo{
		managerInfo("winget", "WinGet", "winget", []string{"--version"}, "Install App Installer from Microsoft Store"),
		managerInfo("scoop", "Scoop", "scoop", []string{"--version"}, "Run: irm get.scoop.sh | iex"),
		managerInfo("choco", "Chocolatey", "choco", []string{"--version"}, "Run: winget install Chocolatey.Chocolatey"),
		managerInfo("pip", "pip", "python", []string{"-m", "pip", "--version"}, "Install Python with pip"),
		managerInfo("npm", "npm", "npm", []string{"--version"}, "Install Node.js"),
		managerInfo("bun", "Bun", "bun", []string{"--version"}, "Run: powershell -c irm bun.sh/install.ps1 | iex"),
		managerInfo("dotnet", ".NET Tool", "dotnet", []string{"--version"}, "Install .NET SDK"),
		managerInfo("psgallery", "PowerShell Gallery", "powershell", []string{"-NoProfile", "-Command", "$PSVersionTable.PSVersion.ToString()"}, "Install PowerShell/PowerShellGet"),
	}
	return managers
}

func managerInfo(id, name, bin string, args []string, hint string) PackageManagerInfo {
	version, err := runText(bin, args...)
	return PackageManagerInfo{
		ID:          id,
		Name:        name,
		Available:   err == nil,
		Version:     firstLine(version),
		InstallHint: hint,
	}
}

// SearchPackages searches one manager. Empty manager searches available managers.
func SearchPackages(manager, query string, limit int) []PackageInfo {
	if limit <= 0 {
		limit = 50
	}
	var out []PackageInfo
	for _, m := range selectedManagers(manager) {
		pkgs, _ := searchManager(m, query, limit)
		out = append(out, pkgs...)
		if len(out) >= limit && manager != "" {
			break
		}
	}
	if len(out) > limit && manager != "" {
		return out[:limit]
	}
	return out
}

// ListInstalledPackages returns installed packages from one or all managers.
func ListInstalledPackages(manager string) []PackageInfo {
	return cachedPackageRows("installed", manager, false)
}

// RefreshInstalledPackages refreshes installed package cache from local CLIs.
func RefreshInstalledPackages(manager string) []PackageInfo {
	return cachedPackageRows("installed", manager, true)
}

// RefreshPackageUpdates refreshes update package cache from local CLIs.
func RefreshPackageUpdates(manager string) []PackageInfo {
	return cachedPackageRows("updates", manager, true)
}

func listInstalledPackagesLive(manager string) []PackageInfo {
	return collectPackageRows(manager, installedManager)
}

// ListPackageUpdates returns available package updates from one or all managers.
func ListPackageUpdates(manager string) []PackageInfo {
	return cachedPackageRows("updates", manager, false)
}

func listPackageUpdatesLive(manager string) []PackageInfo {
	return collectPackageRows(manager, updatesManager)
}

func collectPackageRows(manager string, load func(string) ([]PackageInfo, error)) []PackageInfo {
	var out []PackageInfo
	for _, m := range selectedManagers(manager) {
		pkgs, _ := load(m)
		out = append(out, pkgs...)
	}
	sortPackages(out)
	return out
}

func selectedManagers(manager string) []string {
	if manager != "" && manager != "all" {
		return []string{manager}
	}
	ids := []string{}
	for _, m := range ListPackageManagers() {
		if m.Available {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

func searchManager(manager, query string, limit int) ([]PackageInfo, error) {
	switch manager {
	case "winget":
		args := []string{"search", query, "--accept-source-agreements", "--disable-interactivity"}
		out, err := runText("winget", args...)
		return parseWingetTable(out, "winget", false, false, limit), err
	case "choco":
		out, err := runText("choco", "search", query, "--limit-output")
		return parsePiped(out, "choco", false, false, limit), err
	case "scoop":
		out, err := runText("scoop", "search", query)
		return parseScoopSearch(out, limit), err
	case "pip":
		return nil, errors.New("pip search disabled upstream")
	case "npm":
		out, err := runText("npm", "search", query, "--json")
		return parseNPMSearch(out, limit), err
	case "bun":
		out, err := runText("bun", "pm", "view", query)
		return []PackageInfo{{Manager: "bun", ID: query, Name: query, Description: strings.TrimSpace(out)}}, err
	case "dotnet":
		out, err := runText("dotnet", "tool", "search", query, "--take", fmt.Sprint(limit))
		return parseDotnetSearch(out, limit), err
	case "psgallery":
		script := fmt.Sprintf("Find-Module -Name '*%s*' -Repository PSGallery | Select-Object -First %d Name,Version,Description | ConvertTo-Json", escapePS(query), limit)
		out, err := runText("powershell", "-NoProfile", "-Command", script)
		return parsePSGallery(out, limit), err
	default:
		return nil, fmt.Errorf("unsupported manager %q", manager)
	}
}

func installedManager(manager string) ([]PackageInfo, error) {
	switch manager {
	case "winget":
		out, err := runText("winget", "list", "--disable-interactivity")
		return parseWingetTable(out, "winget", true, false, 0), err
	case "choco":
		out, err := runText("choco", "list", "--local-only", "--limit-output")
		return parsePiped(out, "choco", true, false, 0), err
	case "scoop":
		out, err := runText("scoop", "list")
		return parseScoopList(out, false), err
	case "pip":
		out, err := runText("python", "-m", "pip", "list", "--format=json")
		return parsePipList(out, false), err
	case "npm":
		out, err := runText("npm", "list", "-g", "--depth=0", "--json")
		return parseNPMInstalled(out), err
	case "bun":
		return nil, nil
	case "dotnet":
		out, err := runText("dotnet", "tool", "list", "--global")
		return parseDotnetList(out, false), err
	case "psgallery":
		out, err := runText("powershell", "-NoProfile", "-Command", "Get-InstalledModule | Select-Object Name,Version,Description | ConvertTo-Json")
		return parsePSGalleryInstalled(out), err
	default:
		return nil, fmt.Errorf("unsupported manager %q", manager)
	}
}

func updatesManager(manager string) ([]PackageInfo, error) {
	switch manager {
	case "winget":
		out, err := runText("winget", "upgrade", "--accept-source-agreements", "--disable-interactivity")
		return parseWingetTable(out, "winget", true, true, 0), err
	case "choco":
		out, err := runText("choco", "outdated", "--limit-output")
		return parseChocoOutdated(out), err
	case "scoop":
		out, err := runText("scoop", "status")
		return parseScoopList(out, true), err
	case "pip":
		out, err := runText("python", "-m", "pip", "list", "--outdated", "--format=json")
		return parsePipList(out, true), err
	case "npm":
		out, err := runText("npm", "outdated", "-g", "--json")
		return parseNPMOutdated(out), err
	case "bun":
		return nil, nil
	case "dotnet":
		return nil, nil
	case "psgallery":
		out, err := runText("powershell", "-NoProfile", "-Command", "Get-InstalledModule | ForEach-Object { $l = Find-Module -Name $_.Name -Repository PSGallery -ErrorAction SilentlyContinue; if ($l.Version -gt $_.Version) { [pscustomobject]@{Name=$_.Name;Version=$_.Version;Available=$l.Version;Description=$_.Description} } } | ConvertTo-Json")
		return parsePSGalleryUpdates(out), err
	default:
		return nil, fmt.Errorf("unsupported manager %q", manager)
	}
}

// RunPackageAction executes install/update/uninstall.
func (m *Manager) RunPackageAction(action, manager, id, options string) error {
	key := manager + ":" + id
	m.UpdateProgress(key, "installing", 5, fmt.Sprintf("%s %s %s", manager, action, id))
	var cmd *exec.Cmd
	args := splitOptions(options)
	switch manager {
	case "winget":
		base := map[string][]string{
			"install":   []string{"install", "--id", id, "--accept-source-agreements", "--accept-package-agreements", "--disable-interactivity"},
			"update":    []string{"upgrade", "--id", id, "--accept-source-agreements", "--accept-package-agreements", "--disable-interactivity"},
			"uninstall": []string{"uninstall", "--id", id, "--disable-interactivity"},
		}[action]
		cmd = exec.Command("winget", append(base, args...)...)
	case "choco":
		base := map[string][]string{"install": []string{"install", id, "-y"}, "update": []string{"upgrade", id, "-y"}, "uninstall": []string{"uninstall", id, "-y"}}[action]
		cmd = exec.Command("choco", append(base, args...)...)
	case "scoop":
		base := map[string][]string{"install": []string{"install", id}, "update": []string{"update", id}, "uninstall": []string{"uninstall", id}}[action]
		cmd = exec.Command("scoop", append(base, args...)...)
	case "pip":
		base := map[string][]string{"install": []string{"-m", "pip", "install", id}, "update": []string{"-m", "pip", "install", "--upgrade", id}, "uninstall": []string{"-m", "pip", "uninstall", "-y", id}}[action]
		cmd = exec.Command("python", append(base, args...)...)
	case "npm":
		base := map[string][]string{"install": []string{"install", "-g", id}, "update": []string{"update", "-g", id}, "uninstall": []string{"uninstall", "-g", id}}[action]
		cmd = exec.Command("npm", append(base, args...)...)
	case "bun":
		base := map[string][]string{"install": []string{"add", "-g", id}, "update": []string{"update", "-g", id}, "uninstall": []string{"remove", "-g", id}}[action]
		cmd = exec.Command("bun", append(base, args...)...)
	case "dotnet":
		base := map[string][]string{"install": []string{"tool", "install", "--global", id}, "update": []string{"tool", "update", "--global", id}, "uninstall": []string{"tool", "uninstall", "--global", id}}[action]
		cmd = exec.Command("dotnet", append(base, args...)...)
	case "psgallery":
		verb := map[string]string{"install": "Install-Module", "update": "Update-Module", "uninstall": "Uninstall-Module"}[action]
		script := fmt.Sprintf("%s -Name '%s' -Force -AcceptLicense", verb, escapePS(id))
		cmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	default:
		return fmt.Errorf("unsupported manager %q", manager)
	}
	if cmd == nil {
		return fmt.Errorf("unsupported action %q", action)
	}
	if err := m.runCommand(key, "installing", cmd); err != nil {
		return err
	}
	InvalidatePackageCache(manager)
	m.UpdateProgress(key, "done", 100, "Completed")
	m.AppendLog(key, "info", "Completed")
	return nil
}

func ExportBundle(pkgs []PackageInfo) (string, error) {
	actions := make([]PackageAction, 0, len(pkgs))
	for _, p := range pkgs {
		actions = append(actions, PackageAction{Manager: p.Manager, ID: p.ID, Name: p.Name})
	}
	b := PackageBundle{App: "WinDownloader233", Exported: time.Now().Format(time.RFC3339), Packages: actions}
	data, err := json.MarshalIndent(b, "", "  ")
	return string(data), err
}

func ImportBundle(raw string) ([]PackageAction, error) {
	var bundle PackageBundle
	if err := json.Unmarshal([]byte(raw), &bundle); err != nil {
		return nil, err
	}
	return bundle.Packages, nil
}

func runText(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	hideCommandWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	text := stdout.String()
	if text == "" {
		text = stderr.String()
	}
	return strings.TrimSpace(text), err
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Split(s, "\n")[0]
}

func sortPackages(pkgs []PackageInfo) {
	sort.Slice(pkgs, func(i, j int) bool {
		if pkgs[i].Manager == pkgs[j].Manager {
			return strings.ToLower(pkgs[i].Name) < strings.ToLower(pkgs[j].Name)
		}
		return pkgs[i].Manager < pkgs[j].Manager
	})
}

func parseWingetTable(out, manager string, installed, update bool, limit int) []PackageInfo {
	lines := cleanLines(out)
	header := -1
	for i, line := range lines {
		if strings.Contains(line, "Name") && strings.Contains(line, "Id") && strings.Contains(line, "Version") {
			header = i
			break
		}
	}
	if header < 0 || header+1 >= len(lines) {
		return nil
	}
	h := lines[header]
	idxName, idxID, idxVersion := strings.Index(h, "Name"), strings.Index(h, "Id"), strings.Index(h, "Version")
	idxAvail, idxSource := strings.Index(h, "Available"), strings.Index(h, "Source")
	var rows []PackageInfo
	for _, line := range lines[header+2:] {
		if limit > 0 && len(rows) >= limit {
			break
		}
		if strings.Contains(line, "---") || strings.TrimSpace(line) == "" {
			continue
		}
		name := sliceCol(line, idxName, idxID)
		id := sliceCol(line, idxID, idxVersion)
		version := sliceCol(line, idxVersion, firstPositive(idxAvail, idxSource, len(line)))
		available := ""
		source := ""
		if idxAvail >= 0 {
			available = sliceCol(line, idxAvail, firstPositive(idxSource, len(line)))
		}
		if idxSource >= 0 {
			source = sliceCol(line, idxSource, len(line))
		}
		if id == "" || strings.Contains(id, "upgrades available") {
			continue
		}
		rows = append(rows, PackageInfo{Manager: manager, ID: id, Name: name, Version: version, Available: available, Source: source, Installed: installed, Update: update})
	}
	return rows
}

func parsePiped(out, manager string, installed, update bool, limit int) []PackageInfo {
	var rows []PackageInfo
	for _, line := range cleanLines(out) {
		if limit > 0 && len(rows) >= limit {
			break
		}
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}
		rows = append(rows, PackageInfo{Manager: manager, ID: parts[0], Name: parts[0], Version: parts[1], Installed: installed, Update: update})
	}
	return rows
}

func parseChocoOutdated(out string) []PackageInfo {
	var rows []PackageInfo
	for _, line := range cleanLines(out) {
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		rows = append(rows, PackageInfo{Manager: "choco", ID: parts[0], Name: parts[0], Version: parts[1], Available: parts[2], Installed: true, Update: true})
	}
	return rows
}

func parseScoopSearch(out string, limit int) []PackageInfo {
	var rows []PackageInfo
	re := regexp.MustCompile(`^\s*([a-zA-Z0-9_.+-]+)\s+\(([^\)]+)\)`)
	for _, line := range cleanLines(out) {
		if limit > 0 && len(rows) >= limit {
			break
		}
		m := re.FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		rows = append(rows, PackageInfo{Manager: "scoop", ID: m[1], Name: m[1], Version: m[2]})
	}
	return rows
}

func parseScoopList(out string, update bool) []PackageInfo {
	var rows []PackageInfo
	for _, line := range cleanLines(out) {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.EqualFold(fields[0], "Name") {
			continue
		}
		p := PackageInfo{Manager: "scoop", ID: fields[0], Name: fields[0], Version: fields[1], Installed: true, Update: update}
		if update && len(fields) >= 3 {
			p.Available = fields[2]
		}
		rows = append(rows, p)
	}
	return rows
}

func parsePipList(out string, update bool) []PackageInfo {
	var raw []struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		LatestVersion string `json:"latest_version"`
	}
	if json.Unmarshal([]byte(out), &raw) != nil {
		return nil
	}
	rows := make([]PackageInfo, 0, len(raw))
	for _, r := range raw {
		rows = append(rows, PackageInfo{Manager: "pip", ID: r.Name, Name: r.Name, Version: r.Version, Available: r.LatestVersion, Installed: true, Update: update})
	}
	return rows
}

func parseNPMSearch(out string, limit int) []PackageInfo {
	var raw []struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}
	if json.Unmarshal([]byte(out), &raw) != nil {
		return nil
	}
	rows := make([]PackageInfo, 0, len(raw))
	for i, r := range raw {
		if limit > 0 && i >= limit {
			break
		}
		rows = append(rows, PackageInfo{Manager: "npm", ID: r.Name, Name: r.Name, Version: r.Version, Description: r.Description})
	}
	return rows
}

func parseNPMInstalled(out string) []PackageInfo {
	var raw struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if json.Unmarshal([]byte(out), &raw) != nil {
		return nil
	}
	rows := make([]PackageInfo, 0, len(raw.Dependencies))
	for name, dep := range raw.Dependencies {
		rows = append(rows, PackageInfo{Manager: "npm", ID: name, Name: name, Version: dep.Version, Installed: true})
	}
	return rows
}

func parseNPMOutdated(out string) []PackageInfo {
	var raw map[string]struct {
		Current string `json:"current"`
		Wanted  string `json:"wanted"`
		Latest  string `json:"latest"`
	}
	if json.Unmarshal([]byte(out), &raw) != nil {
		return nil
	}
	rows := make([]PackageInfo, 0, len(raw))
	for name, dep := range raw {
		rows = append(rows, PackageInfo{Manager: "npm", ID: name, Name: name, Version: dep.Current, Available: firstNonEmpty(dep.Latest, dep.Wanted), Installed: true, Update: true})
	}
	return rows
}

func parseDotnetSearch(out string, limit int) []PackageInfo {
	var rows []PackageInfo
	for _, line := range cleanLines(out) {
		if limit > 0 && len(rows) >= limit {
			break
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.Contains(strings.ToLower(line), "package id") || strings.Contains(line, "---") {
			continue
		}
		rows = append(rows, PackageInfo{Manager: "dotnet", ID: fields[0], Name: fields[0], Version: fields[1]})
	}
	return rows
}

func parseDotnetList(out string, update bool) []PackageInfo {
	var rows []PackageInfo
	for _, line := range cleanLines(out) {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.Contains(strings.ToLower(line), "package id") || strings.Contains(line, "---") {
			continue
		}
		rows = append(rows, PackageInfo{Manager: "dotnet", ID: fields[0], Name: fields[0], Version: fields[1], Installed: true, Update: update})
	}
	return rows
}

func parsePSGallery(out string, limit int) []PackageInfo {
	raw := normalizeJSONArray(out)
	var items []struct {
		Name        string `json:"Name"`
		Version     string `json:"Version"`
		Description string `json:"Description"`
	}
	if json.Unmarshal(raw, &items) != nil {
		return nil
	}
	rows := make([]PackageInfo, 0, len(items))
	for i, r := range items {
		if limit > 0 && i >= limit {
			break
		}
		rows = append(rows, PackageInfo{Manager: "psgallery", ID: r.Name, Name: r.Name, Version: r.Version, Description: r.Description})
	}
	return rows
}

func parsePSGalleryInstalled(out string) []PackageInfo {
	rows := parsePSGallery(out, 0)
	for i := range rows {
		rows[i].Installed = true
	}
	return rows
}

func parsePSGalleryUpdates(out string) []PackageInfo {
	raw := normalizeJSONArray(out)
	var items []struct {
		Name        string `json:"Name"`
		Version     string `json:"Version"`
		Available   string `json:"Available"`
		Description string `json:"Description"`
	}
	if json.Unmarshal(raw, &items) != nil {
		return nil
	}
	rows := make([]PackageInfo, 0, len(items))
	for _, r := range items {
		rows = append(rows, PackageInfo{Manager: "psgallery", ID: r.Name, Name: r.Name, Version: r.Version, Available: r.Available, Description: r.Description, Installed: true, Update: true})
	}
	return rows
}

func normalizeJSONArray(out string) []byte {
	out = strings.TrimSpace(out)
	if out == "" {
		return []byte("[]")
	}
	if strings.HasPrefix(out, "[") {
		return []byte(out)
	}
	return []byte("[" + out + "]")
}

func cleanLines(out string) []string {
	out = strings.ReplaceAll(out, "\r\n", "\n")
	raw := strings.Split(out, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func sliceCol(line string, start, end int) string {
	if start < 0 || start >= len(line) {
		return ""
	}
	if end < 0 || end > len(line) {
		end = len(line)
	}
	if end < start {
		end = len(line)
	}
	return strings.TrimSpace(line[start:end])
}

func firstPositive(nums ...int) int {
	for _, n := range nums {
		if n >= 0 {
			return n
		}
	}
	return -1
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func splitOptions(options string) []string {
	options = strings.TrimSpace(options)
	if options == "" {
		return nil
	}
	return strings.Fields(options)
}

func escapePS(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
