// Package dm provides the Dungeon Master engine for OPENROLE,
// an Antarctic-themed D&D 5e retro terminal roleplaying plugin.
package dm

import (
	"github.com/gentleman-programming/openrole/internal/types"
)

// Mode represents the DM's operating mode based on accumulated votes.
type Mode int

const (
	// ModeByTheBook follows strict D&D 5e rules.
	ModeByTheBook Mode = iota
	// ModeFlexible allows rule-of-cool and narrative adjustments.
	ModeFlexible
	// ModeNarrator focuses on story over mechanics.
	ModeNarrator
)

// Personality tracks the DM's evolving character traits.
// It evolves based on accumulated player votes and session history.
type Personality struct {
	// Strictness ranges from 0 (very lenient) to 10 (very strict).
	// Affects DC calculations, ruling calls, and punishment severity.
	Strictness float64

	// Leniency ranges from 0 (harsh) to 10 (generous).
	// Affects how often the DM gives second chances, fudges rolls, or
	// provides helpful hints when players are stuck.
	Leniency float64

	// HumorLevel ranges from 0 (serious) to 10 (chaotic).
	// Controls frequency of jokes, gags, and comedic interruptions.
	HumorLevel float64

	// AntarcticaReferences ranges from 0 (none) to 10 (constant).
	// Controls how often the DM references Antarctic phenomena,
	// wildlife, weather, or research station life.
	AntarcticaReferences float64

	// CurrentMode is the DM's current operating mode.
	CurrentMode Mode

	// SessionsPlayed counts how many sessions this personality has spanned.
	SessionsPlayed int

	// TotalVotes cast across all sessions.
	TotalVotes int

	// SentimentHistory tracks the rolling sentiment of votes.
	sentimentHistory []bool // true = positive, false = negative
}

// NewPersonality creates a default DM personality with balanced traits.
func NewPersonality() Personality {
	return Personality{
		Strictness:           5.0,
		Leniency:             5.0,
		HumorLevel:           3.0,
		AntarcticaReferences: 4.0,
		CurrentMode:          ModeByTheBook,
		SessionsPlayed:       0,
		TotalVotes:           0,
		sentimentHistory:     make([]bool, 0, 100),
	}
}

// Evolve updates personality based on a player's vote.
func (p *Personality) Evolve(sentimentPositive bool) {
	p.TotalVotes++
	p.sentimentHistory = append(p.sentimentHistory, sentimentPositive)

	// Keep history bounded to last 50 votes for rolling average
	if len(p.sentimentHistory) > 50 {
		p.sentimentHistory = p.sentimentHistory[1:]
	}

	// Update traits based on sentiment
	delta := 0.2 // moderate adjustment per vote
	if sentimentPositive {
		// Players like current behavior, slightly increase what made them happy
		p.HumorLevel = min(10, p.HumorLevel+delta)
		p.Leniency = min(10, p.Leniency+delta)
	} else {
		// Players disliked something, adjust
		p.Leniency = max(0, p.Leniency-delta)
		p.Strictness = min(10, p.Strictness+delta)
	}

	// Recalculate mode based on traits
	p.updateMode()
}

// updateMode recalculates the DM's operating mode from current traits.
func (p *Personality) updateMode() {
	// Mode is determined by the dominant trait balance
	if p.Strictness > 7 && p.Leniency < 4 {
		p.CurrentMode = ModeByTheBook
	} else if p.Leniency > 7 && p.Strictness < 4 {
		p.CurrentMode = ModeNarrator
	} else {
		p.CurrentMode = ModeFlexible
	}
}

// RollingSentiment returns the positive vote ratio over the last N votes.
func (p *Personality) RollingSentiment(lastN int) float64 {
	if len(p.sentimentHistory) == 0 {
		return 0.5 // neutral default
	}

	start := 0
	if len(p.sentimentHistory) > lastN {
		start = len(p.sentimentHistory) - lastN
	}

	positive := 0
	for i := start; i < len(p.sentimentHistory); i++ {
		if p.sentimentHistory[i] {
			positive++
		}
	}
	return float64(positive) / float64(len(p.sentimentHistory)-start)
}

