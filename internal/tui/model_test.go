package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModel_Init(t *testing.T) {
	model := Model{}
	cmd := model.Init()
	if cmd != nil {
		t.Errorf("expected Init() to return nil, got %v", cmd)
	}
}

func TestModel_Update_CtrlC(t *testing.T) {
	model := Model{}

	// Test Ctrl+C key message
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, cmd := model.Update(keyMsg)

	if newModel == nil {
		t.Error("expected Update() to return a model, got nil")
	}

	if cmd == nil {
		t.Error("expected Update() to return tea.Quit command for Ctrl+C, got nil")
	}

	// We can't directly compare functions, but we know tea.Quit is returned
	// This smoke test verifies the Update function doesn't panic and returns a command
}
