package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/script-wizards/spells/internal/db"
	"github.com/script-wizards/spells/internal/search"
)

func TestNPCCRUDOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_npc_test")
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

	// Test CreateNPC
	npc := &NPC{
		Name:        "Gareth the Merchant",
		Description: stringPtr("A suspicious merchant who knows too much"),
		Location:    stringPtr("Market Square"),
		Status:      "neutral",
		Motivation:  stringPtr("Hide his involvement with the cult"),
		Secrets:     stringPtr("Knows about the secret passage under the temple"),
		Tags:        stringPtr(`["merchant", "informant", "cult"]`),
	}

	tx, err := database.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	err = CreateNPC(tx, npc)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create NPC: %v", err)
	}

	if npc.ID == 0 {
		tx.Rollback()
		t.Error("NPC ID was not set after creation")
	}

	if npc.CreatedAt.IsZero() {
		tx.Rollback()
		t.Error("NPC CreatedAt was not set after creation")
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Test GetNPC
	retrievedNPC, err := GetNPC(database, npc.ID)
	if err != nil {
		t.Fatalf("Failed to get NPC: %v", err)
	}

	if retrievedNPC == nil {
		t.Fatal("Retrieved NPC is nil")
	}

	if retrievedNPC.ID != npc.ID {
		t.Errorf("Expected NPC ID %d, got %d", npc.ID, retrievedNPC.ID)
	}

	if retrievedNPC.Name != npc.Name {
		t.Errorf("Expected NPC Name %s, got %s", npc.Name, retrievedNPC.Name)
	}

	if retrievedNPC.Description == nil || *retrievedNPC.Description != *npc.Description {
		t.Errorf("Expected NPC Description %s, got %v", *npc.Description, retrievedNPC.Description)
	}

	if retrievedNPC.Status != npc.Status {
		t.Errorf("Expected NPC Status %s, got %s", npc.Status, retrievedNPC.Status)
	}
}

func TestGetNPCNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_npc_test")
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

	npc, err := GetNPC(database, 999)
	if err != nil {
		t.Fatalf("Unexpected error when getting non-existent NPC: %v", err)
	}

	if npc != nil {
		t.Error("Expected nil NPC for non-existent ID")
	}
}

func TestSearchNPC(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_npc_search_test")
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

	// Create test NPCs
	npcs := []*NPC{
		{Name: "Kobold Warrior", Status: "hostile"},
		{Name: "Goblin Shaman", Status: "hostile"},
		{Name: "Hobgoblin Captain", Status: "neutral"},
	}

	tx, err := database.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	for _, npc := range npcs {
		err = CreateNPC(tx, npc)
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to create NPC %s: %v", npc.Name, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build search index
	names, err := GetAllNPCNames(database)
	if err != nil {
		t.Fatalf("Failed to get NPC names: %v", err)
	}
	idx := search.BuildIndex(names)

	// Test search functionality
	results, err := SearchNPC(database, idx, "kobo", 5)
	if err != nil {
		t.Fatalf("Failed to search NPCs: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result for 'kobo' query")
	}

	// Check that Kobold Warrior is in the results (order may vary due to scoring)
	found := false
	for _, result := range results {
		if result.Name == "Kobold Warrior" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Kobold Warrior' to be found in results for 'kobo' query")
	}

	// Test search with partial match
	results, err = SearchNPC(database, idx, "gob", 5)
	if err != nil {
		t.Fatalf("Failed to search NPCs: %v", err)
	}

	found = false
	for _, result := range results {
		if result.Name == "Goblin Shaman" || result.Name == "Hobgoblin Captain" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find Goblin Shaman or Hobgoblin Captain in 'gob' search results")
	}
}

func TestGetAllNPCNames(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_npc_names_test")
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

	// Initially should be empty
	names, err := GetAllNPCNames(database)
	if err != nil {
		t.Fatalf("Failed to get NPC names: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("Expected 0 names, got %d", len(names))
	}

	// Create a test NPC
	npc := &NPC{Name: "Test NPC", Status: "neutral"}

	tx, err := database.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	err = CreateNPC(tx, npc)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create NPC: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Should now have one name
	names, err = GetAllNPCNames(database)
	if err != nil {
		t.Fatalf("Failed to get NPC names: %v", err)
	}

	if len(names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(names))
	}

	if names[0] != "Test NPC" {
		t.Errorf("Expected 'Test NPC', got '%s'", names[0])
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
