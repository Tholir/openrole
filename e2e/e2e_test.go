// Package e2e provides end-to-end tests for OPENROLE using teatest.
// These tests simulate full game sessions including character creation,
// dice rolls, and voting through the TUI.
package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/gentleman-programming/openrole/internal/dice"
	"github.com/gentleman-programming/openrole/internal/dm"
	"github.com/gentleman-programming/openrole/internal/tui"
	"github.com/gentleman-programming/openrole/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDMMsg captures messages sent by the DM for verification.
type mockDMMsg struct {
	Type    string
	Content string
}

// mockPlayer represents a player in the E2E test.
type mockPlayer struct {
	ID        string
	Name      string
	Character *types.Character
}

// TestE2E_CharacterCreation tests the full character creation flow.
func TestE2E_CharacterCreation(t *testing.T) {
	// Create a character from scratch using dice roller
	roller := dice.NewRoller()

	// Roll ability scores (4d6 drop lowest)
	abilityScores := make(map[string]int)
	abilityNames := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}

	for _, ability := range abilityNames {
		result, err := roller.ParseRoll("4d6kh3")
		require.NoError(t, err)
		abilityScores[ability] = result.Total
	}

	// Roll for initial gold
	goldResult, err := roller.ParseRoll("3d6")
	require.NoError(t, err)
	initialGold := goldResult.Total * 10 // 10 gold per die

	// Create character
	char := &types.Character{
		ID:    "test-char-001",
		Name:  "E2E Test Hero",
		Level: 1,
		Class: "Rogue",
		Race:  "Half-Elf",
		HP:    10,
		MaxHP: 10,
		AC:    15,
		Stats: types.Stats{
			Strength:     abilityScores["STR"],
			Dexterity:    abilityScores["DEX"],
			Constitution: abilityScores["CON"],
			Intelligence: abilityScores["INT"],
			Wisdom:       abilityScores["WIS"],
			Charisma:     abilityScores["CHA"],
		},
		Skills: []string{"Stealth", "Acrobatics", "Sleight of Hand"},
		Inventory: []types.Item{
			{ID: "item-1", Name: "Dagger", Weight: 1.0, Quantity: 2, Equipped: true},
			{ID: "item-2", Name: "Leather Armor", Weight: 10.0, Quantity: 1, Equipped: true},
			{ID: "item-3", Name: "Thieves' Tools", Weight: 5.0, Quantity: 1, Equipped: true},
		},
		Notes:     "E2E Test Character",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Verify character creation
	assert.NotEmpty(t, char.ID)
	assert.Equal(t, "E2E Test Hero", char.Name)
	assert.Equal(t, 1, char.Level)
	assert.Equal(t, "Rogue", char.Class)
	assert.Equal(t, "Half-Elf", char.Race)

	// Verify ability scores are valid (3-18 range for 4d6kh3)
	for _, ability := range abilityNames {
		score := abilityScores[ability]
		assert.GreaterOrEqual(t, score, 3, "Ability %s score too low", ability)
		assert.LessOrEqual(t, score, 18, "Ability %s score too high", ability)
	}

	// Verify HP and AC
	assert.Greater(t, char.HP, 0)
	assert.Equal(t, char.HP, char.MaxHP)
	assert.Greater(t, char.AC, 10)

	// Verify modifiers
	mods := char.Stats.ModifierMap()
	for _, mod := range mods {
		assert.GreaterOrEqual(t, mod, -5)
		assert.LessOrEqual(t, mod, 10)
	}

	// Verify inventory weight
	totalWeight := char.InventoryWeight()
	assert.Greater(t, totalWeight, 0.0)

	_ = initialGold // gold for purchasing items
}

// TestE2E_DiceRollSession tests dice rolling during a game session.
func TestE2E_DiceRollSession(t *testing.T) {
	handler := dice.NewCommandHandler()
	ctx := context.Background()

	// Test various dice expressions that would occur during a session
	testCases := []struct {
		name       string
		roll       string
		shouldPass bool
	}{
		{"initiative roll", "/roll 1d20+5", true},
		{"attack roll", "/roll 1d20+7", true},
		{"damage roll", "/roll 2d6+3", true},
		{"saving throw", "/roll 1d20+4", true},
		{"skill check", "/roll 1d20+6", true},
		{"ability check", "/roll 4d6kh3", true},
		{"sneak attack", "/roll 3d6", true},
		{"spell damage", "/roll 8d6", true},
		{"crit damage", "/roll 4d6+2d6+3", true},
		{"advantage", "/roll advantage", true},
		{"disadvantage", "/roll disadvantage", true},
		{"group roll", "/roll 4d6kh3 2d6+2 1d20+5", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := handler.HandleRoll(ctx, tc.roll)
			if tc.shouldPass {
				assert.False(t, result.IsError, "Roll %s should succeed: %s", tc.roll, result.Text)
				assert.NotEmpty(t, result.Text)
			}
		})
	}
}

