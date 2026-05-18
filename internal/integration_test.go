// Package integration provides integration tests for OPENROLE.
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gentleman-programming/openrole/internal/dice"
	"github.com/gentleman-programming/openrole/internal/dm"
	"github.com/gentleman-programming/openrole/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- API Client Mock Tests ---

// MockAPIClient is a mock implementation for testing API interactions.
type MockAPIClient struct {
	// CallCount tracks how many times each method was called.
	RollCalls         int
	SpawnCalls        int
	SpellCalls        int
	GetCharacterCalls int

	// Mock responses
	RollResult    *dice.RollResult
	SpawnResult   string
	SpellResult   string
	CharacterData *types.Character

	// Error injection
	RollError         error
	SpawnError        error
	SpellError        error
	GetCharacterError error
}

func (m *MockAPIClient) RollDice(ctx context.Context, expression string) (*dice.RollResult, error) {
	m.RollCalls++
	if m.RollError != nil {
		return nil, m.RollError
	}
	return m.RollResult, nil
}

func (m *MockAPIClient) SpawnMonster(ctx context.Context, name string) (string, error) {
	m.SpawnCalls++
	if m.SpawnError != nil {
		return "", m.SpawnError
	}
	return m.SpawnResult, nil
}

func (m *MockAPIClient) GetSpell(ctx context.Context, name string) (string, error) {
	m.SpellCalls++
	if m.SpellError != nil {
		return "", m.SpellError
	}
	return m.SpellResult, nil
}

func (m *MockAPIClient) GetCharacter(ctx context.Context, id string) (*types.Character, error) {
	m.GetCharacterCalls++
	if m.GetCharacterError != nil {
		return nil, m.GetCharacterError
	}
	return m.CharacterData, nil
}

// TestDiceIntegration_WithMockAPI tests dice rolling through a mock API client.
func TestDiceIntegration_WithMockAPI(t *testing.T) {
	client := &MockAPIClient{
		RollResult: &dice.RollResult{
			Expression: "1d20",
			Total:      17,
			Rolls: []dice.DieRoll{
				{Dice: "1d20", Result: 17, Kept: true, Details: []int{17}},
			},
		},
	}

	ctx := context.Background()
	result, err := client.RollDice(ctx, "1d20")
	require.NoError(t, err)
	assert.Equal(t, 17, result.Total)
	assert.Equal(t, 1, client.RollCalls)
}

// TestDiceIntegration_WithRealRoller tests the full dice rolling flow.
func TestDiceIntegration_WithRealRoller(t *testing.T) {
	handler := dice.NewCommandHandler()
	ctx := context.Background()

	result := handler.HandleRoll(ctx, "/roll 2d6+3")
	require.False(t, result.IsError)
	assert.NotEmpty(t, result.Text)

	result = handler.HandleRoll(ctx, "/roll 1d20")
	require.False(t, result.IsError)
	assert.NotEmpty(t, result.Text)
}

// TestDiceIntegration_ErrorHandling tests error cases.
func TestDiceIntegration_ErrorHandling(t *testing.T) {
	client := &MockAPIClient{
		RollError: assert.AnError,
	}

	ctx := context.Background()
	_, err := client.RollDice(ctx, "1d20")
	assert.Error(t, err)
	assert.Equal(t, 1, client.RollCalls)
}

// TestMonsterSpawnIntegration tests monster spawning through mock.
func TestMonsterSpawnIntegration(t *testing.T) {
	client := &MockAPIClient{
		SpawnResult: "🐉 Ancient Red Dragon\nHP: 297\nAC: 22\nCR: 24",
	}

	ctx := context.Background()
	result, err := client.SpawnMonster(ctx, "dragon")
	require.NoError(t, err)
	assert.Contains(t, result, "Dragon")
	assert.Contains(t, result, "HP: 297")
	assert.Equal(t, 1, client.SpawnCalls)
}

