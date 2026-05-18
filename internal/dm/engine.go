// Package dm provides the Dungeon Master engine for OPENROLE.
package dm

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/gentleman-programming/openrole/internal/types"
)

// DMEngine is the central Dungeon Master engine that orchestrates
// personality, difficulty calculation, and narration.
type DMEngine struct {
	Personality *Personality
	Difficulty  *DifficultyCalculator
	Narrator    *Narrator
	Persona     *Persona
}

// NewDMEngine creates a new DM engine with the given personality and persona.
func NewDMEngine(personality *Personality, persona *Persona) *DMEngine {
	return &DMEngine{
		Personality: personality,
		Difficulty:  NewDifficultyCalculator(),
		Narrator:    NewNarrator(persona),
		Persona:     persona,
	}
}

// DifficultyCalculator handles DC calculation for action checks.
// Exposes methods from the difficulty module.
type DifficultyCalculator struct{}

// NewDifficultyCalculator creates a new difficulty calculator.
func NewDifficultyCalculator() *DifficultyCalculator {
	return &DifficultyCalculator{}
}

// ResolveAction handles a character action request and returns the narrated outcome.
// It determines whether the action is reasonable (requires dice) or unreasonable
// (denied by DM authority), then narrates accordingly.
func (e *DMEngine) ResolveAction(ctx context.Context, action string, character *types.Character, dmState *types.DMState) (string, error) {
	// Update narrator mood from DM state
	if dmState != nil {
		e.Narrator.SetMood(dmState.Mood)
	}

	// ------------------------------------------------------------
	// [LLM INJECTION POINT]
	// Before the DM engine makes a determination, inject an LLM call here
	// to evaluate the action's reasonableness based on:
	// - Character capabilities and inventory
	// - Current scene and world state
	// - NPC attitudes and quest context
	//
	// Expected LLM input:
	//   - action description
	//   - character profile (stats, inventory, level)
	//   - current scene/NPCs
	//   - dmState.Personality traits
	//
	// Expected LLM output (struct):
	//   - isReasonable: bool
	//   - denialReason: string (if unreasonable)
	//   - suggestedDC: DifficultyClass (if reasonable)
	//   - suggestedStat: string (e.g., "STR", "DEX")
	//   - situationMods: []SituationModOption
	//
	// Example:
	//   result, err := e.llmEvaluateAction(ctx, action, character, dmState)
	//   if err != nil {
	//       return "", fmt.Errorf("LLM evaluation failed: %w", err)
	//   }
	//   isReasonable := result.isReasonable
	//   denialReason := result.denialReason
	//   ...
	// ------------------------------------------------------------

	// Determine if action is reasonable using heuristic logic.
	// This is a fallback until LLM injection is implemented.
	isReasonable, denialReason := e.evaluateActionReasonable(action, character, dmState)

	if !isReasonable {
		return e.narrateDenial(action, denialReason), nil
	}

	// Calculate difficulty and roll for reasonable actions
	dc, statName, err := CalculateDCWithCheck(
		character,
		DCMedium,
		"DEX",
		WithCircumstance(0),
	)
	if err != nil {
		return "", fmt.Errorf("failed to calculate difficulty: %w", err)
	}

	// Perform the dice roll
	roll := e.rollD20()

	// Evaluate the roll
	succeeded, margin, _ := EvaluateRoll(roll, dc)

	// Check for critical results
	isCritical := roll == 20 || (succeeded && roll >= CriticalThreshold(dc))

	// Build roll result for narration
	result := RollResult{
		Roll:             roll,
		DC:               dc,
		StatModifier:     0,
		TotalModifier:    0,
		EffectiveBonus:   0,
		FinalResult:      roll,
		Success:          succeeded,
		Critical:         isCritical,
		Margin:           margin,
		ActionDescription: action,
		AbilityUsed:      statName,
	}

	// Narrate the outcome
	return e.Narrator.NarrateRoll(result), nil
}

// evaluateActionReasonable heuristically determines if an action is reasonable.
// Returns (isReasonable, denialReason).
// This is a placeholder for LLM-based evaluation.
func (e *DMEngine) evaluateActionReasonable(action string, character *types.Character, dmState *types.DMState) (bool, string) {
	// Placeholder logic - replace with LLM call
	// For now, deny obviously impossible actions

	// Check for empty action
	if action == "" {
		return false, "The air grows cold... but you haven't specified what you're trying to do."
	}

	// Deny direct damage to other players/NPCs without consent
	// (This is a game rule, not DM discretion)
	// TODO: Implement proper PvP consent system

	// Deny actions that violate basic physics
	if contains(action, "walk through walls", "breathe underwater without magic", "fly without wings or magic") {
		return false, "Even in this frozen realm, the laws of nature hold... mostly."
	}

	// Default to reasonable if we can't determine otherwise
	return true, ""
}

// narrateDenial produces a flavorful denial without dice rolling.
func (e *DMEngine) narrateDenial(action, reason string) string {
	denialTemplates := []string{
		"The cold freezes your ambition. %s",
		"Not even the aurora australis could justify that. %s",
		"Brrr. That's not how things work here. %s",
		"Ice crack beneath your feet as opportunity slips away. %s",
		"Even at the bottom of the world, we have standards. %s",
	}

	template := denialTemplates[rand.Intn(len(denialTemplates))]
	denial := fmt.Sprintf(template, reason)

	// Apply Antarctic flavor if personality allows
	if e.Personality != nil {
		return e.Personality.FlavorText(denial)
	}

	return denial
}

// rollD20 simulates a d20 roll with advantage/disadvantage consideration.
// Currently rolls a single d20; extend to support advantage/disadvantage.
func (e *DMEngine) rollD20() int {
	return rand.Intn(20) + 1
}

// contains is a simple helper to check if a string contains any of the given substrings.
func contains(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(substr) <= len(s) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}