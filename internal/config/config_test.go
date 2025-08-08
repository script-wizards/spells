package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.TorchDuration != 10 {
		t.Errorf("expected TorchDuration to be 10, got %d", config.TorchDuration)
	}
}

func TestLoad_NoFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.TorchDuration != 10 {
		t.Errorf("expected default TorchDuration of 10, got %d", config.TorchDuration)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}

func TestLoad_WithFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := "torch_duration_turns: 6\n"
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.TorchDuration != 6 {
		t.Errorf("expected TorchDuration to be 6, got %d", config.TorchDuration)
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	tempDir, err := os.MkdirTemp("", "spells-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("XDG_CONFIG_HOME", tempDir)

	config, err := Load("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.TorchDuration != 10 {
		t.Errorf("expected default TorchDuration of 10, got %d", config.TorchDuration)
	}

	expectedPath := filepath.Join(tempDir, "spells", "config.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("expected config file to be created in XDG_CONFIG_HOME/spells/")
	}
}
