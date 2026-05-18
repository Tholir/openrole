// Package dm provides the Dungeon Master engine for OPENROLE.
package dm

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

// VoteType categorizes different voting scenarios.
type VoteType int

const (
	// VoteTypeInitiative is for combat initiative rolls.
	VoteTypeInitiative VoteType = iota
	// VoteTypeGroupDecision is for party-wide decisions.
	VoteTypeGroupDecision
	// VoteTypeContest is for opposed checks between characters/NPCs.
	VoteTypeContest
	// VoteTypePoll is for general player polls.
	VoteTypePoll
)

// Vote represents a single vote with sentiment tracking.
type Vote struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	PlayerID   string    `json:"player_id"`
	PlayerName string    `json:"player_name"`
	VoteType   VoteType  `json:"vote_type"`
	Option     string    `json:"option"` // what they voted for
	OptionIdx  int       `json:"option_idx"`
	Sentiment  bool      `json:"sentiment"` // true=positive, false=negative
	CreatedAt  time.Time `json:"created_at"`
}

// VoteAggregate summarizes votes for a question.
type VoteAggregate struct {
	Question       string         `json:"question"`
	TotalVotes     int            `json:"total_votes"`
	Options        []string       `json:"options"`
	VoteCounts     map[string]int `json:"vote_counts"` // option -> count
	Sentiment      float64        `json:"sentiment"`   // 0-1 positive ratio
	SentimentCount int            `json:"sentiment_count"`
	WinningOption  string         `json:"winning_option"`
	IsComplete     bool           `json:"is_complete"`
}

// VoteTracker manages all votes and provides aggregation.
type VoteTracker struct {
	// activeVotes tracks currently open votes by ID.
	activeVotes map[string]*ActiveVote

	// historicalVotes stores completed votes.
	historicalVotes []*Vote

	// sentimentHistory tracks overall sentiment trend.
	sentimentHistory []bool
}

// ActiveVote wraps a vote with tracking metadata.
type ActiveVote struct {
	ID        string
	SessionID string
	VoteType  VoteType
	Question  string
	Options   []string
	Votes     map[string]Vote // playerID -> Vote
	CreatedAt time.Time
	ClosedAt  *time.Time
}

// NewVoteTracker creates a fresh vote tracker.
func NewVoteTracker() *VoteTracker {
	return &VoteTracker{
		activeVotes:      make(map[string]*ActiveVote),
		historicalVotes:  make([]*Vote, 0, 100),
		sentimentHistory: make([]bool, 0, 50),
	}
}

// CreateVote creates a new active vote for a session.
func (t *VoteTracker) CreateVote(sessionID string, voteType VoteType, question string, options []string) string {
	id := uuid.New().String()
	av := &ActiveVote{
		ID:        id,
		SessionID: sessionID,
		VoteType:  voteType,
		Question:  question,
		Options:   options,
		Votes:     make(map[string]Vote),
		CreatedAt: time.Now(),
	}
	t.activeVotes[id] = av
	return id
}

// CastVote records a player's vote on an active vote.
func (t *VoteTracker) CastVote(voteID string, playerID, playerName string, optionIdx int) error {
	av, exists := t.activeVotes[voteID]
	if !exists {
		return fmt.Errorf("vote %s not found", voteID)
	}

	if av.ClosedAt != nil {
		return fmt.Errorf("vote %s is closed", voteID)
	}

	if optionIdx < 0 || optionIdx >= len(av.Options) {
		return fmt.Errorf("invalid option index %d", optionIdx)
	}

	vote := Vote{
		ID:         uuid.New().String(),
		SessionID:  av.SessionID,
		PlayerID:   playerID,
		PlayerName: playerName,
		VoteType:   av.VoteType,
		Option:     av.Options[optionIdx],
		OptionIdx:  optionIdx,
		Sentiment:  true, // default to positive, can be adjusted
		CreatedAt:  time.Now(),
	}

	av.Votes[playerID] = vote
	return nil
}

// SetSentiment updates a vote's sentiment after the fact.
func (t *VoteTracker) SetSentiment(voteID string, playerID string, positive bool) error {
	av, exists := t.activeVotes[voteID]
	if !exists {
		return fmt.Errorf("vote %s not found", voteID)
	}

	vote, exists := av.Votes[playerID]
	if !exists {
		return fmt.Errorf("player %s has not voted on %s", playerID, voteID)
	}

	vote.Sentiment = positive
	av.Votes[playerID] = vote

	// Track sentiment for personality evolution
	t.sentimentHistory = append(t.sentimentHistory, positive)
	if len(t.sentimentHistory) > 50 {
		t.sentimentHistory = t.sentimentHistory[1:]
	}

	return nil
}