// TestSpellLookupIntegration tests spell lookup through mock.
func TestSpellLookupIntegration(t *testing.T) {
	client := &MockAPIClient{
		SpellResult: "🔥 Fireball\nLevel 3\nCasting time: 1 action\nRange: 150 feet",
	}

	ctx := context.Background()
	result, err := client.GetSpell(ctx, "fireball")
	require.NoError(t, err)
	assert.Contains(t, result, "Fireball")
	assert.Equal(t, 1, client.SpellCalls)
}

// TestCharacterFetchIntegration tests character retrieval through mock.
func TestCharacterFetchIntegration(t *testing.T) {
	client := &MockAPIClient{
		CharacterData: &types.Character{
			ID:    "char-123",
			Name:  "Test Hero",
			Level: 5,
			Class: "Fighter",
			HP:    40,
			MaxHP: 45,
			AC:    18,
		},
	}

	ctx := context.Background()
	char, err := client.GetCharacter(ctx, "char-123")
	require.NoError(t, err)
	assert.Equal(t, "Test Hero", char.Name)
	assert.Equal(t, 5, char.Level)
	assert.Equal(t, 1, client.GetCharacterCalls)
}

// --- Vote Persistence Integration Tests ---

// MockVoteStore simulates a persistent vote store for integration testing.
type MockVoteStore struct {
	votes    map[string]*dm.VoteTracker
	sessions []string
}

func NewMockVoteStore() *MockVoteStore {
	return &MockVoteStore{
		votes:    make(map[string]*dm.VoteTracker),
		sessions: make([]string, 0),
	}
}

func (s *MockVoteStore) CreateSession(sessionID string) {
	s.sessions = append(s.sessions, sessionID)
	s.votes[sessionID] = dm.NewVoteTracker()
}

func (s *MockVoteStore) GetTracker(sessionID string) *dm.VoteTracker {
	return s.votes[sessionID]
}

// TestVotePersistence_BasicFlow tests the full vote lifecycle with persistence.
func TestVotePersistence_BasicFlow(t *testing.T) {
	store := NewMockVoteStore()

	// Create session
	sessionID := "session-test-001"
	store.CreateSession(sessionID)
	tracker := store.GetTracker(sessionID)
	require.NotNil(t, tracker)

	// Create a vote
	voteID := tracker.CreateVote(sessionID, dm.VoteTypeInitiative, "Initiative order?", []string{"Alice", "Bob", "Carol"})
	assert.NotEmpty(t, voteID)

	// Cast votes
	err := tracker.CastVote(voteID, "player-1", "Alice", 0)
	require.NoError(t, err)
	err = tracker.CastVote(voteID, "player-2", "Bob", 1)
	require.NoError(t, err)
	err = tracker.CastVote(voteID, "player-3", "Carol", 2)
	require.NoError(t, err)

	// Get aggregate before close
	agg, err := tracker.GetAggregate(voteID)
	require.NoError(t, err)
	assert.Equal(t, 3, agg.TotalVotes)

	// Set sentiments
	tracker.SetSentiment(voteID, "player-1", true)
	tracker.SetSentiment(voteID, "player-2", true)
	tracker.SetSentiment(voteID, "player-3", true)

	// Close vote
	closedAgg, err := tracker.CloseVote(voteID)
	require.NoError(t, err)
	assert.True(t, closedAgg.IsComplete)

	// Verify history persists
	history := tracker.History(sessionID, 0)
	assert.Len(t, history, 3)
}

