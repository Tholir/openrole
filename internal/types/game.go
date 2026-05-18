// Package types defines the core domain types for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin.
package types

import (
	"time"

	"github.com/google/uuid"
)

// Character represents a D&D 5e character in the game.
type Character struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Class     string    `json:"class"`
	Level     int       `json:"level"`
	Race      string    `json:"race"`
	HP        int       `json:"hp"`
	MaxHP     int       `json:"max_hp"`
	AC        int       `json:"ac"`
	Stats     Stats     `json:"stats"`
	Skills    []string  `json:"skills"`
	Spells    []string  `json:"spells"`
	Inventory []Item    `json:"inventory"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Stats represents the six D&D 5e ability scores.
type Stats struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// Modifier returns the ability modifier for a given score.
func (s Stats) Modifier(score int) int {
	return (score - 10) / 2
}

// Ability names for D&D 5e.
const (
	AbilityStr     = "STR"
	AbilityDex     = "DEX"
	AbilityCon     = "CON"
	AbilityInt     = "INT"
	AbilityWis     = "WIS"
	AbilityCha     = "CHA"
)

// ModifierMap returns a map of ability names to their modifiers.
func (s Stats) ModifierMap() map[string]int {
	return map[string]int{
		AbilityStr: s.Modifier(s.Strength),
		AbilityDex: s.Modifier(s.Dexterity),
		AbilityCon: s.Modifier(s.Constitution),
		AbilityInt: s.Modifier(s.Intelligence),
		AbilityWis: s.Modifier(s.Wisdom),
		AbilityCha: s.Modifier(s.Charisma),
	}
}

// Item represents a physical item in a character's inventory.
type Item struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`   // in pounds
	Quantity int     `json:"quantity"`
	Equipped bool    `json:"equipped"`
}

// InventoryWeight calculates the total weight of all items.
func (c *Character) InventoryWeight() float64 {
	var total float64
	for _, item := range c.Inventory {
		total += item.Weight * float64(item.Quantity)
	}
	return total
}

// EncumbranceStatus returns the encumbrance status based on D&D 5e rules.
// Light load: up to (STR × 3) lbs, Medium: up to (STR × 6) lbs, Heavy: up to (STR × 15) lbs.
func (c *Character) EncumbranceStatus() (string, float64) {
	maxCarry := float64(c.Stats.Strength * 15)
	weight := c.InventoryWeight()
	percent := weight / maxCarry * 100

	switch {
	case percent <= 20:
		return "Unencumbered", percent
	case percent <= 40:
		return "Encumbered", percent
	default:
		return "Heavily Encumbered", percent
	}
}

// EncumbranceMax returns the maximum carry capacity in pounds.
func (c *Character) EncumbranceMax() float64 {
	return float64(c.Stats.Strength * 15)
}

