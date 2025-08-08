package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	configpkg "github.com/script-wizards/spells/internal/config"
	"github.com/script-wizards/spells/internal/db"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new spells campaign database",
	Long:  "Initialize a new spells campaign database with default configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		
		// Open/create the database
		database, err := db.Open(path)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer database.Close()

		// Create default config YAML alongside the database
		configPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".yaml"
		defaultConfig := configpkg.DefaultConfig()
		if err := configpkg.Save(defaultConfig, configPath); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}

		fmt.Println("initialized")
		return nil
	},
}

func init() {
	initCmd.Flags().String("path", "./campaign.db", "path to the database file")
}