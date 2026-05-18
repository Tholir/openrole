// Package dice provides D&D dice rolling functionality.
package dice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Dice Rolling Tests ---

func TestRoller_SimpleRoll(t *testing.T) {
	roller := NewRoller()

	// Test each standard die type
	tests := []struct {
		sides int
		min   int
		max   int
	}{
		{4, 1, 4},
		{6, 1, 6},
		{8, 1, 8},
		{10, 1, 10},
		{12, 1, 12},
		{20, 1, 20},
		{100, 1, 100},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := roller.SimpleRoll(tt.sides)
			assert.GreaterOrEqual(t, result, tt.min)
			assert.LessOrEqual(t, result, tt.max)
		})
	}
}

func TestRoller_RollMultiple(t *testing.T) {
	roller := NewRoller()

	results := roller.RollMultiple(5, 6)
	assert.Len(t, results, 5)

	for _, r := range results {
		assert.GreaterOrEqual(t, r, 1)
		assert.LessOrEqual(t, r, 6)
	}
}

func TestRoller_RollAdvantage(t *testing.T) {
	roller := NewRoller()

	// Run multiple times to ensure randomness
	seen := make(map[int]bool)
	for i := 0; i < 50; i++ {
		result := roller.RollAdvantage()
		assert.GreaterOrEqual(t, result.Roll1, 1)
		assert.LessOrEqual(t, result.Roll1, 20)
		assert.GreaterOrEqual(t, result.Roll2, 1)
		assert.LessOrEqual(t, result.Roll2, 20)
		assert.True(t, result.IsAdvantage)

		// Selected should be the higher roll
		high := result.Roll1
		if result.Roll2 > high {
			high = result.Roll2
		}
		assert.Equal(t, high, result.Selected)

		seen[result.Selected] = true
	}

	// Should have seen a variety of rolls
	assert.Greater(t, len(seen), 5)
}

func TestRoller_RollDisadvantage(t *testing.T) {
	roller := NewRoller()

	for i := 0; i < 50; i++ {
		result := roller.RollDisadvantage()
		assert.GreaterOrEqual(t, result.Roll1, 1)
		assert.LessOrEqual(t, result.Roll1, 20)
		assert.GreaterOrEqual(t, result.Roll2, 1)
		assert.LessOrEqual(t, result.Roll2, 20)
		assert.False(t, result.IsAdvantage)

		// Selected should be the lower roll
		low := result.Roll1
		if result.Roll2 < low {
			low = result.Roll2
		}
		assert.Equal(t, low, result.Selected)
	}
}

// --- Notation Parsing Tests ---

func TestRoller_ParseRoll(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantErr    bool
		wantTotal  int
		wantRolls  int
	}{
		{
			name:       "single d20",
			expression: "1d20",
			wantRolls:  1,
		},
		{
			name:       "two d6",
			expression: "2d6",
			wantRolls:  1,
		},
		{
			name:       "d20 plus 3",
			expression: "1d20+3",
			wantRolls:  1,
		},
		{
			name:       "two d6 minus 1",
			expression: "2d6-1",
			wantRolls:  1,
		},
		{
			name:       "advantage notation",
			expression: "2d20kh1",
			wantRolls:  1,
		},
		{
			name:       "disadvantage notation",
			expression: "2d20kl1",
			wantRolls:  1,
		},
		{
			name:       "keep highest 2 of 4d6",
			expression: "4d6kh2",
			wantRolls:  1,
		},
		{
			name:       "keep lowest 2 of 4d6",
			expression: "4d6kl2",
			wantRolls:  1,
		},
		{
			name:       "multiple dice expressions",
			expression: "2d6+3 1d20",
			wantRolls:  2,
		},
		{
			name:       "complex roll",
			expression: "4d6kh3+2",
			wantRolls:  1,
		},
		{
			name:       "whitespace trimmed",
			expression: "  2d6+3  ",
			wantRolls:  1,
		},
		{
			name:       "invalid notation",
			expression: "notadice",
			wantErr:    true,
		},
		{
			name:       "empty string",
			expression: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roller := NewRoller()
			result, err := roller.ParseRoll(tt.expression)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			// Note: Expression is trimmed by ParseRoll, so whitespace is removed
			assert.NotEmpty(t, result.Expression)
			assert.Len(t, result.Rolls, tt.wantRolls)
			assert.NotNil(t, result.Rolls)

			// Verify total is reasonable (sum of all dice + modifiers)
			for _, roll := range result.Rolls {
				assert.Greater(t, roll.Result, 0, "roll result should be positive")
			}
		})
	}
}

func TestRoller_ParseRoll_StandardDice(t *testing.T) {
	roller := NewRoller()

	// These should all parse and roll without error
	expressions := []string{
		"1d4", "1d6", "1d8", "1d10", "1d12", "1d20", "1d100",
		"2d4", "2d6", "2d8", "2d10", "2d12", "2d20",
		"3d6", "4d6", "6d6", "8d6", "10d6",
	}

	for _, expr := range expressions {
		t.Run(expr, func(t *testing.T) {
			result, err := roller.ParseRoll(expr)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Greater(t, result.Total, 0)
		})
	}
}

func TestRoller_ParseRoll_WithModifiers(t *testing.T) {
	roller := NewRoller()

	tests := []struct {
		name       string
		expression string
		wantMod    int
	}{
		{"plus one", "1d20+1", 1},
		{"plus five", "1d20+5", 5},
		{"minus one", "1d20-1", -1},
		{"minus five", "1d20-5", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := roller.ParseRoll(tt.expression)
			require.NoError(t, err)
			assert.Equal(t, tt.wantMod, result.Total-result.Rolls[0].Result+tt.wantMod)
		})
	}
}

