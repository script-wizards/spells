package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/script-wizards/spells/internal/model"
	_ "modernc.org/sqlite"
)

func createTestDB(t *testing.T) *sqlx.DB {
	database, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	createTableSQL := `
	CREATE TABLE npcs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		location TEXT,
		status TEXT DEFAULT 'neutral',
		motivation TEXT,
		secrets TEXT,
		tags TEXT,
		last_mentioned TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = database.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	return database
}

func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedFM   *FrontMatter
		expectedBody string
		expectError  bool
	}{
		{
			name: "valid NPC markdown",
			content: `---
type: npc
name: Thorg
tags: [orc, warlord]
---

Thorg is a mighty orc warlord who commands respect through fear.`,
			expectedFM: &FrontMatter{
				Type: "npc",
				Name: "Thorg",
				Tags: []string{"orc", "warlord"},
			},
			expectedBody: "Thorg is a mighty orc warlord who commands respect through fear.",
		},
		{
			name: "no front matter",
			content: `Just some regular markdown content.

Nothing special here.`,
			expectedFM: nil,
			expectedBody: `Just some regular markdown content.

Nothing special here.`,
		},
		{
			name: "non-npc type",
			content: `---
type: location
name: Tavern
---

A cozy tavern.`,
			expectedFM: &FrontMatter{
				Type: "location",
				Name: "Tavern",
				Tags: nil,
			},
			expectedBody: "A cozy tavern.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			database := createTestDB(t)
			defer database.Close()

			watcher, err := NewWatcher(database)
			if err != nil {
				t.Fatalf("failed to create watcher: %v", err)
			}
			defer watcher.Stop()

			fm, body, err := watcher.parseFrontMatter(tmpFile)

			if tt.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectedFM == nil && fm != nil {
				t.Fatalf("expected nil front matter, got %+v", fm)
			}
			if tt.expectedFM != nil && fm == nil {
				t.Fatal("expected front matter, got nil")
			}

			if tt.expectedFM != nil {
				if fm.Type != tt.expectedFM.Type {
					t.Errorf("expected type %q, got %q", tt.expectedFM.Type, fm.Type)
				}
				if fm.Name != tt.expectedFM.Name {
					t.Errorf("expected name %q, got %q", tt.expectedFM.Name, fm.Name)
				}
				if len(fm.Tags) != len(tt.expectedFM.Tags) {
					t.Errorf("expected %d tags, got %d", len(tt.expectedFM.Tags), len(fm.Tags))
				}
				for i, tag := range tt.expectedFM.Tags {
					if i < len(fm.Tags) && fm.Tags[i] != tag {
						t.Errorf("expected tag %q at index %d, got %q", tag, i, fm.Tags[i])
					}
				}
			}

			if body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestConvertToNPC(t *testing.T) {
	tests := []struct {
		name     string
		fm       *FrontMatter
		body     string
		expected *model.NPC
	}{
		{
			name: "full NPC conversion",
			fm: &FrontMatter{
				Type: "npc",
				Name: "Thorg",
				Tags: []string{"orc", "warlord"},
			},
			body: "A mighty orc warlord.",
			expected: &model.NPC{
				Name:        "Thorg",
				Description: stringPtr("A mighty orc warlord."),
				Status:      "neutral",
				Tags:        stringPtr(`["orc","warlord"]`),
			},
		},
		{
			name: "NPC without description",
			fm: &FrontMatter{
				Type: "npc",
				Name: "Silent Bob",
				Tags: []string{"human", "mute"},
			},
			body: "",
			expected: &model.NPC{
				Name:        "Silent Bob",
				Description: nil,
				Status:      "neutral",
				Tags:        stringPtr(`["human","mute"]`),
			},
		},
		{
			name: "NPC without tags",
			fm: &FrontMatter{
				Type: "npc",
				Name: "John Doe",
				Tags: nil,
			},
			body: "An ordinary person.",
			expected: &model.NPC{
				Name:        "John Doe",
				Description: stringPtr("An ordinary person."),
				Status:      "neutral",
				Tags:        nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := createTestDB(t)
			defer database.Close()

			watcher, err := NewWatcher(database)
			if err != nil {
				t.Fatalf("failed to create watcher: %v", err)
			}
			defer watcher.Stop()

			result := watcher.convertToNPC(tt.fm, tt.body)

			if result.Name != tt.expected.Name {
				t.Errorf("expected name %q, got %q", tt.expected.Name, result.Name)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("expected status %q, got %q", tt.expected.Status, result.Status)
			}

			if tt.expected.Description == nil && result.Description != nil {
				t.Errorf("expected nil description, got %q", *result.Description)
			}
			if tt.expected.Description != nil && result.Description == nil {
				t.Errorf("expected description %q, got nil", *tt.expected.Description)
			}
			if tt.expected.Description != nil && result.Description != nil &&
				*result.Description != *tt.expected.Description {
				t.Errorf("expected description %q, got %q", *tt.expected.Description, *result.Description)
			}

			if tt.expected.Tags == nil && result.Tags != nil {
				t.Errorf("expected nil tags, got %q", *result.Tags)
			}
			if tt.expected.Tags != nil && result.Tags == nil {
				t.Errorf("expected tags %q, got nil", *tt.expected.Tags)
			}
			if tt.expected.Tags != nil && result.Tags != nil &&
				*result.Tags != *tt.expected.Tags {
				t.Errorf("expected tags %q, got %q", *tt.expected.Tags, *result.Tags)
			}
		})
	}
}

func TestUpsertNPC(t *testing.T) {
	database := createTestDB(t)
	defer database.Close()

	watcher, err := NewWatcher(database)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	npc := &model.NPC{
		Name:        "Test NPC",
		Description: stringPtr("A test character"),
		Status:      "neutral",
		Tags:        stringPtr(`["test","character"]`),
	}

	err = watcher.upsertNPC(npc)
	if err != nil {
		t.Fatalf("failed to upsert NPC: %v", err)
	}

	if npc.ID == 0 {
		t.Error("expected NPC to have an ID after upsert")
	}

	var stored model.NPC
	err = database.Get(&stored, "SELECT id, name, description, status, tags FROM npcs WHERE name = ?", "Test NPC")
	if err != nil {
		t.Fatalf("failed to retrieve stored NPC: %v", err)
	}

	if stored.Name != npc.Name {
		t.Errorf("expected name %q, got %q", npc.Name, stored.Name)
	}
	if stored.Description == nil || *stored.Description != *npc.Description {
		t.Errorf("expected description %q, got %v", *npc.Description, stored.Description)
	}

	updatedNPC := &model.NPC{
		Name:        "Test NPC",
		Description: stringPtr("An updated test character"),
		Status:      "neutral",
		Tags:        stringPtr(`["test","character","updated"]`),
	}

	err = watcher.upsertNPC(updatedNPC)
	if err != nil {
		t.Fatalf("failed to update NPC: %v", err)
	}

	var updatedStored model.NPC
	err = database.Get(&updatedStored, "SELECT id, name, description, status, tags FROM npcs WHERE name = ?", "Test NPC")
	if err != nil {
		t.Fatalf("failed to retrieve updated NPC: %v", err)
	}

	if updatedStored.ID != stored.ID {
		t.Error("expected NPC ID to remain the same after update")
	}
	if updatedStored.Description == nil || *updatedStored.Description != "An updated test character" {
		t.Errorf("expected updated description, got %v", updatedStored.Description)
	}

	var updatedTags []string
	if updatedStored.Tags != nil {
		json.Unmarshal([]byte(*updatedStored.Tags), &updatedTags)
	}
	expectedTags := []string{"test", "character", "updated"}
	if len(updatedTags) != len(expectedTags) {
		t.Errorf("expected %d tags, got %d", len(expectedTags), len(updatedTags))
	}
}

func TestFileWatcher(t *testing.T) {
	database := createTestDB(t)
	defer database.Close()

	watcher, err := NewWatcher(database)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	tmpDir := t.TempDir()

	err = watcher.Start(tmpDir)
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	testFile := filepath.Join(tmpDir, "test_npc.md")
	content := `---
type: npc
name: Watcher Test NPC
tags: [test, watcher]
---

This NPC was created by the file watcher test.`

	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	var npc model.NPC
	err = database.Get(&npc, "SELECT id, name, description, tags FROM npcs WHERE name = ?", "Watcher Test NPC")
	if err != nil {
		t.Fatalf("failed to find NPC created by watcher: %v", err)
	}

	if npc.Name != "Watcher Test NPC" {
		t.Errorf("expected name %q, got %q", "Watcher Test NPC", npc.Name)
	}

	if npc.Description == nil {
		t.Fatal("expected description to be set")
	}
	expectedDesc := "This NPC was created by the file watcher test."
	if *npc.Description != expectedDesc {
		t.Errorf("expected description %q, got %q", expectedDesc, *npc.Description)
	}

	var tags []string
	if npc.Tags != nil {
		json.Unmarshal([]byte(*npc.Tags), &tags)
	}
	expectedTags := []string{"test", "watcher"}
	if len(tags) != len(expectedTags) {
		t.Errorf("expected %d tags, got %d", len(expectedTags), len(tags))
	}
	for i, tag := range expectedTags {
		if i < len(tags) && tags[i] != tag {
			t.Errorf("expected tag %q at index %d, got %q", tag, i, tags[i])
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
