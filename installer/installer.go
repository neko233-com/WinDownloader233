package installer

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// Progress tracks installation progress for a tool.
type Progress struct {
	ToolID  string  `json:"toolId"`
	Status  string  `json:"status"`  // "idle" | "downloading" | "installing" | "done" | "error"
	Percent float64 `json:"percent"` // 0-100
	Message string  `json:"message"` // Human-readable status
}

// Manager tracks installation progress for multiple tools.
type Manager struct {
	mu        sync.RWMutex
	progress  map[string]*Progress
	listeners []func(Progress)
}

// NewManager creates an installer manager.
func NewManager() *Manager {
	return &Manager{
		progress: make(map[string]*Progress),
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

// UpdateProgress updates the progress for a tool and notifies listeners.
func (m *Manager) UpdateProgress(toolID, status string, percent float64, message string) {
	m.mu.Lock()
	m.progress[toolID] = &Progress{
		ToolID:  toolID,
		Status:  status,
		Percent: percent,
		Message: message,
	}
	p := *m.progress[toolID]
	m.mu.Unlock()
	m.emit(p)
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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.UpdateProgress(toolID, "error", 0, err.Error())
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.UpdateProgress(toolID, "error", 0, err.Error())
		return err
	}

	if err := cmd.Start(); err != nil {
		m.UpdateProgress(toolID, "error", 0, err.Error())
		return fmt.Errorf("start winget: %w", err)
	}

	// Read stdout for progress
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "%") {
				pct := parseWingetPercent(line)
				m.UpdateProgress(toolID, "installing", pct, line)
			}
		}
	}()

	// Read stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				m.UpdateProgress(toolID, "installing", -1, line)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		m.UpdateProgress(toolID, "error", 0, fmt.Sprintf("winget failed: %v", err))
		return fmt.Errorf("winget install failed: %w", err)
	}

	m.UpdateProgress(toolID, "done", 100, "Installation complete")
	return nil
}

// IsWingetInstalled checks if a package is already installed via winget.
func IsWingetInstalled(wingetID string) bool {
	if wingetID == "" {
		return false
	}
	cmd := exec.Command("winget", "list", "--id", wingetID, "--disable-interactivity")
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
	if err := cmd.Run(); err != nil {
		m.UpdateProgress(toolID, "error", 0, fmt.Sprintf("uninstall failed: %v", err))
		return err
	}

	m.UpdateProgress(toolID, "done", 100, "Uninstalled")
	return nil
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