// FlavorText returns Antarctic-themed flavor text based on current mood.
// The AntarcticaReferences level controls frequency of usage.
func (p *Personality) FlavorText(baseText string) string {
	// Check if we should inject Antarctic flavor based on trait level
	if !p.ShouldInjectAntarcticaJoke() {
		return baseText
	}

	// AntarcticaReferences threshold: only inject if high enough
	if p.AntarcticaReferences < 3 {
		return baseText
	}

	// Build Antarctic flavor based on HumorLevel
	flavor := p.buildAntarcticFlavor()

	// Prepend flavor to base text
	return flavor + "\n" + baseText
}

// buildAntarcticFlavor generates Antarctic-themed flavor text.
// The tone and content vary based on HumorLevel.
func (p *Personality) buildAntarcticFlavor() string {
	// Flavor templates organized by intensity
	penguinFlavors := []string{
		"A nearby penguin watches with what appears to be disapproval.",
		"An Emperor penguin waddles past, utterly indifferent to your quest.",
		"The wind carries what sounds suspiciously like penguin chattering.",
		"A Adelie penguin stares at you judgmentally from an ice floe.",
	}

	weatherFlavors := []string{
		"The cold creeps in just a little more.",
		"A gust of frigid air sweeps through, carrying ice crystals.",
		"The aurora australis shimmers overhead, painting the sky green.",
		"Snow begins to fall, soft and relentless.",
	}

	stationFlavors := []string{
		"Somewhere in the distance, a research station radio crackles to life.",
		"The hum of station equipment provides an odd sense of comfort.",
		"You notice supply crates marked 'McMurdo Station' nearby.",
		"A weathered flag from an Antarctic expedition flaps in the wind.",
	}

	iceFlavors := []string{
		"The ice beneath your feet groans ominously.",
		"Cracks spread across the frozen surface like veins.",
		"An iceberg the size of a castle drifts past on the horizon.",
		"The ice sheet stretches endlessly in all directions.",
	}

	// Select flavor based on HumorLevel ranges
	var flavors []string
	switch {
	case p.HumorLevel >= 8:
		// High humor: more playful penguin references
		flavors = append(penguinFlavors[:2], append(weatherFlavors[:1], stationFlavors[:1]...)...)
	case p.HumorLevel >= 5:
		// Medium humor: balanced mix
		flavors = append(weatherFlavors[:2], append(iceFlavors[:1], stationFlavors[:1]...)...)
	default:
		// Low humor: subtle, atmospheric
		flavors = append(iceFlavors[:2], weatherFlavors[2:]...)
	}

	// Pick one based on TotalVotes for variety
	idx := p.TotalVotes % len(flavors)
	return flavors[idx]
}

// ShouldInjectAntarcticaJoke returns true based on HumorLevel.
// Higher HumorLevel = more jokes (not fewer).
// At HumorLevel 10, always injects (100%).
// At HumorLevel 0, never injects (0%).
func (p *Personality) ShouldInjectAntarcticaJoke() bool {
	if p.HumorLevel <= 0 {
		return false
	}
	if p.HumorLevel >= 10 {
		return true
	}
	// Higher TotalVotes creates natural variation in injection frequency
	// Mod 10 gives us a value 0-9, we inject when value < HumorLevel scaled to 0-9
	threshold := int(p.HumorLevel)
	return (p.TotalVotes%10) < threshold
}

// GetMood returns the DM's current mood based on accumulated state.
func (p *Personality) GetMood() types.Mood {
	sentiment := p.RollingSentiment(10)

	switch {
	case sentiment > 0.7:
		return types.MoodCheerful
	case sentiment > 0.6:
		return types.MoodWhimsical
	case sentiment < 0.3:
		return types.MoodMelancholy
	case p.Strictness > 7:
		return types.MoodOminous
	case p.HumorLevel > 7:
		return types.MoodWhimsical
	default:
		return types.MoodNeutral
	}
}

// Merge combines another personality's traits into this one.
// Used when loading a persisted DM state.
func (p *Personality) Merge(other Personality) {
	p.Strictness = (p.Strictness + other.Strictness) / 2
	p.Leniency = (p.Leniency + other.Leniency) / 2
	p.HumorLevel = (p.HumorLevel + other.HumorLevel) / 2
	p.AntarcticaReferences = (p.AntarcticaReferences + other.AntarcticaReferences) / 2
	p.SessionsPlayed += other.SessionsPlayed
	p.TotalVotes += other.TotalVotes
	p.updateMode()
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
