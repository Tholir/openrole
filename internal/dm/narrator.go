// Package dm provides the Dungeon Master engine for OPENROLE.
package dm

import (
	"fmt"
	"strings"
	"time"

	"github.com/gentleman-programming/openrole/internal/types"
)

// RollResult represents the outcome of a dice roll.
type RollResult struct {
	// Roll is the raw die result.
	Roll int

	// DC is the difficulty class that was targeted.
	DC int

	// StatModifier applied to the roll.
	StatModifier int

	// TotalModifier is any additional modifier (proficiency, magic, etc).
	TotalModifier int

	// EffectiveBonus is the total bonus to the roll.
	EffectiveBonus int

	// FinalResult is roll + all modifiers.
	FinalResult int

	// Success indicates if the roll met or exceeded the DC.
	Success bool

	// Critical indicates if this was a critical hit or fail.
	Critical bool

	// Margin is how much the roll succeeded or failed by.
	Margin int

	// Description of what was being rolled for.
	ActionDescription string

	// AbilityUsed is the stat used (STR, DEX, etc).
	AbilityUsed string
}

// OutcomeType categorizes the narrative tone of the result.
type OutcomeType int

const (
	// OutcomeCritSuccess is a critical hit.
	OutcomeCritSuccess OutcomeType = iota
	// OutcomeSuccess is a normal success.
	OutcomeSuccess
	// OutcomeFailure is a normal failure.
	OutcomeFailure
	// OutcomeCritFail is a critical failure.
	OutcomeCritFail
)

// Narrator generates flavorful narration for dice roll outcomes.
type Narrator struct {
	persona *Persona
	mood    types.Mood
}

// NewNarrator creates a new DM narrator.
func NewNarrator(persona *Persona) *Narrator {
	return &Narrator{
		persona: persona,
		mood:    types.MoodNeutral,
	}
}

// SetMood updates the narrator's mood for narration flavor.
func (n *Narrator) SetMood(mood types.Mood) {
	n.mood = mood
}

// NarrateRoll generates a full narrative description of a roll result.
func (n *Narrator) NarrateRoll(result RollResult) string {
	var sb strings.Builder

	// Start with DM reaction to the roll
	react := n.persona.ReactToRoll(result.Roll, result.Margin)
	if react != "" {
		sb.WriteString(react)
		sb.WriteString("\n\n")
	}

	// Build the outcome narrative
	sb.WriteString(n.buildOutcomeNarrative(result))

	// Add flavor based on mood
	sb.WriteString(n.addMoodFlavor(result))

	return sb.String()
}

// buildOutcomeNarrative creates the core narrative for the outcome.
func (n *Narrator) buildOutcomeNarrative(result RollResult) string {
	action := result.ActionDescription
	if action == "" {
		action = "the action"
	}

	ability := result.AbilityUsed
	if ability == "" {
		ability = "check"
	}

	// Build the roll summary line
	rollLine := fmt.Sprintf("[%s %s: %d + %d (modifiers) = %d vs DC %d]",
		action, ability, result.Roll, result.TotalModifier, result.FinalResult, result.DC)
	sb := strings.Builder{}
	sb.WriteString(rollLine)
	sb.WriteString("\n\n")

	if result.Critical {
		sb.WriteString(n.narrateCriticalSuccess(result))
	} else if result.Success {
		sb.WriteString(n.narrateSuccess(result))
	} else if result.Critical {
		sb.WriteString(n.narrateCriticalFailure(result))
	} else {
		sb.WriteString(n.narrateFailure(result))
	}

	return sb.String()
}

// narrateCriticalSuccess generates narrative for a critical hit.
func (n *Narrator) narrateCriticalSuccess(r RollResult) string {
	templates := []string{
		"Critical! The dice themselves seem to align with your fate. %s — done with spectacular precision that even the aurora australis would envy.",
		"NATURE ITSELF BOWS TO YOUR SUCCESS! Your %s is executed so flawlessly that even the most hardened adventurers would weep with joy.",
		"The clearest success I've witnessed in all my winters at the station. Your %s achieves what others only dream of.",
	}

	base := templates[time.Now().Unix()%int64(len(templates))]
	return fmt.Sprintf(base, r.ActionDescription)
}

// narrateSuccess generates narrative for a normal success.
func (n *Narrator) narrateSuccess(r RollResult) string {
	marginDesc := ""
	switch {
	case r.Margin >= 10:
		marginDesc = "decisively"
	case r.Margin >= 5:
		marginDesc = "comfortably"
	case r.Margin >= 1:
		marginDesc = "narrowly"
	}

	templates := []string{
		"You succeed %s. Your %s proves adequate — not legendary, but certainly not失败的.",
		"The stars align in your favor, if only briefly. Your %s achieves what was needed.",
		"A solid success. Your %s works, though it won't make the history books.",
	}

	base := templates[time.Now().Unix()%int64(len(templates))]
	return fmt.Sprintf(base, marginDesc, r.ActionDescription)
}

// narrateFailure generates narrative for a normal failure.
func (n *Narrator) narrateFailure(r RollResult) string {
	marginDesc := ""
	switch {
	case r.Margin >= -5:
		marginDesc = "a narrow miss"
	case r.Margin >= -10:
		marginDesc = "a clear failure"
	default:
		marginDesc = "a spectacular failure"
	}

	templates := []string{
		"You fail. Your %s doesn't quite measure up to the challenge. %s is not your destiny today.",
		"Not this time. The dice haven't favored your %s. Perhaps the next roll will be kinder.",
		"The cold reality sets in — your %s has failed. The path forward remains shrouded in uncertainty.",
	}

	base := templates[time.Now().Unix()%int64(len(templates))]
	return fmt.Sprintf(base, r.ActionDescription, marginDesc)
}

