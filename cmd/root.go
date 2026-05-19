package cmd

import (
	"fmt"
	"os"

	"github.com/ashutosh-swain-git/dahmer/internal/port"
	"github.com/ashutosh-swain-git/dahmer/internal/ports"
	"github.com/ashutosh-swain-git/dahmer/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var force bool

var rootCmd = &cobra.Command{
	Use:   "dahmer <port|range>... [kill]",
	Short: "Find and kill the process bound to a TCP port",
	Long: `dahmer inspects and kills processes listening on TCP ports.

  dahmer 3000              show what holds port 3000
  dahmer 3000 3001 8080    show multiple ports
  dahmer 3000-3010         show a range
  dahmer 3000 kill         SIGTERM the process on 3000
  dahmer 3000-3010 kill    SIGTERM every process in the range
  dahmer 3000 kill -f      SIGKILL instead
  dahmer ls                list every listening TCP port`,
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runRoot,
}

var lsCmd = &cobra.Command{
	Use:           "ls",
	Short:         "List every process listening on a TCP port",
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		procs, err := port.ListAll()
		if err != nil {
			return err
		}
		_, err = tea.NewProgram(ui.NewList(procs)).Run()
		return err
	},
}

func runRoot(cmd *cobra.Command, args []string) error {
	mode := ui.ModeShow
	if args[len(args)-1] == "kill" {
		mode = ui.ModeKill
		args = args[:len(args)-1]
		if len(args) == 0 {
			return fmt.Errorf("`kill` requires at least one port")
		}
	}

	parsed, err := ports.Parse(args)
	if err != nil {
		return err
	}

	entries := make([]ui.Entry, 0, len(parsed))
	for _, p := range parsed {
		proc, err := port.Lookup(p)
		if err != nil {
			return err
		}
		entries = append(entries, ui.Entry{Port: p, Proc: proc})
	}

	var m tea.Model
	if mode == ui.ModeKill {
		m = ui.NewKill(entries, force)
	} else {
		m = ui.NewShow(entries)
	}
	_, err = tea.NewProgram(m).Run()
	return err
}

func init() {
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "use SIGKILL instead of SIGTERM")
	rootCmd.AddCommand(lsCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
