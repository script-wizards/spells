package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/script-wizards/spells/internal/db"
	"github.com/script-wizards/spells/internal/model"
)

func TestEngine_Advance(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "engine_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer database.Close()

	eventBus := NewEventBus()
	engine := &Engine{DB: database, EventBus: eventBus}

	var capturedEvents []Event
	eventBus.Subscribe("TurnAdvanced", func(event Event) {
		capturedEvents = append(capturedEvents, event)
	})

	tx, err := database.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	session := &model.Session{CurrentTurn: 1}
	err = session.Create(tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create session: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	err = engine.Advance(session.ID, 2)
	if err != nil {
		t.Fatalf("Failed to advance turn: %v", err)
	}

	err = engine.Advance(session.ID, 3)
	if err != nil {
		t.Fatalf("Failed to advance turn second time: %v", err)
	}

	updatedSession, err := model.GetSession(database, session.ID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	expectedTurn := int64(6) // 1 + 2 + 3
	if updatedSession.CurrentTurn != expectedTurn {
		t.Errorf("Expected turn %d, got %d", expectedTurn, updatedSession.CurrentTurn)
	}

	if len(capturedEvents) != 2 {
		t.Errorf("Expected 2 events, got %d", len(capturedEvents))
	}

	if len(capturedEvents) >= 1 {
		event1 := capturedEvents[0].(TurnAdvanced)
		if event1.SessionID != session.ID {
			t.Errorf("Event 1: Expected session ID %d, got %d", session.ID, event1.SessionID)
		}
		if event1.OldTurn != 1 {
			t.Errorf("Event 1: Expected old turn 1, got %d", event1.OldTurn)
		}
		if event1.NewTurn != 3 {
			t.Errorf("Event 1: Expected new turn 3, got %d", event1.NewTurn)
		}
		if event1.Delta != 2 {
			t.Errorf("Event 1: Expected delta 2, got %d", event1.Delta)
		}
	}

	if len(capturedEvents) >= 2 {
		event2 := capturedEvents[1].(TurnAdvanced)
		if event2.SessionID != session.ID {
			t.Errorf("Event 2: Expected session ID %d, got %d", session.ID, event2.SessionID)
		}
		if event2.OldTurn != 3 {
			t.Errorf("Event 2: Expected old turn 3, got %d", event2.OldTurn)
		}
		if event2.NewTurn != 6 {
			t.Errorf("Event 2: Expected new turn 6, got %d", event2.NewTurn)
		}
		if event2.Delta != 3 {
			t.Errorf("Event 2: Expected delta 3, got %d", event2.Delta)
		}
	}
}
