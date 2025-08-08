package dice

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestRoll(t *testing.T) {
	tests := []struct {
		name          string
		expr          string
		seed          int64
		expectedTotal int
		expectedRolls []int
		expectError   bool
	}{
		{
			name:          "2d6+1 deterministic",
			expr:          "2d6+1",
			seed:          28,
			expectedTotal: 8,
			expectedRolls: []int{3, 4},
			expectError:   false,
		},
		{
			name:          "1d20",
			expr:          "1d20",
			seed:          42,
			expectedTotal: 6,
			expectedRolls: []int{6},
			expectError:   false,
		},
		{
			name:          "3d4-2",
			expr:          "3d4-2",
			seed:          10,
			expectedTotal: 6,
			expectedRolls: []int{3, 1, 4},
			expectError:   false,
		},
		{
			name:        "invalid expression",
			expr:        "invalid",
			expectError: true,
		},
		{
			name:        "zero dice",
			expr:        "0d6",
			expectError: true,
		},
		{
			name:        "zero sides",
			expr:        "2d0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				_, _, err := Roll(tt.expr, rand.New(rand.NewSource(1)))
				if err == nil {
					t.Errorf("Roll() expected error but got none")
				}
				return
			}

			rng := rand.New(rand.NewSource(tt.seed))
			total, breakdown, err := Roll(tt.expr, rng)

			if err != nil {
				t.Errorf("Roll() unexpected error: %v", err)
				return
			}

			if total != tt.expectedTotal {
				t.Errorf("Roll() total = %v, expected %v", total, tt.expectedTotal)
			}

			if !reflect.DeepEqual(breakdown, tt.expectedRolls) {
				t.Errorf("Roll() breakdown = %v, expected %v", breakdown, tt.expectedRolls)
			}
		})
	}
}
