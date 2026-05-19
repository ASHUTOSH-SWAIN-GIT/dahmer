package port

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Process struct {
	PID     int
	Command string
	User    string
	Port    int
}

// Lookup returns the process listening on the given TCP port, or nil if none.
func Lookup(p int) (*Process, error) {
	out, err := exec.Command("lsof", "-iTCP:"+strconv.Itoa(p), "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, nil
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return nil, nil
	}
	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("parse pid: %w", err)
	}
	return &Process{
		PID:     pid,
		Command: fields[0],
		User:    fields[2],
		Port:    p,
	}, nil
}

// Kill terminates the given pid (SIGTERM, then SIGKILL fallback).
func Kill(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("SIGTERM: %w", err)
	}
	return nil
}

// ForceKill sends SIGKILL.
func ForceKill(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}
