package oracle

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/script-wizards/spells/internal/dice"
)

type Resolver struct {
	tables map[string]string
	rng    *rand.Rand
}

func NewResolver(tables map[string]string, rng *rand.Rand) *Resolver {
	return &Resolver{
		tables: tables,
		rng:    rng,
	}
}

func (r *Resolver) Resolve(input string) (string, error) {
	parsed, err := Parse("", []byte(input))
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	parts, ok := parsed.([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected parse result type")
	}

	var result strings.Builder
	for _, part := range parts {
		resolved, err := r.resolvePart(part)
		if err != nil {
			return "", err
		}
		result.WriteString(resolved)
	}

	return result.String(), nil
}

func (r *Resolver) resolvePart(part interface{}) (string, error) {
	switch p := part.(type) {
	case Choice:
		if len(p.Options) == 0 {
			return "", fmt.Errorf("empty choice")
		}
		idx := r.rng.Intn(len(p.Options))
		return strings.TrimSpace(p.Options[idx]), nil

	case Table:
		if r.tables == nil {
			return fmt.Sprintf("[%s]", p.Name), nil
		}
		value, exists := r.tables[p.Name]
		if !exists {
			return fmt.Sprintf("[%s]", p.Name), nil
		}
		return r.Resolve(value)

	case Dice:
		expr := fmt.Sprintf("%dd%d", p.Count, p.Sides)
		total, _, err := dice.Roll(expr, r.rng)
		if err != nil {
			return "", fmt.Errorf("dice roll error: %w", err)
		}
		return fmt.Sprintf("%d", total), nil

	case Text:
		return p.Value, nil

	default:
		return "", fmt.Errorf("unknown part type: %T", part)
	}
}
