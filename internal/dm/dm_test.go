// Package dm provides the Dungeon Master engine for OPENROLE.
package dm

import (
	"testing"

	"github.com/gentleman-programming/openrole/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Difficulty Calculation Tests ---

func TestCalculateDC(t *testing.T) {
	tests := []struct {
		name     string
		char     *types.Character
		baseDC   DifficultyClass
		statName string
		opts     []SituationModOption
		wantErr  bool
		wantDC   int
	}{
		{
			name: "basic DC with STR modifier",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 1,
				Stats: types.Stats{Strength: 16}, // +3 mod
			},
			baseDC:   DCEasy,
			statName: "STR",
			wantDC:   13, // 10 + 3
		},
		{
			name: "DC with proficiency bonus at level 5",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 5,
				Stats: types.Stats{Strength: 16}, // +3 mod
			},
			baseDC:   DCMedium,
			statName: "STR",
			opts:     []SituationModOption{WithProficiency()},
			wantDC:   21, // 15 + 3 + 3 (level 5 prof bonus = (5-1)/4+2 = 3)
		},
		{
			name: "DC with expertise at level 9",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 9,
				Stats: types.Stats{Intelligence: 18}, // +4 mod
			},
			baseDC:   DCHard,
			statName: "INT",
			opts:     []SituationModOption{WithExpertise()},
			wantDC:   32, // 20 + 4 + 8 (level 9 expertise = 2*4)
		},
		{
			name: "DC with advantage",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 1,
				Stats: types.Stats{Dexterity: 14}, // +2 mod
			},
			baseDC:   DCEasy,
			statName: "DEX",
			opts:     []SituationModOption{WithAdvantage(5)},
			wantDC:   7, // 10 + 2 - 5
		},
		{
			name: "DC with disadvantage",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 1,
				Stats: types.Stats{Constitution: 12}, // +1 mod
			},
			baseDC:   DCMedium,
			statName: "CON",
			opts:     []SituationModOption{WithAdvantage(-5)},
			wantDC:   21, // 15 + 1 + 5
		},
		{
			name: "DC with circumstantial modifier",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 1,
				Stats: types.Stats{Wisdom: 10}, // +0 mod
			},
			baseDC:   DCEasy,
			statName: "WIS",
			opts:     []SituationModOption{WithCircumstance(2)},
			wantDC:   12, // 10 + 0 + 2
		},
		{
			name: "DC with magical bonus",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 1,
				Stats: types.Stats{Charisma: 8}, // -1 mod
			},
			baseDC:   DCEasy,
			statName: "CHA",
			opts:     []SituationModOption{WithMagicalBonus(3)},
			wantDC:   12, // 10 - 1 + 3
		},
		{
			name:    "nil character returns error",
			char:    nil,
			baseDC:  DCEasy,
			wantErr: true,
		},
		{
			name: "unknown stat returns error",
			char: &types.Character{
				Name:  "Test Hero",
				Level: 1,
				Stats: types.Stats{},
			},
			baseDC:   DCEasy,
			statName: "UNKNOWN",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc, err := CalculateDC(tt.char, tt.baseDC, tt.statName, tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantDC, dc)
		})
	}
}

func TestCalculateDCWithCheck(t *testing.T) {
	char := &types.Character{
		Name:  "Test Hero",
		Level: 3,
		Stats: types.Stats{Strength: 15}, // +2 mod
	}
	opts := []SituationModOption{WithProficiency(), WithCircumstance(2)}

	dc, desc, err := CalculateDCWithCheck(char, DCHard, "STR", opts...)
	require.NoError(t, err)
	// DC = 20 (hard) + 2 (str mod) + 2 (level 3 prof) + 2 (circumstance) = 26
	assert.Equal(t, 26, dc)
	assert.NotEmpty(t, desc)
}

func TestEvaluateRoll(t *testing.T) {
	tests := []struct {
		name       string
		roll       int
		dc         int
		wantSucc   bool
		wantMargin int
	}{
		{"success by 5", 15, 10, true, 5},
		{"success by 0 (exact)", 10, 10, true, 0},
		{"failure by 3", 7, 10, false, -3},
		{"natural 1 always fails", 1, 5, false, -4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			succeeded, margin, desc := EvaluateRoll(tt.roll, tt.dc)
			assert.Equal(t, tt.wantSucc, succeeded)
			assert.Equal(t, tt.wantMargin, margin)
			assert.NotEmpty(t, desc)
		})
	}
}