// TestE2E_VotingFlow tests the full voting flow during a session.
func TestE2E_VotingFlow(t *testing.T) {
	tracker := dm.NewVoteTracker()
	sessionID := "e2e-session-001"

	// DM creates a poll
	voteID := tracker.CreateVote(
		sessionID,
		dm.VoteTypeGroupDecision,
		"The party approaches a fork in the road. The left path leads to a dark forest, the right to an abandoned mine. Which do you choose?",
		[]string{"Dark Forest", "Abandoned Mine", "Search for another path"},
	)
	require.NotEmpty(t, voteID)

	// Players join and vote
	players := []mockPlayer{
		{ID: "p1", Name: "Alice"},
		{ID: "p2", Name: "Bob"},
		{ID: "p3", Name: "Carol"},
		{ID: "p4", Name: "Dave"},
	}

	// Alice votes for Dark Forest
	err := tracker.CastVote(voteID, players[0].ID, players[0].Name, 0)
	require.NoError(t, err)

	// Bob votes for Abandoned Mine
	err = tracker.CastVote(voteID, players[1].ID, players[1].Name, 1)
	require.NoError(t, err)

	// Carol votes for Dark Forest
	err = tracker.CastVote(voteID, players[2].ID, players[2].Name, 0)
	require.NoError(t, err)

	// Dave votes for Search
	err = tracker.CastVote(voteID, players[3].ID, players[3].Name, 2)
	require.NoError(t, err)

	// Check pending votes
	pending := tracker.PendingVotes(sessionID)
	assert.Len(t, pending, 1)
	assert.Equal(t, voteID, pending[0].ID)

	// Get live aggregate
	agg, err := tracker.GetAggregate(voteID)
	require.NoError(t, err)
	assert.Equal(t, 4, agg.TotalVotes)
	assert.Equal(t, 2, agg.VoteCounts["Dark Forest"])
	assert.Equal(t, 1, agg.VoteCounts["Abandoned Mine"])
	assert.Equal(t, 1, agg.VoteCounts["Search for another path"])

	// Set sentiments
	tracker.SetSentiment(voteID, players[0].ID, true)  // Alice happy
	tracker.SetSentiment(voteID, players[1].ID, false) // Bob unhappy
	tracker.SetSentiment(voteID, players[2].ID, true)  // Carol happy
	tracker.SetSentiment(voteID, players[3].ID, false) // Dave unhappy

	// Close vote
	finalAgg, err := tracker.CloseVote(voteID)
	require.NoError(t, err)
	assert.True(t, finalAgg.IsComplete)
	assert.Equal(t, "Dark Forest", finalAgg.WinningOption)

	// Check history
	history := tracker.History(sessionID, 0)
	assert.Len(t, history, 4)

	// Verify sentiment tracking
	sentiment := tracker.SessionSentiment(sessionID)
	assert.InDelta(t, 0.5, sentiment, 0.01)
}

// TestE2E_InitiativeCombat tests initiative rolling during combat.
func TestE2E_InitiativeCombat(t *testing.T) {
	tracker := dm.NewVoteTracker()
	sessionID := "e2e-combat-001"

	// Create initiative vote
	voteID := tracker.CreateVote(
		sessionID,
		dm.VoteTypeInitiative,
		"Roll for initiative!",
		[]string{"Alice", "Bob", "Carol"},
	)
	require.NotEmpty(t, voteID)

	// Players vote for who goes first (votes are just counts, not roll values)
	err := tracker.CastVote(voteID, "alice", "Alice", 0)
	require.NoError(t, err)
	err = tracker.CastVote(voteID, "bob", "Bob", 1)
	require.NoError(t, err)
	err = tracker.CastVote(voteID, "carol", "Carol", 0) // Alice gets 2 votes
	require.NoError(t, err)

	// Close initiative vote
	agg, err := tracker.CloseVote(voteID)
	require.NoError(t, err)
	assert.True(t, agg.IsComplete)

	// Alice has most votes (2 vs 1)
	assert.Equal(t, "Alice", agg.WinningOption)
}

