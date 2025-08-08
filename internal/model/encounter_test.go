package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/script-wizards/spells/internal/db"
)

func TestAddCombatantAndListActiveBySort(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_encounter_test")
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

	tx, err := database.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	session := &Session{CurrentTurn: 0}
	err = session.Create(tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create session: %v", err)
	}

	encounter := &Encounter{
		SessionID:   session.ID,
		Name:        stringPtr("Test Encounter"),
		Description: stringPtr("A test encounter"),
		IsActive:    true,
	}
	err = CreateEncounter(tx, encounter)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create encounter: %v", err)
	}

	npc := &NPC{
		Name:   "Goblin",
		Status: "hostile",
	}
	err = CreateNPC(tx, npc)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create NPC: %v", err)
	}

	// Add combatants with different initiative values
	combatants := []struct {
		name       *string
		npcID      *int64
		initiative int
		hpCurrent  *int
		hpMax      *int
	}{
		{stringPtr("Alice"), nil, 20, intPtr(25), intPtr(25)},   // PC, highest initiative
		{nil, &npc.ID, 15, intPtr(8), intPtr(8)},                // NPC, middle initiative
		{stringPtr("Bob"), nil, 10, intPtr(30), intPtr(30)},     // PC, lowest initiative
		{stringPtr("Charlie"), nil, 18, intPtr(22), intPtr(22)}, // PC, second highest
	}

	for _, c := range combatants {
		_, err := AddCombatant(tx, encounter.ID, c.npcID, c.name, c.initiative, c.hpCurrent, c.hpMax)
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to add combatant: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Test ListActiveBySort - should be ordered by initiative DESC, then by ID ASC
	combatantList, err := ListActiveBySort(database, encounter.ID)
	if err != nil {
		t.Fatalf("Failed to list active combatants: %v", err)
	}

	if len(combatantList) != 4 {
		t.Errorf("Expected 4 combatants, got %d", len(combatantList))
	}

	// Expected order: Alice (20), Charlie (18), Goblin (15), Bob (10)
	expectedOrder := []struct {
		name       string
		initiative int
		isNPC      bool
	}{
		{"Alice", 20, false},
		{"Charlie", 18, false},
		{"Goblin", 15, true},
		{"Bob", 10, false},
	}

	for i, expected := range expectedOrder {
		if i >= len(combatantList) {
			t.Errorf("Missing combatant at index %d", i)
			continue
		}

		combatant := combatantList[i]
		if combatant.Name != expected.name {
			t.Errorf("Expected combatant %d name to be %s, got %s", i, expected.name, combatant.Name)
		}
		if combatant.Initiative != expected.initiative {
			t.Errorf("Expected combatant %d initiative to be %d, got %d", i, expected.initiative, combatant.Initiative)
		}
		if combatant.IsNPC != expected.isNPC {
			t.Errorf("Expected combatant %d IsNPC to be %v, got %v", i, expected.isNPC, combatant.IsNPC)
		}

		// Test HP values
		if combatant.HPCurrent == nil || combatant.HPMax == nil {
			t.Errorf("Expected combatant %d to have HP values, got nil", i)
		}
	}
}

func TestGetActiveEncounter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_encounter_test")
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

	tx, err := database.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	session := &Session{CurrentTurn: 0}
	err = session.Create(tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create inactive encounter
	inactiveEncounter := &Encounter{
		SessionID: session.ID,
		Name:      stringPtr("Inactive Encounter"),
		IsActive:  false,
	}
	err = CreateEncounter(tx, inactiveEncounter)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create inactive encounter: %v", err)
	}

	// Create active encounter
	activeEncounter := &Encounter{
		SessionID: session.ID,
		Name:      stringPtr("Active Encounter"),
		IsActive:  true,
	}
	err = CreateEncounter(tx, activeEncounter)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create active encounter: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Test getting active encounter
	retrieved, err := GetActiveEncounter(database, session.ID)
	if err != nil {
		t.Fatalf("Failed to get active encounter: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected to find an active encounter, got nil")
	}

	if retrieved.ID != activeEncounter.ID {
		t.Errorf("Expected active encounter ID %d, got %d", activeEncounter.ID, retrieved.ID)
	}

	if retrieved.Name == nil || *retrieved.Name != "Active Encounter" {
		name := "nil"
		if retrieved.Name != nil {
			name = *retrieved.Name
		}
		t.Errorf("Expected active encounter name 'Active Encounter', got %s", name)
	}

	if !retrieved.IsActive {
		t.Error("Expected retrieved encounter to be active")
	}
}

func TestGetActiveEncounterNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_encounter_test")
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

	// Test getting active encounter for non-existent session
	encounter, err := GetActiveEncounter(database, 999)
	if err != nil {
		t.Fatalf("Unexpected error when getting non-existent active encounter: %v", err)
	}

	if encounter != nil {
		t.Error("Expected nil encounter for non-existent session")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
