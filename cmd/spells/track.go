package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/script-wizards/spells/internal/db"
	"github.com/script-wizards/spells/internal/engine"
	"github.com/script-wizards/spells/internal/tui"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Start the interactive spell tracking TUI",
	Long:  "Launch the terminal user interface for tracking spells and sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		sessionID, _ := cmd.Flags().GetInt64("session-id")

		// Open the database
		database, err := db.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Create engine
		eng := &engine.Engine{
			DB: database,
		}

		// Create TUI model
		model, err := tui.NewModel(eng, sessionID)
		if err != nil {
			return fmt.Errorf("failed to create TUI model: %w", err)
		}

		program := tea.NewProgram(model)

		if _, err := program.Run(); err != nil {
			return fmt.Errorf("failed to start TUI: %w", err)
		}

		return nil
	},
}

func init() {
	trackCmd.Flags().String("path", "./campaign.db", "path to the database file")
	trackCmd.Flags().Int64("session-id", 1, "session ID to track")
}