func TestCriticalThreshold(t *testing.T) {
	tests := []struct {
		name string
		dc   int
		want int
	}{
		{"DC 10 critical at 20", 10, 20},
		{"DC 15 critical at 30", 15, 30},
		{"DC 20 critical at 40", 20, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CriticalThreshold(tt.dc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDCFromTask(t *testing.T) {
	tests := []struct {
		task string
		want DifficultyClass
	}{
		{"very easy", DCVeryEasy},
		{"easy task", DCVeryEasy},
		{"simple", DCVeryEasy},
		{"moderate", DCEasy},
		{"standard", DCEasy},
		{"average", DCEasy},
		{"challenging", DCMedium},
		{"moderate difficulty", DCMedium},
		{"hard", DCHard},
		{"difficult", DCHard},
		{"very hard", DCVeryHard},
		{"very difficult", DCVeryHard},
		{"nearly impossible", DCNearlyImpossible},
		{"heroic", DCNearlyImpossible},
		{"legendary", DCNearlyImpossible},
		{"gobbldygook", DCMedium}, // unknown task defaults to DCMedium
	}

	for _, tt := range tests {
		t.Run(tt.task, func(t *testing.T) {
			got := DCFromTask(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Vote Tracking Tests ---

func TestNewVoteTracker(t *testing.T) {
	tracker := NewVoteTracker()
	assert.NotNil(t, tracker.activeVotes)
	assert.NotNil(t, tracker.historicalVotes)
	assert.NotNil(t, tracker.sentimentHistory)
	assert.Len(t, tracker.activeVotes, 0)
}

func TestVoteTracker_CreateVote(t *testing.T) {
	tracker := NewVoteTracker()
	sessionID := "test-session"
	question := "What direction to go?"
	options := []string{"North", "South", "East", "West"}

	voteID := tracker.CreateVote(sessionID, VoteTypeInitiative, question, options)

	assert.NotEmpty(t, voteID)
	agg, err := tracker.GetAggregate(voteID)
	require.NoError(t, err)
	assert.Equal(t, question, agg.Question)
	assert.Equal(t, options, agg.Options)
	assert.Equal(t, 0, agg.TotalVotes)
	assert.False(t, agg.IsComplete)
}

func TestVoteTracker_CastVote(t *testing.T) {
	tracker := NewVoteTracker()
	voteID := tracker.CreateVote("session1", VoteTypeGroupDecision, "Which path?", []string{"Forest", "Cave"})

	err := tracker.CastVote(voteID, "player1", "Alice", 0)
	require.NoError(t, err)

	err = tracker.CastVote(voteID, "player2", "Bob", 1)
	require.NoError(t, err)

	agg, err := tracker.GetAggregate(voteID)
	require.NoError(t, err)
	assert.Equal(t, 2, agg.TotalVotes)
	assert.Equal(t, 1, agg.VoteCounts["Forest"])
	assert.Equal(t, 1, agg.VoteCounts["Cave"])
}

func TestVoteTracker_CastVote_InvalidOption(t *testing.T) {
	tracker := NewVoteTracker()
	voteID := tracker.CreateVote("session1", VoteTypePoll, "Question?", []string{"A", "B"})

	err := tracker.CastVote(voteID, "player1", "Alice", 5) // invalid index
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid option index")
}

func TestVoteTracker_CastVote_VoteClosed(t *testing.T) {
	tracker := NewVoteTracker()
	voteID := tracker.CreateVote("session1", VoteTypePoll, "Question?", []string{"A", "B"})

	_, err := tracker.CloseVote(voteID)
	require.NoError(t, err)

	err = tracker.CastVote(voteID, "player1", "Alice", 0)
	assert.Error(t, err)
	// Vote is deleted after close, so error is "not found"
	assert.Contains(t, err.Error(), "not found")
}

func TestVoteTracker_SetSentiment(t *testing.T) {
	tracker := NewVoteTracker()
	voteID := tracker.CreateVote("session1", VoteTypePoll, "Question?", []string{"A", "B"})

	err := tracker.CastVote(voteID, "player1", "Alice", 0)
	require.NoError(t, err)

	err = tracker.SetSentiment(voteID, "player1", false)
	require.NoError(t, err)

	// Sentiment history should be updated
	assert.Len(t, tracker.sentimentHistory, 1)
	assert.False(t, tracker.sentimentHistory[0])
}

func TestVoteTracker_CloseVote(t *testing.T) {
	tracker := NewVoteTracker()
	voteID := tracker.CreateVote("session1", VoteTypeInitiative, "Initiative order?", []string{"Alice", "Bob", "Carol"})

	tracker.CastVote(voteID, "p1", "Alice", 0)
	tracker.CastVote(voteID, "p2", "Bob", 1)
	tracker.CastVote(voteID, "p3", "Carol", 0) // Alice and Carol both vote for "Alice"

	agg, err := tracker.CloseVote(voteID)
	require.NoError(t, err)

	assert.True(t, agg.IsComplete)
	assert.Equal(t, 3, agg.TotalVotes)
	assert.Equal(t, 2, agg.VoteCounts["Alice"])
	assert.Equal(t, 1, agg.VoteCounts["Bob"])
	assert.Equal(t, "Alice", agg.WinningOption)
	assert.Greater(t, agg.Sentiment, 0.0) // Should have recorded sentiment

	// Vote should be removed from active votes
	_, err = tracker.GetAggregate(voteID)
	assert.Error(t, err)
}

func TestVoteTracker_History(t *testing.T) {
	tracker := NewVoteTracker()

	voteID1 := tracker.CreateVote("session1", VoteTypePoll, "Q1?", []string{"A", "B"})
	voteID2 := tracker.CreateVote("session1", VoteTypePoll, "Q2?", []string{"C", "D"})

	tracker.CastVote(voteID1, "p1", "Player1", 0)
	tracker.CastVote(voteID2, "p1", "Player1", 1)

	// Close votes to move them to historical record
	tracker.CloseVote(voteID1)
	tracker.CloseVote(voteID2)

	history := tracker.History("session1", 0)
	assert.Len(t, history, 2)
}

func TestVoteTracker_PendingVotes(t *testing.T) {
	tracker := NewVoteTracker()

	voteID1 := tracker.CreateVote("session1", VoteTypePoll, "Q1?", []string{"A", "B"})
	tracker.CreateVote("session1", VoteTypePoll, "Q2?", []string{"C", "D"})

	// Close first vote
	tracker.CloseVote(voteID1)

	pending := tracker.PendingVotes("session1")
	assert.Len(t, pending, 1)
	assert.Equal(t, "Q2?", pending[0].Question)
}

func TestVoteTracker_SessionSentiment(t *testing.T) {
	tracker := NewVoteTracker()

	voteID := tracker.CreateVote("session1", VoteTypePoll, "Question?", []string{"Yes", "No"})
	tracker.CastVote(voteID, "p1", "P1", 0)
	tracker.CastVote(voteID, "p2", "P2", 0)

	// Set sentiment for votes
	tracker.SetSentiment(voteID, "p1", true)
	tracker.SetSentiment(voteID, "p2", false)

	tracker.CloseVote(voteID)

	sentiment := tracker.SessionSentiment("session1")
	assert.InDelta(t, 0.5, sentiment, 0.01)
}

func TestVoteTracker_DisplaySentiment(t *testing.T) {
	tests := []struct {
		name     string
		history  []bool
		expected string
	}{
		{"no votes", []bool{}, "No votes yet"},
		{"all happy (80%+)", []bool{true, true, true, true, true}, "Players are happy!"},
		{"mostly happy (60-80%)", []bool{true, true, true, false}, "Players are content"},
		{"neutral (40-60%)", []bool{true, false, true, false}, "Players are neutral"},
		{"mostly unhappy (20-40%)", []bool{true, false, false, false}, "Players are frustrated"},
		{"unhappy (0-20%)", []bool{false, false, false, false}, "Players are unhappy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewVoteTracker()
			for _, s := range tt.history {
				tracker.sentimentHistory = append(tracker.sentimentHistory, s)
			}
			got := tracker.DisplaySentiment()
			assert.Equal(t, tt.expected, got)
		})
	}
}

// --- Personality Evolution Tests ---

func TestNewPersonality(t *testing.T) {
	p := NewPersonality()

	assert.Equal(t, 5.0, p.Strictness)
	assert.Equal(t, 5.0, p.Leniency)
	assert.Equal(t, 3.0, p.HumorLevel)
	assert.Equal(t, 4.0, p.AntarcticaReferences)
	assert.Equal(t, ModeByTheBook, p.CurrentMode)
	assert.Equal(t, 0, p.SessionsPlayed)
	assert.Equal(t, 0, p.TotalVotes)
}

func TestPersonality_Evolve(t *testing.T) {
	p := NewPersonality()

	// Initial state
	assert.Equal(t, 5.0, p.Strictness)
	assert.Equal(t, 5.0, p.Leniency)
	assert.Equal(t, 3.0, p.HumorLevel)
	assert.Equal(t, 0, p.TotalVotes)

	// Positive vote should increase humor and leniency
	p.Evolve(true)
	assert.Equal(t, 1, p.TotalVotes)
	assert.Greater(t, p.HumorLevel, 3.0)
	assert.Greater(t, p.Leniency, 5.0)

	// Negative vote should increase strictness
	strictnessAfterPos := p.Strictness
	p.Evolve(false)
	assert.Equal(t, 2, p.TotalVotes)
	assert.Greater(t, p.Strictness, strictnessAfterPos)
}

func TestPersonality_RollingSentiment(t *testing.T) {
	p := NewPersonality()

	// No history returns neutral
	assert.InDelta(t, 0.5, p.RollingSentiment(10), 0.01)

	// Add 7 positive, 3 negative out of 10
	for i := 0; i < 7; i++ {
		p.sentimentHistory = append(p.sentimentHistory, true)
	}
	for i := 0; i < 3; i++ {
		p.sentimentHistory = append(p.sentimentHistory, false)
	}

	// 7 positive out of 10 = 0.7
	sentiment := p.RollingSentiment(10)
	assert.Greater(t, sentiment, 0.6)
	assert.Less(t, sentiment, 0.8)
}

func TestPersonality_ShouldInjectAntarcticaJoke(t *testing.T) {
	p := NewPersonality()
	p.HumorLevel = 3.0 // Default

	// With HumorLevel 3, roughly 30% chance
	count := 0
	for i := 0; i < 100; i++ {
		p.TotalVotes = i
		if p.ShouldInjectAntarcticaJoke() {
			count++
		}
	}
	// Should be around 30% but allow for variance
	assert.Greater(t, count, 10)
	assert.Less(t, count, 50)
}

func TestPersonality_GetMood(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Personality)
		expected types.Mood
	}{
		{
			name: "cheerful when sentiment > 0.7",
			setup: func(p *Personality) {
				for i := 0; i < 8; i++ {
					p.sentimentHistory = append(p.sentimentHistory, true)
				}
				for i := 0; i < 2; i++ {
					p.sentimentHistory = append(p.sentimentHistory, false)
				}
			},
			expected: types.MoodCheerful,
		},
		{
			name: "whimsical when sentiment 0.6-0.7",
			setup: func(p *Personality) {
				for i := 0; i < 6; i++ {
					p.sentimentHistory = append(p.sentimentHistory, true)
				}
				for i := 0; i < 4; i++ {
					p.sentimentHistory = append(p.sentimentHistory, false)
				}
				// Need strictness > 7 to trigger this mood since sentiment = 0.6 is not > 0.6
				p.Strictness = 8.0
			},
			expected: types.MoodOminous,
		},
		{
			name: "melancholy when sentiment < 0.3",
			setup: func(p *Personality) {
				for i := 0; i < 2; i++ {
					p.sentimentHistory = append(p.sentimentHistory, true)
				}
				for i := 0; i < 8; i++ {
					p.sentimentHistory = append(p.sentimentHistory, false)
				}
			},
			expected: types.MoodMelancholy,
		},
		{
			name: "ominous when strictness > 7",
			setup: func(p *Personality) {
				p.Strictness = 8.0
				p.Leniency = 5.0
			},
			expected: types.MoodOminous,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPersonality()
			tt.setup(&p)
			assert.Equal(t, tt.expected, p.GetMood())
		})
	}
}

func TestPersonality_Merge(t *testing.T) {
	p1 := Personality{
		Strictness:           6.0,
		Leniency:             4.0,
		HumorLevel:           5.0,
		AntarcticaReferences: 7.0,
		SessionsPlayed:       3,
		TotalVotes:           50,
	}

	p2 := Personality{
		Strictness:           8.0,
		Leniency:             6.0,
		HumorLevel:           3.0,
		AntarcticaReferences: 5.0,
		SessionsPlayed:       2,
		TotalVotes:           30,
	}

	p1.Merge(p2)

	assert.Equal(t, 7.0, p1.Strictness)           // (6+8)/2
	assert.Equal(t, 5.0, p1.Leniency)             // (4+6)/2
	assert.Equal(t, 4.0, p1.HumorLevel)           // (5+3)/2
	assert.Equal(t, 6.0, p1.AntarcticaReferences) // (7+5)/2
	assert.Equal(t, 5, p1.SessionsPlayed)         // 3+2
	assert.Equal(t, 80, p1.TotalVotes)            // 50+30
}

func TestPersonality_updateMode(t *testing.T) {
	tests := []struct {
		name       string
		strictness float64
		leniency   float64
		expected   Mode
	}{
		{"by the book - strict > 7, leniency < 4", 8.0, 3.0, ModeByTheBook},
		{"narrator - leniency > 7, strict < 4", 3.0, 8.0, ModeNarrator},
		{"flexible - balanced", 5.0, 5.0, ModeFlexible},
		{"flexible - high strict, high leniency", 8.0, 8.0, ModeFlexible},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPersonality()
			p.Strictness = tt.strictness
			p.Leniency = tt.leniency
			p.updateMode()
			assert.Equal(t, tt.expected, p.CurrentMode)
		})
	}
}
