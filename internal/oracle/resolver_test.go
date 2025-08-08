package oracle

import (
	"math/rand"
	"strings"
	"testing"
)

// Create a deterministic RNG for testing
func newTestRng(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func TestResolveChoice(t *testing.T) {
	rng := newTestRng(1)
	resolver := NewResolver(nil, rng)

	result, err := resolver.Resolve("{option1|option2}")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	if result != "option1" && result != "option2" {
		t.Fatalf("Expected 'option1' or 'option2', got %s", result)
	}
}

func TestResolveTable(t *testing.T) {
	tables := map[string]string{
		"creature": "orc",
		"weapon":   "{sword|axe}",
	}
	rng := newTestRng(1)
	resolver := NewResolver(tables, rng)

	result, err := resolver.Resolve("[creature]")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	if result != "orc" {
		t.Fatalf("Expected 'orc', got %s", result)
	}
}

func TestResolveTableWithChoice(t *testing.T) {
	tables := map[string]string{
		"weapon": "{sword|axe}",
	}
	rng := newTestRng(2)
	resolver := NewResolver(tables, rng)

	result, err := resolver.Resolve("[weapon]")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	if result != "sword" && result != "axe" {
		t.Fatalf("Expected 'sword' or 'axe', got %s", result)
	}
}

func TestResolveDice(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	resolver := NewResolver(nil, rng)

	result, err := resolver.Resolve("1d4")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	// Check that result is a number between 1-4
	if result != "1" && result != "2" && result != "3" && result != "4" {
		t.Fatalf("Expected dice result 1-4, got %s", result)
	}
}

func TestResolveText(t *testing.T) {
	resolver := NewResolver(nil, nil)

	result, err := resolver.Resolve("hello world")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	if result != "hello world" {
		t.Fatalf("Expected 'hello world', got %s", result)
	}
}

func TestResolveNested(t *testing.T) {
	tables := map[string]string{
		"creature": "rat",
	}
	rng := newTestRng(3)
	resolver := NewResolver(tables, rng)

	result, err := resolver.Resolve("You encounter 1d4 [creature]")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	if !strings.Contains(result, "You encounter") {
		t.Fatalf("Expected result to contain 'You encounter', got %s", result)
	}
	if !strings.Contains(result, "rat") {
		t.Fatalf("Expected result to contain 'rat', got %s", result)
	}
}

func TestResolveMissingTable(t *testing.T) {
	resolver := NewResolver(nil, nil)

	result, err := resolver.Resolve("[missing]")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	if result != "[missing]" {
		t.Fatalf("Expected '[missing]', got %s", result)
	}
}