// CloseVote ends voting and returns the aggregate results.
func (t *VoteTracker) CloseVote(voteID string) (*VoteAggregate, error) {
	av, exists := t.activeVotes[voteID]
	if !exists {
		return nil, fmt.Errorf("vote %s not found", voteID)
	}

	now := time.Now()
	av.ClosedAt = &now

	// Build aggregate
	agg := &VoteAggregate{
		Question:   av.Question,
		TotalVotes: len(av.Votes),
		Options:    av.Options,
		VoteCounts: make(map[string]int),
		IsComplete: true,
	}

	for _, opt := range av.Options {
		agg.VoteCounts[opt] = 0
	}

	sentimentPositive := 0
	for _, vote := range av.Votes {
		agg.VoteCounts[vote.Option]++
		if vote.Sentiment {
			sentimentPositive++
		}
		t.historicalVotes = append(t.historicalVotes, &vote)
	}

	if len(av.Votes) > 0 {
		agg.Sentiment = float64(sentimentPositive) / float64(len(av.Votes))
	}
	agg.SentimentCount = sentimentPositive

	// Determine winner
	maxCount := 0
	for opt, count := range agg.VoteCounts {
		if count > maxCount {
			maxCount = count
			agg.WinningOption = opt
		}
	}

	delete(t.activeVotes, voteID)
	return agg, nil
}

// GetAggregate returns current vote counts without closing.
func (t *VoteTracker) GetAggregate(voteID string) (*VoteAggregate, error) {
	av, exists := t.activeVotes[voteID]
	if !exists {
		return nil, fmt.Errorf("vote %s not found", voteID)
	}

	agg := &VoteAggregate{
		Question:   av.Question,
		TotalVotes: len(av.Votes),
		Options:    av.Options,
		VoteCounts: make(map[string]int),
		IsComplete: false,
	}

	for _, opt := range av.Options {
		agg.VoteCounts[opt] = 0
	}

	for _, vote := range av.Votes {
		agg.VoteCounts[vote.Option]++
	}

	return agg, nil
}

// DisplaySentiment returns a human-readable sentiment indicator.
func (t *VoteTracker) DisplaySentiment() string {
	if len(t.sentimentHistory) == 0 {
		return "No votes yet"
	}

	positive := 0
	for _, s := range t.sentimentHistory {
		if s {
			positive++
		}
	}
	ratio := float64(positive) / float64(len(t.sentimentHistory))

	switch {
	case ratio >= 0.8:
		return "Players are happy!"
	case ratio >= 0.6:
		return "Players are content"
	case ratio >= 0.4:
		return "Players are neutral"
	case ratio >= 0.2:
		return "Players are frustrated"
	default:
		return "Players are unhappy"
	}
}

// History returns the vote history for a session.
func (t *VoteTracker) History(sessionID string, limit int) []*Vote {
	var result []*Vote
	for _, v := range t.historicalVotes {
		if v.SessionID == sessionID {
			result = append(result, v)
		}
	}

	// Sort by created_at descending (most recent first)
	slices.SortFunc(result, func(a, b *Vote) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	if limit > 0 && len(result) > limit {
		return result[:limit]
	}
	return result
}

// PendingVotes returns all open votes for a session.
func (t *VoteTracker) PendingVotes(sessionID string) []*ActiveVote {
	var result []*ActiveVote
	for _, av := range t.activeVotes {
		if av.SessionID == sessionID && av.ClosedAt == nil {
			result = append(result, av)
		}
	}
	return result
}

// SessionSentiment returns the rolling sentiment for a session.
func (t *VoteTracker) SessionSentiment(sessionID string) float64 {
	var votes []*Vote
	for _, v := range t.historicalVotes {
		if v.SessionID == sessionID {
			votes = append(votes, v)
		}
	}

	if len(votes) == 0 {
		return 0.5 // neutral
	}

	positive := 0
	for _, v := range votes {
		if v.Sentiment {
			positive++
		}
	}
	return float64(positive) / float64(len(votes))
}
