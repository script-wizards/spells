package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/script-wizards/spells/internal/version"
)

var (
	config     string
	showVersion bool
)

var rootCmd = &cobra.Command{
	Use:   "spells",
	Short: "A spell management tool",
	Long:  "A command-line tool for managing and casting spells",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Version)
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.Flags().StringVar(&config, "config", "", "config file")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "show version")
	
	rootCmd.AddCommand(initCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}