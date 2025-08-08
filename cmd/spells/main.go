package main

import (
	"fmt"
	"os"

	"github.com/script-wizards/spells/internal/version"
	"github.com/spf13/cobra"
)

var (
	config      string
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
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(oracleCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