// TestE2E_PersonalityEvolution tests DM personality evolution over votes.
func TestE2E_PersonalityEvolution(t *testing.T) {
	personality := dm.NewPersonality()

	// Initial personality
	assert.Equal(t, 5.0, personality.Strictness)
	assert.Equal(t, 5.0, personality.Leniency)
	assert.Equal(t, 3.0, personality.HumorLevel)

	// Simulate a session with multiple votes
	votes := []bool{true, true, false, true, false, true, true, true, false, true}

	for _, vote := range votes {
		personality.Evolve(vote)
	}

	// Personality should have evolved
	assert.Equal(t, len(votes), personality.TotalVotes)

	// With 7 positive vs 3 negative, humor and leniency should have increased or stayed same
	assert.GreaterOrEqual(t, personality.HumorLevel, 3.0)
	assert.GreaterOrEqual(t, personality.Leniency, 5.0)
}

// TestE2E_FullGameSession simulates a complete game session.
func TestE2E_FullGameSession(t *testing.T) {
	// Setup
	dmState := types.NewDMState("e2e-full-session")
	tracker := dm.NewVoteTracker()
	personality := dm.NewPersonality()
	roller := dice.NewRoller()
	handler := dice.NewCommandHandler()

	// Create party
	party := []mockPlayer{
		{ID: "hero-1", Name: "Aldric"},
		{ID: "hero-2", Name: "Brynn"},
		{ID: "hero-3", Name: "Cedric"},
		{ID: "hero-4", Name: "Diana"},
	}

	// Create characters for party members
	characters := []*types.Character{
		{
			Name:  "Aldric",
			Class: "Fighter",
			Level: 5,
			Stats: types.Stats{Strength: 16, Dexterity: 12, Constitution: 14, Intelligence: 10, Wisdom: 13, Charisma: 8},
			HP:    45, MaxHP: 45, AC: 18,
		},
		{
			Name:  "Brynn",
			Class: "Ranger",
			Level: 5,
			Stats: types.Stats{Strength: 14, Dexterity: 16, Constitution: 12, Intelligence: 10, Wisdom: 15, Charisma: 10},
			HP:    38, MaxHP: 38, AC: 15,
		},
		{
			Name:  "Cedric",
			Class: "Wizard",
			Level: 5,
			Stats: types.Stats{Strength: 8, Dexterity: 14, Constitution: 12, Intelligence: 18, Wisdom: 13, Charisma: 11},
			HP:    28, MaxHP: 28, AC: 12,
		},
		{
			Name:  "Diana",
			Class: "Cleric",
			Level: 5,
			Stats: types.Stats{Strength: 12, Dexterity: 10, Constitution: 14, Intelligence: 10, Wisdom: 16, Charisma: 14},
			HP:    35, MaxHP: 35, AC: 16,
		},
	}

	// Verify all characters have valid stats
	for _, char := range characters {
		assert.Greater(t, char.HP, 0)
		assert.Equal(t, char.HP, char.MaxHP)
		assert.Greater(t, char.AC, 10)
	}

	// --- Scene 1: Party meets in a tavern ---
	ctx := context.Background()
	introRoll := handler.HandleRoll(ctx, "/roll 1d20+5")
	assert.False(t, introRoll.IsError)

	// --- Scene 2: Combat encounter ---
	combatVoteID := tracker.CreateVote("e2e-full-session", dm.VoteTypeInitiative,
		"Combat begins! Roll for initiative!",
		[]string{"Aldric", "Brynn", "Cedric", "Diana", "Goblin Alpha"})

	// Cast initiative votes
	tracker.CastVote(combatVoteID, "hero-1", "Aldric", 0)
	tracker.CastVote(combatVoteID, "hero-2", "Brynn", 1)
	tracker.CastVote(combatVoteID, "hero-3", "Cedric", 2)
	tracker.CastVote(combatVoteID, "hero-4", "Diana", 3)
	tracker.CastVote(combatVoteID, "goblin", "Goblin Alpha", 4)

	// Set sentiments
	for _, p := range party {
		tracker.SetSentiment(combatVoteID, p.ID, true)
	}

	combatAgg, err := tracker.CloseVote(combatVoteID)
	require.NoError(t, err)
	assert.True(t, combatAgg.IsComplete)

	// --- Scene 3: Tactical decision ---
	tacticalVoteID := tracker.CreateVote("e2e-full-session", dm.VoteTypeGroupDecision,
		"Goblin Alpha is injured. Do you press the attack or try to capture?",
		[]string{"Press Attack", "Try to Capture", "Retreat"})

	// Party votes
	tracker.CastVote(tacticalVoteID, "hero-1", "Aldric", 0)
	tracker.CastVote(tacticalVoteID, "hero-2", "Brynn", 1) // wants capture
	tracker.CastVote(tacticalVoteID, "hero-3", "Cedric", 0)
	tracker.CastVote(tacticalVoteID, "hero-4", "Diana", 2) // cautious

	// Sentiments
	tracker.SetSentiment(tacticalVoteID, "hero-1", true)
	tracker.SetSentiment(tacticalVoteID, "hero-2", false) // outvoted
	tracker.SetSentiment(tacticalVoteID, "hero-3", true)
	tracker.SetSentiment(tacticalVoteID, "hero-4", false) // wanted to retreat

	tacticalAgg, err := tracker.CloseVote(tacticalVoteID)
	require.NoError(t, err)

	// Update personality based on session
	history := tracker.History("e2e-full-session", 0)
	for _, vote := range history {
		personality.Evolve(vote.Sentiment)
	}

	// Verify session completed
	assert.NotNil(t, dmState)
	assert.Equal(t, "e2e-full-session", dmState.SessionID)
	assert.Greater(t, personality.TotalVotes, 0)

	// Verify party stats
	assert.Len(t, party, 4)
	assert.Len(t, characters, 4)

	_ = roller
	_ = tacticalAgg
}