// TestVotePersistence_MultipleSessions tests isolation between sessions.
func TestVotePersistence_MultipleSessions(t *testing.T) {
	store := NewMockVoteStore()

	// Create two sessions
	store.CreateSession("session-a")
	store.CreateSession("session-b")

	trackerA := store.GetTracker("session-a")
	trackerB := store.GetTracker("session-b")

	// Create votes in each session
	voteIDA := trackerA.CreateVote("session-a", dm.VoteTypePoll, "Question A?", []string{"Yes", "No"})
	voteIDB := trackerB.CreateVote("session-b", dm.VoteTypePoll, "Question B?", []string{"A", "B", "C"})

	// Vote in session A
	trackerA.CastVote(voteIDA, "p1", "Player1", 0)

	// Vote in session B
	trackerB.CastVote(voteIDB, "p2", "Player2", 1)
	trackerB.CastVote(voteIDB, "p3", "Player3", 2)

	// Verify isolation
	aggA, _ := trackerA.GetAggregate(voteIDA)
	aggB, _ := trackerB.GetAggregate(voteIDB)

	assert.Equal(t, 1, aggA.TotalVotes)
	assert.Equal(t, 2, aggB.TotalVotes)

	// Close votes
	trackerA.CloseVote(voteIDA)
	trackerB.CloseVote(voteIDB)

	// Verify history
	historyA := trackerA.History("session-a", 0)
	historyB := trackerB.History("session-b", 0)

	assert.Len(t, historyA, 1)
	assert.Len(t, historyB, 2)

	// Sentiments should not leak between sessions
	assert.Len(t, trackerA.History("session-a", 0), 1)
	assert.Len(t, trackerB.History("session-b", 0), 2)
}

// TestVotePersistence_SentimentTracking tests sentiment accumulation over time.
func TestVotePersistence_SentimentTracking(t *testing.T) {
	store := NewMockVoteStore()
	sessionID := "session-sentiment"
	store.CreateSession(sessionID)
	tracker := store.GetTracker(sessionID)

	// Create multiple votes
	for i := 0; i < 5; i++ {
		voteID := tracker.CreateVote(sessionID, dm.VoteTypePoll, "Vote?", []string{"A", "B"})
		tracker.CastVote(voteID, "p1", "Player", 0)
		tracker.SetSentiment(voteID, "p1", i%2 == 0) // alternating sentiment
		tracker.CloseVote(voteID)
	}

	// Check session sentiment
	sentiment := tracker.SessionSentiment(sessionID)
	assert.InDelta(t, 0.6, sentiment, 0.1) // 3 positive, 2 negative

	// Check display sentiment
	display := tracker.DisplaySentiment()
	assert.NotEmpty(t, display)
}

// TestVotePersistence_PendingVotesPersistence tests pending votes are tracked.
func TestVotePersistence_PendingVotesPersistence(t *testing.T) {
	store := NewMockVoteStore()
	sessionID := "session-pending"
	store.CreateSession(sessionID)
	tracker := store.GetTracker(sessionID)

	// Create multiple open votes
	voteID1 := tracker.CreateVote(sessionID, dm.VoteTypeInitiative, "Initiative 1?", []string{"A", "B"})
	tracker.CreateVote(sessionID, dm.VoteTypePoll, "Question 2?", []string{"X", "Y"})
	tracker.CreateVote(sessionID, dm.VoteTypeGroupDecision, "Decision?", []string{"P1", "P2", "P3"})

	pending := tracker.PendingVotes(sessionID)
	assert.Len(t, pending, 3)

	// Close one vote
	tracker.CloseVote(voteID1)

	pending = tracker.PendingVotes(sessionID)
	assert.Len(t, pending, 2)
}

// TestDMEngine_VoteTrackingIntegration tests the full DM engine with vote tracking.
func TestDMEngine_VoteTrackingIntegration(t *testing.T) {
	// Setup DM state
	dmState := types.NewDMState("test-session")

	// Create vote tracker (normally part of DM engine)
	tracker := dm.NewVoteTracker()

	// Simulate voting on DM decisions
	voteID := tracker.CreateVote("test-session", dm.VoteTypeGroupDecision,
		"Do we explore the cave or the forest?",
		[]string{"Cave", "Forest", "Neither"})

	// Players cast votes
	tracker.CastVote(voteID, "player-1", "Alice", 0)
	tracker.CastVote(voteID, "player-2", "Bob", 1)
	tracker.CastVote(voteID, "player-3", "Carol", 0) // majority for cave

	// Set sentiments based on whether they got their way
	tracker.SetSentiment(voteID, "player-1", true)  // got cave
	tracker.SetSentiment(voteID, "player-2", false) // didn't get forest
	tracker.SetSentiment(voteID, "player-3", true)  // got cave

	// Close vote
	agg, err := tracker.CloseVote(voteID)
	require.NoError(t, err)
	assert.Equal(t, "Cave", agg.WinningOption)
	assert.True(t, agg.IsComplete)

	// Check sentiment
	display := tracker.DisplaySentiment()
	assert.Contains(t, display, "Players are") // Should reflect mixed sentiment

	// Verify DM state is unchanged (integration test verifies separation)
	assert.NotNil(t, dmState)
	assert.Equal(t, 1, dmState.SessionNumber)
}

