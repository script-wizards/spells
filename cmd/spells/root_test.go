package main

import (
	"os/exec"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run spells --version: %v", err)
	}
	
	expected := "dev\n"
	actual := string(output)
	if actual != expected {
		t.Errorf("Expected output %q, got %q", expected, actual)
	}
}