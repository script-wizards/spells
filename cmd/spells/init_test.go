package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test.db")

	cmd := exec.Command("go", "run", ".", "init", "--path", testDBPath)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed with error: %v\nOutput: %s", err, string(output))
	}

	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
	}

	if _, err := os.Stat(testDBPath); os.IsNotExist(err) {
		t.Fatalf("Database file does not exist: %s", testDBPath)
	}

	expectedConfigPath := filepath.Join(tmpDir, "test.yaml")
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Fatalf("Config file does not exist: %s", expectedConfigPath)
	}
}
