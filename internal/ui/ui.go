package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ashutosh-swain-git/dahmer/internal/port"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5F87FF")).
			Padding(1, 2)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777")).
			Italic(true)

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD166")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#06D6A0")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF476F")).
			Bold(true)
)

type Mode int

const (
	ModeShow Mode = iota
	ModeKill
	ModeList
)

// Entry pairs a requested port with the process found on it (nil if free).
type Entry struct {
	Port int
	Proc *port.Process
}

type killResult struct {
	pid int
	err error
}

type model struct {
	mode    Mode
	entries []Entry
	force   bool
	results map[int]killResult // pid -> result
	pending int
	title   string
}

func NewShow(entries []Entry) tea.Model {
	return &model{mode: ModeShow, entries: entries, title: titleFor(entries, "show")}
}

func NewKill(entries []Entry, force bool) tea.Model {
	pending := 0
	for _, e := range entries {
		if e.Proc != nil {
			pending++
		}
	}
	verb := "kill"
	if force {
		verb = "force-kill"
	}
	return &model{
		mode:    ModeKill,
		entries: entries,
		force:   force,
		results: map[int]killResult{},
		pending: pending,
		title:   titleFor(entries, verb),
	}
}

func NewList(procs []port.Process) tea.Model {
	sort.Slice(procs, func(i, j int) bool { return procs[i].Port < procs[j].Port })
	entries := make([]Entry, len(procs))
	for i := range procs {
		p := procs[i]
		entries[i] = Entry{Port: p.Port, Proc: &p}
	}
	return &model{mode: ModeList, entries: entries, title: "dahmer :: listening ports"}
}

func titleFor(entries []Entry, verb string) string {
	if len(entries) == 1 {
		return fmt.Sprintf("dahmer :: %s port %d", verb, entries[0].Port)
	}
	return fmt.Sprintf("dahmer :: %s %d ports", verb, len(entries))
}

type killMsg killResult

func killCmd(pid int, force bool) tea.Cmd {
	return func() tea.Msg {
		return killMsg{pid: pid, err: port.Kill(pid, force)}
	}
}

func (m *model) Init() tea.Cmd {
	if m.mode != ModeKill {
		return tea.Quit
	}
	var cmds []tea.Cmd
	for _, e := range m.entries {
		if e.Proc != nil {
			cmds = append(cmds, killCmd(e.Proc.PID, m.force))
		}
	}
	if len(cmds) == 0 {
		return tea.Quit
	}
	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case killMsg:
		m.results[msg.pid] = killResult(msg)
		m.pending--
		if m.pending <= 0 {
			return m, tea.Quit
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() string {
	title := titleStyle.Render(m.title)

	if len(m.entries) == 0 {
		body := emptyStyle.Render("Nothing is listening.")
		return boxStyle.Render(title + "\n\n" + body)
	}

	if m.mode == ModeShow && len(m.entries) == 1 {
		return boxStyle.Render(title + "\n\n" + m.renderSingle(m.entries[0]) + "\n\n" + m.footerSingle(m.entries[0]))
	}

	return boxStyle.Render(title + "\n\n" + m.renderTable() + "\n\n" + m.footerMulti())
}

func (m *model) renderSingle(e Entry) string {
	if e.Proc == nil {
		return emptyStyle.Render(fmt.Sprintf("No process is listening on port %d.", e.Port))
	}
	row := func(k, v string) string {
		return labelStyle.Render(fmt.Sprintf("%-10s", k)) + valueStyle.Render(v)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		row("PID", fmt.Sprintf("%d", e.Proc.PID)),
		row("Command", e.Proc.Command),
		row("User", e.Proc.User),
	)
}

func (m *model) footerSingle(e Entry) string {
	if m.mode == ModeShow {
		if e.Proc == nil {
			return hintStyle.Render("nothing to do")
		}
		return hintStyle.Render(fmt.Sprintf("run `dahmer %d kill` to terminate it", e.Port))
	}
	// ModeKill
	if e.Proc == nil {
		return emptyStyle.Render("nothing to kill")
	}
	r, ok := m.results[e.Proc.PID]
	if !ok {
		return hintStyle.Render(signalLabel(m.force) + "…")
	}
	if r.err != nil {
		return errorStyle.Render("failed: " + r.err.Error())
	}
	return successStyle.Render(fmt.Sprintf("✓ killed PID %d", e.Proc.PID))
}

func (m *model) renderTable() string {
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render(pad("PORT", 7)),
		labelStyle.Render(pad("PID", 8)),
		labelStyle.Render(pad("COMMAND", 18)),
		labelStyle.Render(pad("USER", 12)),
		labelStyle.Render("STATUS"),
	)
	var rows []string
	rows = append(rows, header)
	for _, e := range m.entries {
		rows = append(rows, m.renderRow(e))
	}
	return strings.Join(rows, "\n")
}

func (m *model) renderRow(e Entry) string {
	portCell := valueStyle.Render(pad(fmt.Sprintf("%d", e.Port), 7))
	if e.Proc == nil {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			portCell,
			dimStyle.Render(pad("—", 8)),
			dimStyle.Render(pad("—", 18)),
			dimStyle.Render(pad("—", 12)),
			emptyStyle.Render("free"),
		)
	}
	status := m.rowStatus(e)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		portCell,
		valueStyle.Render(pad(fmt.Sprintf("%d", e.Proc.PID), 8)),
		valueStyle.Render(pad(trunc(e.Proc.Command, 17), 18)),
		valueStyle.Render(pad(trunc(e.Proc.User, 11), 12)),
		status,
	)
}

func (m *model) rowStatus(e Entry) string {
	switch m.mode {
	case ModeKill:
		r, ok := m.results[e.Proc.PID]
		if !ok {
			return hintStyle.Render(signalLabel(m.force) + "…")
		}
		if r.err != nil {
			return errorStyle.Render("failed")
		}
		return successStyle.Render("killed")
	default:
		return dimStyle.Render("listening")
	}
}

func (m *model) footerMulti() string {
	switch m.mode {
	case ModeShow:
		return hintStyle.Render("append `kill` to terminate, add `-f` to SIGKILL")
	case ModeList:
		return hintStyle.Render(fmt.Sprintf("%d listening port(s) — `dahmer <port> kill` to terminate", len(m.entries)))
	case ModeKill:
		killed, failed := 0, 0
		for _, r := range m.results {
			if r.err != nil {
				failed++
			} else {
				killed++
			}
		}
		parts := []string{successStyle.Render(fmt.Sprintf("✓ %d killed", killed))}
		if failed > 0 {
			parts = append(parts, errorStyle.Render(fmt.Sprintf("✗ %d failed", failed)))
		}
		return strings.Join(parts, "  ")
	}
	return ""
}

func signalLabel(force bool) string {
	if force {
		return "sending SIGKILL"
	}
	return "sending SIGTERM"
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