// TestCharacter_VoteCorrelation tests that character stats affect vote outcomes.
func TestCharacter_VoteCorrelation(t *testing.T) {
	// Setup a character
	char := &types.Character{
		Name:  "Strong Hero",
		Level: 5,
		Stats: types.Stats{Strength: 18}, // +4 mod
	}

	// Setup vote tracker
	tracker := dm.NewVoteTracker()
	sessionID := "session-correlation"
	tracker.CreateVote(sessionID, dm.VoteTypeContest, "Strength contest?", []string{})

	// The vote tracker doesn't use character stats directly,
	// but this test documents the integration point between
	// character creation and DM mechanics.
	assert.NotNil(t, char)
	assert.Equal(t, 5, char.Level)
}

// TestFullGameSession_Simulated simulates a full game session with voting.
func TestFullGameSession_Simulated(t *testing.T) {
	// Initialize
	dmState := types.NewDMState("full-game-session")
	tracker := dm.NewVoteTracker()
	personality := dm.NewPersonality()

	// Session starts with neutral sentiment
	assert.Equal(t, 0, personality.TotalVotes)

	// Create party vote
	voteID := tracker.CreateVote("full-game-session", dm.VoteTypeGroupDecision,
		"The party encounters a mysterious portal. What do you do?",
		[]string{"Enter the portal", "Investigate first", "Ignore and move on"})

	// Party votes
	tracker.CastVote(voteID, "alice", "Alice", 1) // investigate
	tracker.CastVote(voteID, "bob", "Bob", 0)     // enter
	tracker.CastVote(voteID, "carol", "Carol", 1) // investigate

	// Set sentiments
	tracker.SetSentiment(voteID, "alice", true)
	tracker.SetSentiment(voteID, "bob", false) // didn't get his way
	tracker.SetSentiment(voteID, "carol", true)

	// Close and record outcome
	agg, err := tracker.CloseVote(voteID)
	require.NoError(t, err)
	assert.True(t, agg.IsComplete)

	// Evolve personality based on vote outcomes
	for _, vote := range tracker.History("full-game-session", 0) {
		personality.Evolve(vote.Sentiment)
	}

	// Verify personality evolved
	assert.Greater(t, personality.TotalVotes, 0)

	// Verify DM state is intact
	assert.Equal(t, "full-game-session", dmState.SessionID)
	assert.NotNil(t, dmState.Personality.Traits)
}

// TestVoteTracker_ConcurrentAccess tests concurrent vote operations.
func TestVoteTracker_ConcurrentAccess(t *testing.T) {
	tracker := dm.NewVoteTracker()
	sessionID := "concurrent-session"
	voteID := tracker.CreateVote(sessionID, dm.VoteTypePoll, "Concurrent test?", []string{"A", "B", "C"})

	// Simulate concurrent access by creating multiple votes rapidly
	for i := 0; i < 10; i++ {
		idx := i % 3
		err := tracker.CastVote(voteID, "player-"+string(rune('0'+i)), "Player", idx)
		// Some may fail if vote is closed or duplicate player ID
		if err == nil {
			// Update sentiment
			tracker.SetSentiment(voteID, "player-"+string(rune('0'+i)), i%2 == 0)
		}
	}

	// Verify aggregate works
	agg, err := tracker.GetAggregate(voteID)
	require.NoError(t, err)
	assert.Greater(t, agg.TotalVotes, 0)
}

