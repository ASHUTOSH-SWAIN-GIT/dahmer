package ui

import (
	"fmt"

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
)

type model struct {
	mode    Mode
	port    int
	proc    *port.Process
	killed  bool
	killErr error
	done    bool
}

func New(mode Mode, p int, proc *port.Process) tea.Model {
	return &model{mode: mode, port: p, proc: proc}
}

type killMsg struct{ err error }

func (m *model) killCmd() tea.Msg {
	if m.proc == nil {
		return killMsg{err: nil}
	}
	return killMsg{err: port.Kill(m.proc.PID)}
}

func (m *model) Init() tea.Cmd {
	if m.mode == ModeKill && m.proc != nil {
		return m.killCmd
	}
	m.done = true
	return tea.Quit
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case killMsg:
		m.killErr = msg.err
		m.killed = msg.err == nil
		m.done = true
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() string {
	title := titleStyle.Render(fmt.Sprintf("dahmer :: port %d", m.port))

	if m.proc == nil {
		body := emptyStyle.Render(fmt.Sprintf("No process is listening on port %d.", m.port))
		return boxStyle.Render(title + "\n\n" + body)
	}

	row := func(k, v string) string {
		return labelStyle.Render(fmt.Sprintf("%-10s", k)) + valueStyle.Render(v)
	}

	info := lipgloss.JoinVertical(lipgloss.Left,
		row("PID", fmt.Sprintf("%d", m.proc.PID)),
		row("Command", m.proc.Command),
		row("User", m.proc.User),
	)

	var footer string
	switch m.mode {
	case ModeShow:
		footer = hintStyle.Render(fmt.Sprintf("run `dahmer %d kill` to terminate it", m.port))
	case ModeKill:
		if !m.done {
			footer = hintStyle.Render("sending SIGTERM…")
		} else if m.killErr != nil {
			footer = errorStyle.Render("failed: " + m.killErr.Error())
		} else {
			footer = successStyle.Render(fmt.Sprintf("✓ killed PID %d", m.proc.PID))
		}
	}

	return boxStyle.Render(title + "\n\n" + info + "\n\n" + footer)
}
