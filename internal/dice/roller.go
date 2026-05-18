// Package dice provides D&D dice rolling functionality.
package dice

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// Roller handles dice rolling with standard D&D notation.
type Roller struct {
	rng *rand.Rand
}

// RollResult represents the result of a single dice expression.
type RollResult struct {
	Expression  string   `json:"expression"`
	Total       int      `json:"total"`
	Rolls       []DieRoll `json:"rolls"`
	Description string   `json:"description"`
	IsAnimating bool     `json:"is_animating"`
}

// DieRoll represents a single die roll with its result.
type DieRoll struct {
	Dice    string `json:"dice"`
	Result  int    `json:"result"`
	Kept    bool   `json:"kept"`
	Details []int  `json:"details,omitempty"`
}

// AdvantageResult represents advantage/disadvantage rolls.
type AdvantageResult struct {
	Roll1     int  `json:"roll1"`
	Roll2     int  `json:"roll2"`
	Selected  int  `json:"selected"`
	IsAdvantage bool `json:"is_advantage"`
}

// Modifier captures bonus/penalty to the roll.
type Modifier struct {
	Value    int    `json:"value"`
	IsBonus  bool   `json:"is_bonus"`
	Description string `json:"description"`
}

var (
	// Standard dice notation: NdX, NdX+M, NdX-M
	// Examples: 1d20, 2d6+3, 1d8-1, 2d20kh1 (keep highest), 2d20kl1 (keep lowest)
	dicePattern = regexp.MustCompile(`(\d+)d(\d+)(kh\d+|kl\d+|kh|kl)?([+-]\d+)?`)
)