// Player represents a human player in the roleplaying session.
type Player struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CharacterID  string    `json:"character_id"`
	SessionID    string    `json:"session_id"`
	IsConnected  bool      `json:"is_connected"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

// NewPlayer creates a new player with a generated UUID.
func NewPlayer(name string) *Player {
	now := time.Now()
	return &Player{
		ID:          uuid.New().String(),
		Name:        name,
		IsConnected: true,
		JoinedAt:    now,
		LastActiveAt: now,
	}
}

// DMState represents the Dungeon Master's session state.
// The DM is an Antarctic-based AI with an evolving, quirky personality.
type DMState struct {
	ID             string         `json:"id"`
	SessionID      string         `json:"session_id"`
	Personality    Personality    `json:"personality"`
	CurrentScene   string         `json:"current_scene"`
	ActiveNPCs     []NPC          `json:"active_npcs"`
	QuestLog       []Quest        `json:"quest_log"`
	Mood           Mood           `json:"mood"`
	StoryFlags     map[string]bool `json:"story_flags"`
	SessionNumber  int            `json:"session_number"`
	TotalPlayTime  time.Duration  `json:"total_play_time"`
}

// Personality defines the DM's evolving character traits.
type Personality struct {
	BaseDescription string   `json:"base_description"`
	Traits          []string `json:"traits"`
	Quirks          []string `json:"quirks"`
	IcebreakerFacts []string `json:"icebreaker_facts"` // Antarctic-themed conversation starters
}

// Mood represents the DM's current emotional state affecting narration.
type Mood int

const (
	MoodNeutral Mood = iota
	MoodDramatic
	MoodWhimsical
	MoodOminous
	MoodCheerful
	MoodMelancholy
)

// String returns a human-readable mood description.
func (m Mood) String() string {
	switch m {
	case MoodDramatic:
		return "dramatic"
	case MoodWhimsical:
		return "whimsical"
	case MoodOminous:
		return "ominous"
	case MoodCheerful:
		return "cheerful"
	case MoodMelancholy:
		return "melancholy"
	default:
		return "neutral"
	}
}

// NPC represents a non-player character controlled by the DM.
type NPC struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Traits      []string `json:"traits"`
	IsHostile   bool     `json:"is_hostile"`
	Location    string   `json:"location"`
}

// Quest represents an adventure quest or objective.
type Quest struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	IsComplete  bool     `json:"is_complete"`
	Reward      string   `json:"reward"`
}

// NewDMState creates a new DM state with Antarctic-themed personality.
func NewDMState(sessionID string) *DMState {
	return &DMState{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Personality: Personality{
			BaseDescription: "A Dungeon Master based in Antarctica with an eccentric personality shaped by isolation and extreme weather. Speaks with dry wit and occasional pops of enthusiasm when dice rolls go unusually.",
			Traits: []string{
				"dry humor",
				"observant",
				"dramatic timing",
				"oddly comforting",
			},
			Quirks: []string{
				"References Antarctic wildlife in analogies",
				"Mentions the aurora australis when describing magical phenomena",
				"Occasionally mentions the cold seeping in",
				"Uses 'brrr' to emphasize chills",
			},
			IcebreakerFacts: []string{
				"Did you know Emperor penguins can dive deeper than most submarines?",
				"The aurora australis is caused by solar winds colliding with atmospheric gases.",
				"Antarctica has no native ant species whatsoever.",
				"Vostok Station recorded the coldest temperature ever: -89°C.",
			},
		},
		Mood:          MoodNeutral,
		StoryFlags:    make(map[string]bool),
		SessionNumber: 1,
	}
}

// Vote represents a player vote during gameplay (e.g., initiative, group decisions).
type Vote struct {
	ID          string      `json:"id"`
	SessionID   string      `json:"session_id"`
	VoteType    VoteType    `json:"vote_type"`
	Question    string      `json:"question"`
	Options     []string    `json:"options"`
	Votes       map[string]int `json:"votes"` // option index -> count
	Voters      map[string]string `json:"voters"` // playerID -> option index
	IsOpen      bool        `json:"is_open"`
	CreatedAt   time.Time   `json:"created_at"`
	ClosedAt    *time.Time  `json:"closed_at,omitempty"`
}

// VoteType categorizes different voting scenarios.
type VoteType int

const (
	VoteTypeInitiative VoteType = iota
	VoteTypeGroupDecision
	VoteTypeContest
	VoteTypePoll
)

// NewVote creates a new open vote.
func NewVote(sessionID string, voteType VoteType, question string, options []string) *Vote {
	return &Vote{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		VoteType:  voteType,
		Question:  question,
		Options:   options,
		Votes:     make(map[string]int),
		Voters:    make(map[string]string),
		IsOpen:    true,
		CreatedAt: time.Now(),
	}
}

// CastVote records a player's vote.
func (v *Vote) CastVote(playerID string, optionIndex string) bool {
	if !v.IsOpen {
		return false
	}
	if _, exists := v.Voters[playerID]; exists {
		return false // already voted
	}
	v.Voters[playerID] = optionIndex
	v.Votes[optionIndex]++
	return true
}

// Close ends the voting period.
func (v *Vote) Close() {
	v.IsOpen = false
	now := time.Now()
	v.ClosedAt = &now
}

// WinningOption returns the option index with the most votes.
func (v *Vote) WinningOption() string {
	maxCount := 0
	winner := ""
	for opt, count := range v.Votes {
		if count > maxCount {
			maxCount = count
			winner = opt
		}
	}
	return winner
}