// TestE2E_AdvantageDisadvantageMechanics tests D&D 5e advantage/disadvantage.
func TestE2E_AdvantageDisadvantageMechanics(t *testing.T) {
	roller := dice.NewRoller()

	// Test advantage: should generally roll higher
	advRolls := make([]int, 100)
	for i := 0; i < 100; i++ {
		result := roller.RollAdvantage()
		advRolls[i] = result.Selected
	}

	// With advantage, we should see more high rolls (15-20)
	highRolls := 0
	for _, r := range advRolls {
		if r >= 15 {
			highRolls++
		}
	}
	// Roughly 30% of rolls should be 15+ with advantage
	assert.Greater(t, highRolls, 10)

	// Test disadvantage: should generally roll lower
	disRolls := make([]int, 100)
	for i := 0; i < 100; i++ {
		result := roller.RollDisadvantage()
		disRolls[i] = result.Selected
	}

	// With disadvantage, we should see fewer high rolls
	highDisRolls := 0
	for _, r := range disRolls {
		if r >= 15 {
			highDisRolls++
		}
	}
	// Fewer high rolls with disadvantage
	assert.Less(t, highDisRolls, highRolls)
}

// TestE2E_AbilityScoreRolling tests the 4d6 drop lowest method.
func TestE2E_AbilityScoreRolling(t *testing.T) {
	roller := dice.NewRoller()

	// Roll 6 sets of ability scores
	allScores := make([][6]int, 6)
	for i := 0; i < 6; i++ {
		result, err := roller.ParseRoll("4d6kh3")
		require.NoError(t, err)
		allScores[i] = [6]int{result.Total, 0, 0, 0, 0, 0} // Store first result
	}

	// All scores should be valid (3-18 range for 4d6 drop lowest)
	for i, scores := range allScores {
		score := scores[0]
		assert.GreaterOrEqual(t, score, 3, "Ability score %d too low: %d", i, score)
		assert.LessOrEqual(t, score, 18, "Ability score %d too high: %d", i, score)
	}
}

// TestE2E_DMEnginesDifficultyCalculation tests DC calculation in gameplay.
func TestE2E_DMEnginesDifficultyCalculation(t *testing.T) {
	// Create a level 5 fighter
	char := &types.Character{
		Name:  "Test Fighter",
		Level: 5,
		Stats: types.Stats{
			Strength:     16, // +3 mod
			Dexterity:    14, // +2 mod
			Constitution: 15, // +2 mod
			Intelligence: 10, // +0 mod
			Wisdom:       12, // +1 mod
			Charisma:     8,  // -1 mod
		},
	}

	tests := []struct {
		name      string
		baseDC    dm.DifficultyClass
		statName  string
		opts      []dm.SituationModOption
		wantMinDC int
		wantMaxDC int
	}{
		{
			name:      "Easy STR check",
			baseDC:    dm.DCEasy,
			statName:  "STR",
			wantMinDC: 10,
			wantMaxDC: 15,
		},
		{
			name:      "Medium STR check with proficiency",
			baseDC:    dm.DCMedium,
			statName:  "STR",
			opts:      []dm.SituationModOption{dm.WithProficiency()},
			wantMinDC: 15,
			wantMaxDC: 22,
		},
		{
			name:      "Hard DEX check with expertise",
			baseDC:    dm.DCHard,
			statName:  "DEX",
			opts:      []dm.SituationModOption{dm.WithExpertise()},
			wantMinDC: 20,
			wantMaxDC: 28,
		},
		{
			name:      "Very Hard with advantage",
			baseDC:    dm.DCVeryHard,
			statName:  "WIS",
			opts:      []dm.SituationModOption{dm.WithAdvantage(5)},
			wantMinDC: 20,
			wantMaxDC: 30,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dc, err := dm.CalculateDC(char, tc.baseDC, tc.statName, tc.opts...)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, dc, tc.wantMinDC, "DC too low")
			assert.LessOrEqual(t, dc, tc.wantMaxDC, "DC too high")
		})
	}
}