// NewRoller creates a new dice roller with a seeded random source.
func NewRoller() *Roller {
	return &Roller{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ParseRoll parses a dice notation string and returns a structured RollResult.
func (r *Roller) ParseRoll(expression string) (*RollResult, error) {
	expression = strings.TrimSpace(expression)
	matches := dicePattern.FindAllStringSubmatch(expression, -1)
	
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid dice notation: %s", expression)
	}

	result := &RollResult{
		Expression: expression,
		Rolls:      make([]DieRoll, 0),
		IsAnimating: true,
	}

	total := 0

	for _, match := range matches {
		count := 0
		sides := 0
		modifier := 0
		keepHigh := 0
		keepLow := 0
		keep := false

		// Parse match groups
		if n, err := fmt.Sscanf(match[1], "%d", &count); err != nil || n != 1 {
			continue
		}
		if n, err := fmt.Sscanf(match[2], "%d", &sides); err != nil || n != 1 {
			continue
		}
		if len(match) > 4 && match[4] != "" {
			fmt.Sscanf(match[4], "%d", &modifier)
		}

		// Parse keep modifier
		if len(match) > 3 && match[3] != "" {
			keep = true
			keepPart := match[3]
			if strings.HasPrefix(keepPart, "kh") {
				fmt.Sscanf(keepPart[2:], "%d", &keepHigh)
				if keepHigh == 0 {
					keepHigh = 1
				}
			} else if strings.HasPrefix(keepPart, "kl") {
				fmt.Sscanf(keepPart[2:], "%d", &keepLow)
				if keepLow == 0 {
					keepLow = 1
				}
			} else if keepPart == "kh" {
				keepHigh = 1
			} else if keepPart == "kl" {
				keepLow = 1
			}
		}

		// Roll the dice
		var rolls []int
		for i := 0; i < count; i++ {
			roll := r.rng.Intn(sides) + 1
			rolls = append(rolls, roll)
		}

		// Apply keep modifiers
		keptRolls := make([]int, 0)
		if keep {
			if keepHigh > 0 {
				// Sort descending
				sorted := make([]int, len(rolls))
				copy(sorted, rolls)
				for i := range sorted {
					for j := i + 1; j < len(sorted); j++ {
						if sorted[j] > sorted[i] {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					}
				}
				for i := 0; i < keepHigh && i < len(sorted); i++ {
					keptRolls = append(keptRolls, sorted[i])
				}
			} else if keepLow > 0 {
				// Sort ascending
				sorted := make([]int, len(rolls))
				copy(sorted, rolls)
				for i := range sorted {
					for j := i + 1; j < len(sorted); j++ {
						if sorted[j] < sorted[i] {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					}
				}
				for i := 0; i < keepLow && i < len(sorted); i++ {
					keptRolls = append(keptRolls, sorted[i])
				}
			}
		} else {
			keptRolls = rolls
		}

		// Calculate subtotal
		subtotal := modifier
		for _, roll := range keptRolls {
			subtotal += roll
		}
		total += subtotal

		dieRoll := DieRoll{
			Dice:   match[0],
			Result: subtotal,
			Kept:   true,
			Details: rolls,
		}
		if keep {
			dieRoll.Kept = false // Will be set correctly below
		}
		result.Rolls = append(result.Rolls, dieRoll)
	}

	result.Total = total
	return result, nil
}

// RollAdvantage rolls with advantage (2d20 keep highest).
func (r *Roller) RollAdvantage() AdvantageResult {
	roll1 := r.rng.Intn(20) + 1
	roll2 := r.rng.Intn(20) + 1
	
	selected := roll1
	if roll2 > roll1 {
		selected = roll2
	}

	return AdvantageResult{
		Roll1:      roll1,
		Roll2:      roll2,
		Selected:   selected,
		IsAdvantage: true,
	}
}

// RollDisadvantage rolls with disadvantage (2d20 keep lowest).
func (r *Roller) RollDisadvantage() AdvantageResult {
	roll1 := r.rng.Intn(20) + 1
	roll2 := r.rng.Intn(20) + 1
	
	selected := roll1
	if roll2 < roll1 {
		selected = roll2
	}

	return AdvantageResult{
		Roll1:      roll1,
		Roll2:      roll2,
		Selected:   selected,
		IsAdvantage: false,
	}
}

// SimpleRoll rolls a single die with given number of sides.
func (r *Roller) SimpleRoll(sides int) int {
	return r.rng.Intn(sides) + 1
}

// RollMultiple rolls multiple dice and returns all results.
func (r *Roller) RollMultiple(count, sides int) []int {
	results := make([]int, count)
	for i := 0; i < count; i++ {
		results[i] = r.rng.Intn(sides) + 1
	}
	return results
}

// FormatResult formats the roll result in retro ASCII style.
func FormatResult(result *RollResult) string {
	var sb strings.Builder
	sb.WriteString("┌─────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ 🎲 %-17s │\n", result.Expression))
	sb.WriteString("├─────────────────────┤\n")

	for _, roll := range result.Rolls {
		if len(roll.Details) > 0 {
			details := make([]string, len(roll.Details))
			for i, d := range roll.Details {
				details[i] = fmt.Sprintf("%d", d)
			}
			sb.WriteString(fmt.Sprintf("│ %s: [%s] = %-6d │\n", roll.Dice, strings.Join(details, ","), roll.Result))
		} else {
			sb.WriteString(fmt.Sprintf("│ %s = %-12d │\n", roll.Dice, roll.Result))
		}
	}

	sb.WriteString("├─────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ TOTAL: %-11d │\n", result.Total))
	sb.WriteString("└─────────────────────┘\n")
	return sb.String()
}

// FormatAdvantage formats the advantage/disadvantage result.
func FormatAdvantage(result AdvantageResult) string {
	var sb strings.Builder
	rollType := "DISADVANTAGE"
	keepStr := "kl1"
	if result.IsAdvantage {
		rollType = "ADVANTAGE"
		keepStr = "kh1"
	}

	sb.WriteString("┌─────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ 🎲 2d20%s        │\n", keepStr))
	sb.WriteString("├─────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Roll 1: %-11d │\n", result.Roll1))
	sb.WriteString(fmt.Sprintf("│ Roll 2: %-11d │\n", result.Roll2))
	sb.WriteString("├─────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ %s     │\n", rollType))
	sb.WriteString(fmt.Sprintf("│ Selected: %-7d │\n", result.Selected))
	sb.WriteString("└─────────────────────┘\n")
	return sb.String()
}