// narrateCriticalFailure generates narrative for a critical fail.
func (n *Narrator) narrateCriticalFailure(r RollResult) string {
	templates := []string{
		"Critical failure! The dice mock your efforts with their coldest嘲笑. Even the penguins at the station are cringing.",
		"Nature itself seems to conspire against you. Your %s fails in the most spectacular way possible — a cautionary tale for the ages.",
		"When the dice show 1, even the most optimistic DM feels a chill. Your %s fails catastrophically. The aurora weeps.",
	}

	base := templates[time.Now().Unix()%int64(len(templates))]
	return fmt.Sprintf(base, r.ActionDescription)
}

// addMoodFlavor appends flavor text based on current mood.
func (n *Narrator) addMoodFlavor(r RollResult) string {
	switch n.mood {
	case types.MoodDramatic:
		if r.Success {
			return "\n\n*The crowd goes wild! ...well, the imaginary crowd I have in my head.*"
		}
		return "\n\n*silence falls over the frozen landscape*"
	case types.MoodWhimsical:
		gag := n.persona.MayInjectGag()
		if gag != "" {
			return "\n" + gag
		}
		return ""
	case types.MoodOminous:
		if !r.Success {
			return "\n\n*The shadows grow longer, as if the darkness itself celebrates your failure.*"
		}
		return ""
	case types.MoodCheerful:
		if r.Success {
			return "\n\n*The warmest celebration a frozen heart can manage.*"
		}
		return ""
	case types.MoodMelancholy:
		return "\n\n*Even the endless ice seems to mourn this moment.*"
	default:
		return ""
	}
}

// NarrateContest generates narration for opposed rolls (contests).
func (n *Narrator) NarrateContest(attacker, defender RollResult, attackerName, defenderName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Contest: %s vs %s**\n\n", attackerName, defenderName))
	sb.WriteString(fmt.Sprintf("%s: Roll %d vs DC %d (final: %d)\n", attackerName, attacker.Roll, attacker.DC, attacker.FinalResult))
	sb.WriteString(fmt.Sprintf("%s: Roll %d vs DC %d (final: %d)\n\n", defenderName, defender.Roll, defender.DC, defender.FinalResult))

	if attacker.FinalResult > defender.FinalResult {
		margin := attacker.FinalResult - defender.FinalResult
		sb.WriteString(fmt.Sprintf("%s wins the contest by %d! ", attackerName, margin))
		sb.WriteString(n.persona.ReactToRoll(attacker.Roll, margin))
	} else if defender.FinalResult > attacker.FinalResult {
		margin := defender.FinalResult - attacker.FinalResult
		sb.WriteString(fmt.Sprintf("%s wins the contest by %d! ", defenderName, margin))
		sb.WriteString(n.persona.ReactToRoll(defender.Roll, margin))
	} else {
		sb.WriteString("It's a tie! The result hangs in the balance like a frozen pendulum.")
	}

	return sb.String()
}

// NarrateGroupCheck generates narration for group skill checks.
func (n *Narrator) NarrateGroupCheck(results []RollResult, dc int, taskName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Group %s Check (DC %d)**\n\n", taskName, dc))

	successes := 0
	failures := 0
	critSuccesses := 0
	critFailures := 0

	for _, r := range results {
		status := "FAIL"
		if r.Success {
			status = "SUCCESS"
			successes++
			if r.Critical {
				status = "CRIT!"
				critSuccesses++
			}
		} else {
			failures++
			if r.Critical {
				status = "CRIT FAIL"
				critFailures++
			}
		}
		sb.WriteString(fmt.Sprintf("- %s: %d (final: %d) — **%s**\n",
			r.ActionDescription, r.Roll, r.FinalResult, status))
	}

	sb.WriteString("\n")

	switch {
	case critFailures > 0 && successes == 0:
		sb.WriteString("Total party failure! Even the Antarctic winds seem to mock your collective effort.")
	case critSuccesses > 0 && failures == 0:
		sb.WriteString("Complete and spectacular success! Even the aurora australis celebrates your triumph.")
	case successes > failures:
		sb.WriteString(fmt.Sprintf("Party prevails! %d succeed, %d fail. The path forward remains open.", successes, failures))
	case failures > successes:
		sb.WriteString(fmt.Sprintf("Party struggles! Only %d of %d make it through. Caution is advised.", successes, len(results)))
	default:
		sb.WriteString("Mixed results. The outcome remains uncertain, like the endless Antarctic ice.")
	}

	return sb.String()
}

// QuickNarrate provides a one-line narration for quick results.
func (n *Narrator) QuickNarrate(roll, dc int, success bool) string {
	if roll == 20 || (roll >= dc+10 && success) {
		return "✨ CRITICAL! ✨"
	}
	if roll == 1 || (!success && roll <= 1) {
		return "💨 Critical fail..."
	}
	if success {
		return fmt.Sprintf("✓ Success (rolled %d vs %d)", roll, dc)
	}
	return fmt.Sprintf("✗ Failure (rolled %d vs %d)", roll, dc)
}