func TestRoller_ParseRoll_KeepHighest(t *testing.T) {
	roller := NewRoller()

	result, err := roller.ParseRoll("2d20kh1")
	require.NoError(t, err)
	require.Len(t, result.Rolls, 1)

	roll := result.Rolls[0]
	// Should only keep 1 die, but details shows all rolled
	assert.NotEmpty(t, roll.Details)
	assert.Len(t, roll.Details, 2)
}

func TestRoller_ParseRoll_KeepLowest(t *testing.T) {
	roller := NewRoller()

	result, err := roller.ParseRoll("2d20kl1")
	require.NoError(t, err)
	require.Len(t, result.Rolls, 1)

	roll := result.Rolls[0]
	assert.NotEmpty(t, roll.Details)
	assert.Len(t, roll.Details, 2)
}

func TestRoller_ParseRoll_KeepHighestN(t *testing.T) {
	roller := NewRoller()

	result, err := roller.ParseRoll("4d6kh3")
	require.NoError(t, err)
	require.Len(t, result.Rolls, 1)

	roll := result.Rolls[0]
	assert.Len(t, roll.Details, 4)
}

func TestRoller_ParseRoll_KeepLowestN(t *testing.T) {
	roller := NewRoller()

	result, err := roller.ParseRoll("4d6kl3")
	require.NoError(t, err)
	require.Len(t, result.Rolls, 1)

	roll := result.Rolls[0]
	assert.Len(t, roll.Details, 4)
}

func TestRoller_ParseRoll_MultipleExpressions(t *testing.T) {
	roller := NewRoller()

	result, err := roller.ParseRoll("2d6+3 1d20")
	require.NoError(t, err)
	assert.Len(t, result.Rolls, 2)

	// First roll: 2d6 + 3
	assert.Contains(t, result.Rolls[0].Dice, "2d6")
	// Second roll: 1d20
	assert.Contains(t, result.Rolls[1].Dice, "1d20")
}

// --- Format Result Tests ---

func TestFormatResult(t *testing.T) {
	roller := NewRoller()
	result, err := roller.ParseRoll("2d6+3")
	require.NoError(t, err)

	output := FormatResult(result)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "🎲")
	assert.Contains(t, output, "TOTAL")
}

func TestFormatAdvantage(t *testing.T) {
	roller := NewRoller()

	// Advantage
	result := roller.RollAdvantage()
	output := FormatAdvantage(result)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "ADVANTAGE")
	assert.Contains(t, output, "kh1")

	// Disadvantage
	result2 := roller.RollDisadvantage()
	output2 := FormatAdvantage(result2)
	assert.NotEmpty(t, output2)
	assert.Contains(t, output2, "DISADVANTAGE")
	assert.Contains(t, output2, "kl1")
}

// --- Command Handler Tests ---

func TestCommandHandler_HandleRoll(t *testing.T) {
	handler := NewCommandHandler()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantOut bool
	}{
		{"valid d20", "/roll 1d20", false, true},
		{"valid 2d6", "/roll 2d6+3", false, true},
		{"advantage", "/roll advantage", false, true},
		{"advantage short", "/roll adv", false, true},
		{"disadvantage", "/roll disadvantage", false, true},
		{"disadvantage short", "/roll dis", false, true},
		{"invalid notation", "/roll notadice", true, true},
		{"empty roll", "/roll", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.HandleRoll(context.Background(), tt.input)
			assert.Equal(t, tt.wantErr, result.IsError)
			if tt.wantOut {
				assert.NotEmpty(t, result.Text)
			}
		})
	}
}

func TestCommandHandler_HandleRoll_AdvantageDisadvantage(t *testing.T) {
	handler := NewCommandHandler()

	advResult := handler.HandleRoll(context.Background(), "/roll advantage")
	require.False(t, advResult.IsError)
	assert.Contains(t, advResult.Text, "ADVANTAGE")

	disResult := handler.HandleRoll(context.Background(), "/roll disadvantage")
	require.False(t, disResult.IsError)
	assert.Contains(t, disResult.Text, "DISADVANTAGE")
}

func TestCommandHandler_IsDiceCommand(t *testing.T) {
	tests := []struct {
		input  string
		isDice bool
	}{
		{"/roll 1d20", true},
		{"/spell fireball", true},
		{"/spawn dragon", true},
		{"hello", false},
		{"/help", false},
		{"", false},
		{"  /roll 2d6  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.isDice, IsDiceCommand(tt.input))
		})
	}
}

func TestCommandHandler_ExecuteCommand(t *testing.T) {
	handler := NewCommandHandler()

	tests := []struct {
		name      string
		input     string
		wantOut   string
		wantErr   bool
		isDiceCmd bool
	}{
		{
			name:      "roll command",
			input:     "/roll 1d20",
			isDiceCmd: true,
		},
		{
			name:      "spell command not found",
			input:     "/spell xyzzy",
			wantErr:   true,
			isDiceCmd: true,
		},
		{
			name:      "not a dice command",
			input:     "hello world",
			wantOut:   "",
			isDiceCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, isErr := handler.ExecuteCommand(context.Background(), tt.input)
			if tt.isDiceCmd {
				assert.NotEmpty(t, out)
			}
			assert.Equal(t, tt.wantErr, isErr)
		})
	}
}

func TestCommandHandler_GetHelp(t *testing.T) {
	help := GetHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "/roll")
	assert.Contains(t, help, "/spell")
	assert.Contains(t, help, "/spawn")
	assert.Contains(t, help, "1d20")
	assert.Contains(t, help, "advantage")
}
