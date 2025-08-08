package engine

import (
	"os"
	"sync"
	"testing"

	"github.com/script-wizards/spells/internal/db"
	"github.com/script-wizards/spells/internal/model"
)

func TestConcurrentTurnAdvancement(t *testing.T) {
	// Create a temporary database for testing
	tempDB, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db: %v", err)
	}
	defer os.Remove(tempDB.Name())
	tempDB.Close()

	// Open database connection
	testDB, err := db.Open(tempDB.Name())
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create a test session
	session := &model.Session{CurrentTurn: 0}
	tx, err := testDB.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	err = session.Create(tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create session: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit session creation: %v", err)
	}

	sessionID := session.ID

	// Create engine instances for each goroutine
	engine1 := &Engine{DB: testDB, EventBus: nil}
	engine2 := &Engine{DB: testDB, EventBus: nil}

	// Use WaitGroup to coordinate goroutines
	var wg sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	// Function to collect errors safely
	collectError := func(err error) {
		if err != nil {
			errorsMutex.Lock()
			errors = append(errors, err)
			errorsMutex.Unlock()
		}
	}

	// Start two goroutines that will each advance the turn 1000 times
	wg.Add(2)

	// Goroutine 1
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			err := engine1.Advance(sessionID, 1)
			collectError(err)
		}
	}()

	// Goroutine 2
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			err := engine2.Advance(sessionID, 1)
			collectError(err)
		}
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Check for any errors
	if len(errors) > 0 {
		t.Fatalf("Encountered %d errors during concurrent execution. First error: %v", len(errors), errors[0])
	}

	// Verify the final turn count is exactly 2000
	finalSession, err := model.GetSession(testDB, sessionID)
	if err != nil {
		t.Fatalf("Failed to get final session state: %v", err)
	}

	if finalSession == nil {
		t.Fatal("Session not found after test")
	}

	expectedTurn := int64(2000)
	if finalSession.CurrentTurn != expectedTurn {
		t.Errorf("Expected final turn to be %d, got %d", expectedTurn, finalSession.CurrentTurn)
	}

	t.Logf("Successfully completed concurrent test. Final turn: %d", finalSession.CurrentTurn)
}
