package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ashutosh-swain-git/dahmer/internal/port"
	"github.com/ashutosh-swain-git/dahmer/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dahmer <port> [kill]",
	Short: "Find and kill the process bound to a TCP port",
	Long: `dahmer inspects the process listening on a TCP port.

  dahmer 3000        show the PID and command holding port 3000
  dahmer 3000 kill   terminate that process (SIGTERM)`,
	Args:          cobra.RangeArgs(1, 2),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := strconv.Atoi(args[0])
		if err != nil || p < 1 || p > 65535 {
			return fmt.Errorf("invalid port %q (must be 1-65535)", args[0])
		}

		mode := ui.ModeShow
		if len(args) == 2 {
			if args[1] != "kill" {
				return fmt.Errorf("unknown subcommand %q (expected `kill`)", args[1])
			}
			mode = ui.ModeKill
		}

		proc, err := port.Lookup(p)
		if err != nil {
			return err
		}

		prog := tea.NewProgram(ui.New(mode, p, proc))
		_, err = prog.Run()
		return err
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
