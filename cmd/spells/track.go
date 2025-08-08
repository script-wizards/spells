package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/script-wizards/spells/internal/db"
	"github.com/script-wizards/spells/internal/tui"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Start the interactive spell tracking TUI",
	Long:  "Launch the terminal user interface for tracking spells and sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")

		// Open the database
		database, err := db.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Create and start the TUI
		model := tui.Model{}
		program := tea.NewProgram(model)

		if _, err := program.Run(); err != nil {
			return fmt.Errorf("failed to start TUI: %w", err)
		}

		return nil
	},
}

func init() {
	trackCmd.Flags().String("path", "./campaign.db", "path to the database file")
}
