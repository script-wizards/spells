package search

import (
	"testing"
)

func TestTrigramSearch(t *testing.T) {
	terms := []string{"kobold", "goblin"}
	idx := BuildIndex(terms)

	results := Query(idx, "kobol", 10)

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	if results[0].Value != "kobold" {
		t.Errorf("expected first result to be 'kobold', got '%s'", results[0].Value)
	}

	if results[0].Score <= 0 {
		t.Errorf("expected positive score, got %f", results[0].Score)
	}
}

func TestEmptyQuery(t *testing.T) {
	terms := []string{"kobold", "goblin"}
	idx := BuildIndex(terms)

	results := Query(idx, "", 10)
	if results != nil {
		t.Error("expected nil results for empty query")
	}
}

func TestEmptyIndex(t *testing.T) {
	idx := BuildIndex([]string{})

	results := Query(idx, "test", 10)
	if results != nil {
		t.Error("expected nil results for empty index")
	}
}

func TestLimitResults(t *testing.T) {
	terms := []string{"kobold", "goblin", "hobgoblin", "bugbear"}
	idx := BuildIndex(terms)

	results := Query(idx, "gob", 2)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
