package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/script-wizards/spells/internal/oracle"
	"github.com/spf13/cobra"
)

var oracleCmd = &cobra.Command{
	Use:   "oracle [text]",
	Short: "Parse and resolve oracle text with choices, tables, and dice",
	Long: `Parse oracle text containing:
- Choices: {option1|option2|option3}  
- Tables: [table_name]
- Dice: 1d4, 2d6, etc.
- Plain text

Returns JSON with the resolved result.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		// Create resolver with empty tables and random seed
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		resolver := oracle.NewResolver(map[string]string{}, rng)

		result, err := resolver.Resolve(input)
		if err != nil {
			return fmt.Errorf("failed to resolve oracle: %w", err)
		}

		// Create JSON output
		output := map[string]interface{}{
			"input":  input,
			"result": result,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		cmd.Println(string(jsonBytes))
		return nil
	},
}
