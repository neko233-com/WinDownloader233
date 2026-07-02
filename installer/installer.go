package installer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Progress tracks installation progress for a tool.
type Progress struct {
	ToolID    string  `json:"toolId"`
	Status    string  `json:"status"`  // "idle" | "downloading" | "installing" | "done" | "error"
	Percent   float64 `json:"percent"` // 0-100, -1 means unknown
	Message   string  `json:"message"` // Human-readable status
	StartedAt string  `json:"startedAt"`
	UpdatedAt string  `json:"updatedAt"`
	LogLines  int     `json:"logLines"`
	ExitCode  int     `json:"exitCode"`
}

// LogEntry is one user-shareable installation/debug log line.
type LogEntry struct {
	Time    string `json:"time"`
	ToolID  string `json:"toolId"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// Manager tracks installation progress for multiple tools.
type Manager struct {
	mu        sync.RWMutex
	progress  map[string]*Progress
	logs      map[string][]LogEntry
	logDir    string
	listeners []func(Progress)
}

// NewManager creates an installer manager.
func NewManager() *Manager {
	logDir := filepath.Join(DefaultCacheDir(), "logs")
	_ = os.MkdirAll(logDir, 0755)
	return &Manager{
		progress: make(map[string]*Progress),
		logs:     make(map[string][]LogEntry),
		logDir:   logDir,
	}
}

// OnProgress registers a callback for progress updates.
func (m *Manager) OnProgress(fn func(Progress)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, fn)
}

func (m *Manager) emit(p Progress) {
	m.mu.RLock()
	listeners := make([]func(Progress), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()
	for _, fn := range listeners {
		fn(p)
	}
}

// AppendLog adds a debug log line in memory and on disk.
func (m *Manager) AppendLog(toolID, level, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	lines := cleanLines(message)
	if len(lines) == 0 {
		lines = []string{message}
	}
	for _, line := range lines {
		entry := LogEntry{
			Time:    time.Now().Format(time.RFC3339),
			ToolID:  toolID,
			Level:   level,
			Message: line,
		}
		m.mu.Lock()
		m.logs[toolID] = append(m.logs[toolID], entry)
		if len(m.logs[toolID]) > 800 {
			m.logs[toolID] = m.logs[toolID][len(m.logs[toolID])-800:]
		}
		lineCount := len(m.logs[toolID])
		if p, ok := m.progress[toolID]; ok {
			p.LogLines = lineCount
			p.UpdatedAt = entry.Time
		}
		m.mu.Unlock()

		_ = os.MkdirAll(m.logDir, 0755)
		f, err := os.OpenFile(m.logPath(toolID), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			_, _ = fmt.Fprintf(f, "%s [%s] %s\n", entry.Time, entry.Level, entry.Message)
			_ = f.Close()
		}
	}
}

// GetLogs returns recent in-memory logs for a task.
func (m *Manager) GetLogs(toolID string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rows := m.logs[toolID]
	out := make([]LogEntry, len(rows))
	copy(out, rows)
	return out
}

// ExportLog returns full disk log text when present, with memory fallback.
func (m *Manager) ExportLog(toolID string) string {
	if data, err := os.ReadFile(m.logPath(toolID)); err == nil {
		return string(data)
	}
	var b strings.Builder
	for _, row := range m.GetLogs(toolID) {
		_, _ = fmt.Fprintf(&b, "%s [%s] %s\n", row.Time, row.Level, row.Message)
	}
	return b.String()
}

// LogDirectory returns persistent log folder.
func (m *Manager) LogDirectory() string {
	return m.logDir
}

func (m *Manager) logPath(toolID string) string {
	return filepath.Join(m.logDir, safeFilename(toolID)+".log")
}

// UpdateProgress updates the progress for a tool and notifies listeners.
func (m *Manager) UpdateProgress(toolID, status string, percent float64, message string) {
	now := time.Now().Format(time.RFC3339)
	m.mu.Lock()
	startedAt := now
	logLines := len(m.logs[toolID])
	exitCode := 0
	if old, ok := m.progress[toolID]; ok {
		if old.StartedAt != "" {
			startedAt = old.StartedAt
		}
		exitCode = old.ExitCode
	}
	m.progress[toolID] = &Progress{
		ToolID:    toolID,
		Status:    status,
		Percent:   percent,
		Message:   message,
		StartedAt: startedAt,
		UpdatedAt: now,
		LogLines:  logLines,
		ExitCode:  exitCode,
	}
	p := *m.progress[toolID]
	m.mu.Unlock()
	m.emit(p)
}

// SetExitCode stores process exit code for diagnostics.
func (m *Manager) SetExitCode(toolID string, code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.progress[toolID]; ok {
		p.ExitCode = code
	}
}

// GetProgress returns the current progress for a tool.
func (m *Manager) GetProgress(toolID string) Progress {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.progress[toolID]; ok {
		return *p
	}
	return Progress{ToolID: toolID, Status: "idle"}
}

// GetAllProgress returns progress for all tracked tools.
func (m *Manager) GetAllProgress() []Progress {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Progress, 0, len(m.progress))
	for _, p := range m.progress {
		out = append(out, *p)
	}
	return out
}

// InstallWinget installs a tool using winget.
func (m *Manager) InstallWinget(toolID, wingetID string) error {
	if wingetID == "" {
		return fmt.Errorf("winget ID is empty for tool %s", toolID)
	}

	m.UpdateProgress(toolID, "installing", 10, fmt.Sprintf("winget install %s", wingetID))

	cmd := exec.Command("winget", "install", "--id", wingetID,
		"--accept-source-agreements", "--accept-package-agreements",
		"--silent", "--disable-interactivity")
	if err := m.runCommand(toolID, "installing", cmd); err != nil {
		return fmt.Errorf("winget install failed: %w", err)
	}

	m.UpdateProgress(toolID, "done", 100, "Installation complete")
	m.AppendLog(toolID, "info", "Installation complete")
	InvalidatePackageCache("winget")
	return nil
}

// IsWingetInstalled checks if a package is already installed via winget.
func IsWingetInstalled(wingetID string) bool {
	if wingetID == "" {
		return false
	}
	cmd := exec.Command("winget", "list", "--id", wingetID, "--disable-interactivity")
	hideCommandWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// winget list returns the package in the output if installed
	return strings.Contains(string(out), wingetID) &&
		!strings.Contains(string(out), "No installed package found")
}

// UninstallWinget removes a tool using winget.
func (m *Manager) UninstallWinget(toolID, wingetID string) error {
	if wingetID == "" {
		return fmt.Errorf("winget ID is empty for tool %s", toolID)
	}

	m.UpdateProgress(toolID, "installing", 50, fmt.Sprintf("winget uninstall %s", wingetID))

	cmd := exec.Command("winget", "uninstall", "--id", wingetID,
		"--silent", "--disable-interactivity")
	if err := m.runCommand(toolID, "installing", cmd); err != nil {
		return err
	}

	m.UpdateProgress(toolID, "done", 100, "Uninstalled")
	m.AppendLog(toolID, "info", "Uninstalled")
	InvalidatePackageCache("winget")
	return nil
}

func (m *Manager) runCommand(toolID, status string, cmd *exec.Cmd) error {
	hideCommandWindow(cmd)
	m.AppendLog(toolID, "command", commandLine(cmd))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.AppendLog(toolID, "error", err.Error())
		m.UpdateProgress(toolID, "error", 0, err.Error())
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.AppendLog(toolID, "error", err.Error())
		m.UpdateProgress(toolID, "error", 0, err.Error())
		return err
	}

	if err := cmd.Start(); err != nil {
		m.AppendLog(toolID, "error", err.Error())
		m.UpdateProgress(toolID, "error", 0, err.Error())
		return fmt.Errorf("start command: %w", err)
	}

	var wg sync.WaitGroup
	scan := func(r io.Reader, level string) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			m.AppendLog(toolID, level, line)
			pct := parseWingetPercent(line)
			if pct >= 0 {
				m.UpdateProgress(toolID, status, pct, line)
			} else if strings.TrimSpace(line) != "" {
				m.UpdateProgress(toolID, status, -1, line)
			}
		}
		if err := scanner.Err(); err != nil {
			m.AppendLog(toolID, "error", err.Error())
		}
	}
	wg.Add(2)
	go scan(stdout, "stdout")
	go scan(stderr, "stderr")

	err = cmd.Wait()
	wg.Wait()
	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		m.SetExitCode(toolID, exitCode)
		msg := fmt.Sprintf("command failed: exit %d: %v", exitCode, err)
		m.AppendLog(toolID, "error", msg)
		m.UpdateProgress(toolID, "error", 0, msg)
		return err
	}
	return nil
}

func commandLine(cmd *exec.Cmd) string {
	if len(cmd.Args) == 0 {
		return cmd.Path
	}
	return strings.Join(cmd.Args, " ")
}

func parseWingetPercent(line string) float64 {
	// winget output like: "  42%  ████..."
	line = strings.TrimSpace(line)
	idx := strings.Index(line, "%")
	if idx < 1 {
		return -1
	}
	// find the number before %
	start := idx - 1
	for start >= 0 && line[start] >= '0' && line[start] <= '9' {
		start--
	}
	start++
	numStr := line[start:idx]
	var pct float64
	fmt.Sscanf(numStr, "%f", &pct)
	return pct
}
