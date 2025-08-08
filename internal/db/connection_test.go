package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_db_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	var exists int
	err = db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&exists)
	if err != nil {
		t.Fatalf("schema_version table does not exist: %v", err)
	}

	if exists != 1 {
		t.Error("schema_version table was not created")
	}

	var migrationExists int
	err = db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&migrationExists)
	if err != nil {
		t.Fatalf("migrations table does not exist: %v", err)
	}

	if migrationExists != 1 {
		t.Error("migrations table was not created")
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = '0001_init.sql'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check migration record: %v", err)
	}

	if count != 1 {
		t.Error("Initial migration was not recorded")
	}
}

func TestRunMigrationsIdempotence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spells_migrate_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database first time: %v", err)
	}
	db1.Close()

	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database second time: %v", err)
	}
	defer db2.Close()

	var migrationCount int
	err = db2.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = '0001_init.sql'").Scan(&migrationCount)
	if err != nil {
		t.Fatalf("Failed to check migration count: %v", err)
	}

	if migrationCount != 1 {
		t.Errorf("Expected 1 migration record, got %d", migrationCount)
	}

	var exists int
	err = db2.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&exists)
	if err != nil {
		t.Fatalf("schema_version table does not exist after second open: %v", err)
	}

	if exists != 1 {
		t.Error("schema_version table was not found after idempotent migration")
	}
}