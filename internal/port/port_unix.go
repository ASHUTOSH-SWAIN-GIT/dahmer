//go:build darwin || linux

package port

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func Lookup(p int) (*Process, error) {
	out, err := exec.Command("lsof", "-iTCP:"+strconv.Itoa(p), "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}
	procs := parseLsof(string(out))
	for i := range procs {
		if procs[i].Port == p {
			return &procs[i], nil
		}
	}
	if len(procs) > 0 {
		procs[0].Port = p
		return &procs[0], nil
	}
	return nil, nil
}

func ListAll() ([]Process, error) {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}
	return parseLsof(string(out)), nil
}

func Kill(pid int, force bool) error {
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	if err := syscall.Kill(pid, sig); err != nil {
		return fmt.Errorf("kill %d: %w", pid, err)
	}
	return nil
}

func parseLsof(out string) []Process {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return nil
	}
	var procs []Process
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		port := parseLsofPort(fields[8])
		if port == 0 {
			continue
		}
		procs = append(procs, Process{
			PID:     pid,
			Command: fields[0],
			User:    fields[2],
			Port:    port,
		})
	}
	return procs
}

// parseLsofPort extracts the port from lsof NAME column entries like
// "127.0.0.1:3000", "[::1]:8080", or "*:443".
func parseLsofPort(name string) int {
	if i := strings.LastIndex(name, ":"); i >= 0 {
		tail := name[i+1:]
		if j := strings.Index(tail, " "); j >= 0 {
			tail = tail[:j]
		}
		if j := strings.Index(tail, "-"); j >= 0 {
			tail = tail[:j]
		}
		if n, err := strconv.Atoi(tail); err == nil {
			return n
		}
	}
	return 0
}
