//go:build !windows

package installer

import "os/exec"

func hideCommandWindow(cmd *exec.Cmd) {}