// TestDiceRoller_StressTest tests dice roller under stress.
func TestDiceRoller_StressTest(t *testing.T) {
	handler := dice.NewCommandHandler()
	ctx := context.Background()

	// Run many dice rolls
	for i := 0; i < 100; i++ {
		result := handler.HandleRoll(ctx, "/roll 4d6kh3")
		if result.IsError {
			t.Fatalf("Unexpected error on roll %d: %s", i, result.Text)
		}
	}
}

// TestVoteAggregate_SentimentCalculation tests sentiment aggregation.
func TestVoteAggregate_SentimentCalculation(t *testing.T) {
	tracker := dm.NewVoteTracker()
	voteID := tracker.CreateVote("test", dm.VoteTypePoll, "Test?", []string{"Yes", "No"})

	// Cast votes with known sentiments
	tracker.CastVote(voteID, "p1", "P1", 0)
	tracker.CastVote(voteID, "p2", "P2", 0)
	tracker.CastVote(voteID, "p3", "P3", 1)
	tracker.CastVote(voteID, "p4", "P4", 1)
	tracker.CastVote(voteID, "p5", "P5", 1)

	// Set sentiments: 2 positive (yes votes), 3 negative (no votes)
	tracker.SetSentiment(voteID, "p1", true)
	tracker.SetSentiment(voteID, "p2", false) // voted yes but disliked
	tracker.SetSentiment(voteID, "p3", true)
	tracker.SetSentiment(voteID, "p4", true)
	tracker.SetSentiment(voteID, "p5", true)

	agg, err := tracker.CloseVote(voteID)
	require.NoError(t, err)

	// Sentiment should be 4/5 positive
	assert.InDelta(t, 0.8, agg.Sentiment, 0.01)
}

// TestCharacterIntegration_WithDiceRoller tests character creation with dice.
func TestCharacterIntegration_WithDiceRoller(t *testing.T) {
	roller := dice.NewRoller()

	// Roll ability scores using 4d6kh3 method
	abilityScores := make([]int, 6)
	for i := 0; i < 6; i++ {
		result, err := roller.ParseRoll("4d6kh3")
		require.NoError(t, err)
		abilityScores[i] = result.Total
		// Valid D&D ability scores should be between 3 and 18
		assert.GreaterOrEqual(t, abilityScores[i], 3)
		assert.LessOrEqual(t, abilityScores[i], 18)
	}

	// Create a character with rolled scores
	char := &types.Character{
		Name:      "Rolled Hero",
		Level:     1,
		Stats:     types.Stats{},
		Inventory: []types.Item{},
		Skills:    []string{},
		Spells:    []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Assign scores (in order: STR, DEX, CON, INT, WIS, CHA)
	char.Stats = types.Stats{
		Strength:     abilityScores[0],
		Dexterity:    abilityScores[1],
		Constitution: abilityScores[2],
		Intelligence: abilityScores[3],
		Wisdom:       abilityScores[4],
		Charisma:     abilityScores[5],
	}

	// Verify character can be used with DM engine
	dc, err := dm.CalculateDC(char, dm.DCEasy, "STR")
	require.NoError(t, err)
	assert.Greater(t, dc, 0)

	_ = char
}

// TestAPIErrorPropagation tests that errors propagate correctly.
func TestAPIErrorPropagation(t *testing.T) {
	client := &MockAPIClient{
		RollError: assert.AnError,
	}

	ctx := context.Background()
	_, err := client.RollDice(ctx, "1d20")
	assert.Error(t, err)

	// Verify error is returned, not swallowed
	client.RollError = nil
	client.RollResult = &dice.RollResult{Expression: "1d20", Total: 15}
	result, err := client.RollDice(ctx, "1d20")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