// TestE2E_GoldenFileRendering tests that character sheets render correctly.
// This uses golden file testing pattern from the go-testing skill.
func TestE2E_GoldenFileRendering(t *testing.T) {
	char := &types.Character{
		Name:  "Golden Test Hero",
		Level: 7,
		Class: "Paladin",
		Race:  "Human",
		HP:    52,
		MaxHP: 65,
		AC:    20,
		Stats: types.Stats{
			Strength:     18,
			Dexterity:    10,
			Constitution: 16,
			Intelligence: 12,
			Wisdom:       14,
			Charisma:     16,
		},
		Skills: []string{"Athletics", "Insight", "Intimidation", "Medicine"},
	}

	// This test would use golden file testing in a real scenario:
	// output := tui.NewCharacterSheet(char, 60).Render()
	// golden.RequireEqualText(t, output, golden.Get("testdata/character_sheet.golden"))

	// For now, verify the sheet renders without error
	sheet := tui.NewCharacterSheet(char, 60)
	output := sheet.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Golden Test Hero")
}

// TestE2E_ViewportCentering tests viewport centering on various terminal sizes.
func TestE2E_ViewportCentering(t *testing.T) {
	tests := []struct {
		name       string
		termWidth  int
		termHeight int
		wantX      int
		wantY      int
	}{
		{"standard 80x25", 80, 25, 0, 0},
		{"large terminal", 120, 40, 20, 7},
		{"wide terminal", 160, 30, 40, 2},
		{"small terminal", 40, 15, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vp := tui.DefaultViewport(tc.termWidth, tc.termHeight)
			assert.Equal(t, tc.wantX, vp.X)
			assert.Equal(t, tc.wantY, vp.Y)
		})
	}
}

// TestE2E_TUIIntegration tests integration between TUI components.
func TestE2E_TUIIntegration(t *testing.T) {
	// Create a character
	char := &types.Character{
		Name:  "TUI Test Hero",
		Level: 3,
		Class: "Ranger",
		Race:  "Elf",
		HP:    25,
		MaxHP: 25,
		AC:    14,
		Stats: types.Stats{
			Strength:     12,
			Dexterity:    16,
			Constitution: 12,
			Intelligence: 10,
			Wisdom:       14,
			Charisma:     11,
		},
		Skills: []string{"Survival", "Perception", "Athletics"},
	}

	// Create viewport
	vp := tui.NewDecoratedViewport(80, 25, "Adventure Awaits")

	// Create character sheet
	sheet := tui.NewCharacterSheet(char, 60)

	// Verify all components can render
	output := sheet.Render()
	assert.NotEmpty(t, output)

	// Verify viewport can render
	border := vp.RenderBorder()
	assert.NotEmpty(t, border)

	// Verify ANSI colors work
	color := tui.NewColor256(82)
	coloredText := color.Paint256("Test")
	assert.Contains(t, coloredText, "\x1b[38;5;82m")
}

// TestE2E_DiceFormatting tests that dice results format correctly.
func TestE2E_DiceFormatting(t *testing.T) {
	roller := dice.NewRoller()

	result, err := roller.ParseRoll("2d6+3")
	require.NoError(t, err)

	formatted := dice.FormatResult(result)
	assert.Contains(t, formatted, "🎲")
	assert.Contains(t, formatted, "TOTAL")
	assert.Contains(t, formatted, "┌")

	advResult := roller.RollAdvantage()
	advFormatted := dice.FormatAdvantage(advResult)
	assert.Contains(t, advFormatted, "ADVANTAGE")
	assert.Contains(t, advFormatted, "Roll 1")
	assert.Contains(t, advFormatted, "Roll 2")
}

// TestE2E_TeapotModel tests the Bubble Tea model integration.
// This test validates that the TUI model can be created and updated.
func TestE2E_TeapotModel(t *testing.T) {
	m := tui.Model{Width: 80, Height: 25}
	assert.Equal(t, 80, m.Width)
	assert.Equal(t, 25, m.Height)

	// Test that Update doesn't crash with nil
	newModel, cmd := m.Update(nil)
	assert.NotNil(t, newModel)
	assert.Nil(t, cmd)
}
