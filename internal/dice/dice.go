package dice

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
)

var diceExpr = regexp.MustCompile(`^(\d+)d(\d+)(?:([+-])(\d+))?$`)

func Roll(expr string, rng *rand.Rand) (total int, breakdown []int, err error) {
	matches := diceExpr.FindStringSubmatch(expr)
	if matches == nil {
		return 0, nil, fmt.Errorf("invalid dice expression: %s", expr)
	}

	numDice, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, nil, fmt.Errorf("invalid number of dice: %s", matches[1])
	}

	diceSides, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, nil, fmt.Errorf("invalid dice sides: %s", matches[2])
	}

	if numDice <= 0 || diceSides <= 0 {
		return 0, nil, fmt.Errorf("dice count and sides must be positive")
	}

	breakdown = make([]int, numDice)
	total = 0

	for i := 0; i < numDice; i++ {
		roll := rng.Intn(diceSides) + 1
		breakdown[i] = roll
		total += roll
	}

	if matches[3] != "" && matches[4] != "" {
		modifier, err := strconv.Atoi(matches[4])
		if err != nil {
			return 0, nil, fmt.Errorf("invalid modifier: %s", matches[4])
		}

		if matches[3] == "+" {
			total += modifier
		} else {
			total -= modifier
		}
	}

	return total, breakdown, nil
}
