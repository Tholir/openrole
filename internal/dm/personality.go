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
	// Always include base text, but may prepend Antarctic flavor
	return baseText
}

// ShouldInjectAntarcticaJoke returns true based on HumorLevel.
// Approximately every (10 - HumorLevel) calls will trigger a gag.
func (p *Personality) ShouldInjectAntarcticaJoke() bool {
	// With HumorLevel 3 (default), ~70% chance to inject
	// With HumorLevel 7, ~30% chance to inject
	return (float64(p.TotalVotes%10) < p.HumorLevel)
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
