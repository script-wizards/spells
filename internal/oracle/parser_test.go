package oracle

import (
	"reflect"
	"testing"
)

func TestParseChoice(t *testing.T) {
	input := "{option1|option2}"
	result, err := Parse("", []byte(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	parts := result.([]interface{})
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}

	choice, ok := parts[0].(Choice)
	if !ok {
		t.Fatalf("Expected Choice, got %T", parts[0])
	}

	expected := []string{"option1", "option2"}
	if !reflect.DeepEqual(choice.Options, expected) {
		t.Fatalf("Expected %v, got %v", expected, choice.Options)
	}
}

func TestParseTable(t *testing.T) {
	input := "[table]"
	result, err := Parse("", []byte(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	parts := result.([]interface{})
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}

	table, ok := parts[0].(Table)
	if !ok {
		t.Fatalf("Expected Table, got %T", parts[0])
	}

	if table.Name != "table" {
		t.Fatalf("Expected 'table', got %s", table.Name)
	}
}

func TestParseDice(t *testing.T) {
	input := "1d4"
	result, err := Parse("", []byte(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	parts := result.([]interface{})
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}

	dice, ok := parts[0].(Dice)
	if !ok {
		t.Fatalf("Expected Dice, got %T", parts[0])
	}

	if dice.Count != 1 || dice.Sides != 4 {
		t.Fatalf("Expected 1d4, got %dd%d", dice.Count, dice.Sides)
	}
}

func TestParseText(t *testing.T) {
	input := "hello world"
	result, err := Parse("", []byte(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	parts := result.([]interface{})
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}

	text, ok := parts[0].(Text)
	if !ok {
		t.Fatalf("Expected Text, got %T", parts[0])
	}

	if text.Value != "hello world" {
		t.Fatalf("Expected 'hello world', got %s", text.Value)
	}
}

func TestParseNested(t *testing.T) {
	input := "You encounter {1d4|2d6} [creature] in the {forest|cave}"
	result, err := Parse("", []byte(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	parts := result.([]interface{})
	if len(parts) < 5 {
		t.Fatalf("Expected at least 5 parts, got %d", len(parts))
	}

	// Just verify we get the expected types in sequence
	foundChoice := false
	foundTable := false
	foundText := false

	for _, part := range parts {
		switch part.(type) {
		case Choice:
			foundChoice = true
		case Table:
			foundTable = true
		case Text:
			foundText = true
		}
	}

	if !foundChoice {
		t.Fatal("Expected to find Choice")
	}
	if !foundTable {
		t.Fatal("Expected to find Table")
	}
	if !foundText {
		t.Fatal("Expected to find Text")
	}
}
