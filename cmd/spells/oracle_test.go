package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestOracleCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:     "simple text",
			args:     []string{"hello world"},
			wantErr:  false,
			contains: []string{"hello world", "input", "result"},
		},
		{
			name:     "dice roll",
			args:     []string{"1d4"},
			wantErr:  false,
			contains: []string{"1d4", "input", "result"},
		},
		{
			name:     "choice",
			args:     []string{"{option1|option2}"},
			wantErr:  false,
			contains: []string{"input", "result"},
		},
		{
			name:     "table reference",
			args:     []string{"[creature]"},
			wantErr:  false,
			contains: []string{"[creature]", "input", "result"},
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many arguments",
			args:    []string{"arg1", "arg2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the oracle command
			cmd := &cobra.Command{
				Use:   oracleCmd.Use,
				Short: oracleCmd.Short,
				Args:  oracleCmd.Args,
				RunE:  oracleCmd.RunE,
			}

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Set arguments
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check error expectation
			if tt.wantErr && err == nil {
				t.Errorf("Expected error, but got none")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// If we expected an error, we're done
			if tt.wantErr {
				return
			}

			// Check output
			output := buf.String()

			// Verify it's valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("Output is not valid JSON: %v\nOutput: %s", err, output)
				return
			}

			// Check required fields
			if _, exists := result["input"]; !exists {
				t.Errorf("Output missing 'input' field")
			}
			if _, exists := result["result"]; !exists {
				t.Errorf("Output missing 'result' field")
			}

			// Check contains
			for _, contain := range tt.contains {
				if !strings.Contains(output, contain) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", contain, output)
				}
			}
		})
	}
}

func TestOracleCommandComplexInput(t *testing.T) {
	// Create a new command instance
	cmd := &cobra.Command{
		Use:   "oracle [text]",
		Short: "Parse and resolve oracle text with choices, tables, and dice",
		Args:  cobra.ExactArgs(1),
		RunE:  oracleCmd.RunE,
	}

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set arguments with complex input
	complexInput := "You encounter 1d4 {goblins|orcs} in the [location]"
	cmd.SetArgs([]string{complexInput})

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check output
	output := buf.String()

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Check that input is preserved
	if result["input"] != complexInput {
		t.Errorf("Expected input to be %q, got %q", complexInput, result["input"])
	}

	// Check that result is a string and not empty
	resultStr, ok := result["result"].(string)
	if !ok {
		t.Errorf("Expected result to be a string, got %T", result["result"])
	}
	if resultStr == "" {
		t.Errorf("Expected result to be non-empty")
	}

	// Verify some transformation occurred (dice should be resolved to numbers)
	if strings.Contains(resultStr, "1d4") {
		t.Errorf("Expected dice to be resolved, but found '1d4' in result: %s", resultStr)
	}
}
