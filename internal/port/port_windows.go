//go:build windows

package port

import (
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func Lookup(p int) (*Process, error) {
	all, err := ListAll()
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Port == p {
			return &all[i], nil
		}
	}
	return nil, nil
}

func ListAll() ([]Process, error) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}
	rows := parseNetstat(string(out))

	// Resolve image names once per PID.
	cache := map[int]string{}
	var mu sync.Mutex
	var procs []Process
	for _, r := range rows {
		mu.Lock()
		name, ok := cache[r.pid]
		mu.Unlock()
		if !ok {
			name = tasklistName(r.pid)
			mu.Lock()
			cache[r.pid] = name
			mu.Unlock()
		}
		procs = append(procs, Process{
			PID:     r.pid,
			Command: name,
			User:    "",
			Port:    r.port,
		})
	}
	return procs, nil
}

func Kill(pid int, force bool) error {
	args := []string{"/PID", strconv.Itoa(pid)}
	if force {
		args = append(args, "/F")
	}
	out, err := exec.Command("taskkill", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskkill %d: %w: %s", pid, err, strings.TrimSpace(string(out)))
	}
	return nil
}

type netstatRow struct {
	port int
	pid  int
}

func parseNetstat(out string) []netstatRow {
	var rows []netstatRow
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if !strings.EqualFold(fields[0], "TCP") {
			continue
		}
		if !strings.EqualFold(fields[3], "LISTENING") {
			continue
		}
		port := parseNetstatPort(fields[1])
		if port == 0 {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			continue
		}
		rows = append(rows, netstatRow{port: port, pid: pid})
	}
	return rows
}

// parseNetstatPort handles "0.0.0.0:135", "[::]:80", "127.0.0.1:3000".
func parseNetstatPort(addr string) int {
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		if n, err := strconv.Atoi(addr[i+1:]); err == nil {
			return n
		}
	}
	return 0
}

// tasklistName returns the image name for a PID, or "?" if unknown.
func tasklistName(pid int) string {
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return "?"
	}
	r := csv.NewReader(strings.NewReader(string(out)))
	record, err := r.Read()
	if err != nil || len(record) == 0 {
		return "?"
	}
	return record[0]
}
