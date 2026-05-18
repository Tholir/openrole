// Package dm provides the Dungeon Master engine for OPENROLE.
package dm

import (
	"fmt"
	"math"

	"github.com/gentleman-programming/openrole/internal/types"
)

// DifficultyClass represents the DC for a given check.
type DifficultyClass int

// Standard D&D 5e DCs by task severity.
const (
	DCVeryEasy         DifficultyClass = 5
	DCEasy             DifficultyClass = 10
	DCMedium           DifficultyClass = 15
	DCHard             DifficultyClass = 20
	DCVeryHard         DifficultyClass = 25
	DCNearlyImpossible DifficultyClass = 30
)

// SituationMod represents situational modifiers to the roll.
type SituationMod struct {
	// Advantage adds to the roll (positive) or reduces it (negative).
	// Use +5 for advantage, -5 for disadvantage, or custom values.
	Advantage float64

	// Proficiency indicates if the character has proficiency in this skill.
	HasProficiency bool

	// Expertise indicates if the character has expertise (2x proficiency).
	HasExpertise bool

	// CircumstanceModifier is a flat bonus or penalty due to circumstances.
	CircumstanceModifier int

	// MagicalBonus is any bonus from spells, magic items, etc.
	MagicalBonus int
}

// SituationModOption is a functional option for building SituationMod.
type SituationModOption func(*SituationMod)

// WithAdvantage sets the advantage/disadvantage modifier.
func WithAdvantage(bonus float64) SituationModOption {
	return func(s *SituationMod) {
		s.Advantage = bonus
	}
}

// WithProficiency marks the character as proficient.
func WithProficiency() SituationModOption {
	return func(s *SituationMod) {
		s.HasProficiency = true
	}
}

// WithExpertise marks the character as having expertise.
func WithExpertise() SituationModOption {
	return func(s *SituationMod) {
		s.HasExpertise = true
	}
}

// WithCircumstance adds a circumstance modifier.
func WithCircumstance(mod int) SituationModOption {
	return func(s *SituationMod) {
		s.CircumstanceModifier = mod
	}
}

// WithMagicalBonus adds a magical bonus.
func WithMagicalBonus(bonus int) SituationModOption {
	return func(s *SituationMod) {
		s.MagicalBonus = bonus
	}
}

// NewSituationMod creates a SituationMod with the given options.
func NewSituationMod(opts ...SituationModOption) SituationMod {
	s := SituationMod{}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

// CalculateDC computes the final DC for a check.
//
// It follows D&D 5e rules:
// - Base DC is determined by the task difficulty
// - Stat modifier is added from the character's relevant ability score
// - Proficiency bonus is added if the character is proficient (and has expertise)
// - Situational modifiers adjust the final result
func CalculateDC(char *types.Character, baseDC DifficultyClass, statName string, opts ...SituationModOption) (int, error) {
	if char == nil {
		return 0, fmt.Errorf("character is nil")
	}

	statValue, ok := getStatByName(char.Stats, statName)
	if !ok {
		return 0, fmt.Errorf("unknown stat: %s", statName)
	}

	statMod := types.Stats{}.Modifier(statValue)

	situation := NewSituationMod(opts...)

	// Calculate proficiency bonus based on level (D&D 5e standard)
	profBonus := 0
	if situation.HasExpertise {
		profBonus = 2 * ((char.Level-1)/4 + 2)
	} else if situation.HasProficiency {
		profBonus = (char.Level-1)/4 + 2
	}

	finalDC := int(baseDC) + statMod + profBonus + situation.CircumstanceModifier + situation.MagicalBonus

	// Apply advantage/disadvantage (approximate effect on DC)
	// Advantage: effectively +5 to the roll (equivalent to -5 DC)
	// Disadvantage: effectively -5 to the roll (equivalent to +5 DC)
	finalDC -= int(situation.Advantage)

	return finalDC, nil
}

// CalculateDCWithCheck is a convenience function that returns both
// the DC and a description of what contributes to it.
func CalculateDCWithCheck(char *types.Character, baseDC DifficultyClass, statName string, opts ...SituationModOption) (int, string, error) {
	dc, err := CalculateDC(char, baseDC, statName, opts...)
	if err != nil {
		return 0, "", err
	}

	statValue, _ := getStatByName(char.Stats, statName)
	statMod := types.Stats{}.Modifier(statValue)

	situation := NewSituationMod(opts...)
	profBonus := 0
	if situation.HasExpertise {
		profBonus = 2 * ((char.Level-1)/4 + 2)
	} else if situation.HasProficiency {
		profBonus = (char.Level-1)/4 + 2
	}

	desc := fmt.Sprintf("Base %d + Stat %d + Proficiency %d + Circumstance %d + Magic %d + Advantage %.0f",
		baseDC, statMod, profBonus, situation.CircumstanceModifier, situation.MagicalBonus, situation.Advantage)

	return dc, desc, nil
}

// EvaluateRoll determines if a roll succeeds against the DC.
// Returns: succeeded (bool), margin (int), description (string).
func EvaluateRoll(roll int, dc int) (bool, int, string) {
	margin := roll - dc
	succeeded := roll >= dc

	var desc string
	if !succeeded {
		desc = fmt.Sprintf("Failed by %d (rolled %d vs DC %d)", -margin, roll, dc)
	} else if margin == 0 {
		desc = fmt.Sprintf("Succeeded by 0 (exact DC %d)", dc)
	} else {
		desc = fmt.Sprintf("Succeeded by %d (rolled %d vs DC %d)", margin, roll, dc)
	}

	return succeeded, margin, desc
}

// CriticalThreshold returns the minimum roll needed for a critical success.
// On a d20, rolling a 20 is always a hit (even if DC > 20), and rolling 1 is always a fail.
// However, if roll + modifiers >= 2 * DC, it's a critical success.
func CriticalThreshold(dc int) int {
	// Critical hit if total result is at least double the DC
	return int(math.Ceil(float64(dc) * 2))
}

// getStatByName returns the stat value by its name string.
// Returns ok=false if stat not found.
func getStatByName(stats types.Stats, name string) (int, bool) {
	switch name {
	case types.AbilityStr:
		return stats.Strength, true
	case types.AbilityDex:
		return stats.Dexterity, true
	case types.AbilityCon:
		return stats.Constitution, true
	case types.AbilityInt:
		return stats.Intelligence, true
	case types.AbilityWis:
		return stats.Wisdom, true
	case types.AbilityCha:
		return stats.Charisma, true
	default:
		return 0, false
	}
}

// DCFromTask returns an appropriate base DC for a task description.
func DCFromTask(task string) DifficultyClass {
	switch task {
	case "very easy", "easy task", "simple":
		return DCVeryEasy
	case "moderate", "standard", "average":
		return DCEasy
	case "challenging", "moderate difficulty":
		return DCMedium
	case "hard", "difficult":
		return DCHard
	case "very hard", "very difficult":
		return DCVeryHard
	case "nearly impossible", "heroic", "legendary":
		return DCNearlyImpossible
	default:
		return DCMedium // default to medium
	}
}
