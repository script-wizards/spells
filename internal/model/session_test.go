package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/script-wizards/spells/internal/db"
)

func TestSessionCRUDOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_session_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	session := &Session{
		CurrentTurn: 0,
	}

	tx, err := database.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	err = session.Create(tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == 0 {
		tx.Rollback()
		t.Error("Session ID was not set after creation")
	}

	err = session.AdvanceTurn(tx, 3)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to advance turn: %v", err)
	}

	if session.CurrentTurn != 3 {
		tx.Rollback()
		t.Errorf("Expected CurrentTurn to be 3, got %d", session.CurrentTurn)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	retrievedSession, err := GetSession(database, session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSession == nil {
		t.Fatal("Retrieved session is nil")
	}

	if retrievedSession.ID != session.ID {
		t.Errorf("Expected session ID %d, got %d", session.ID, retrievedSession.ID)
	}

	if retrievedSession.CurrentTurn != 3 {
		t.Errorf("Expected CurrentTurn to be 3, got %d", retrievedSession.CurrentTurn)
	}
}

func TestGetSessionNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_session_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	session, err := GetSession(database, 999)
	if err != nil {
		t.Fatalf("Unexpected error when getting non-existent session: %v", err)
	}

	if session != nil {
		t.Error("Expected nil session for non-existent ID")
	}
}