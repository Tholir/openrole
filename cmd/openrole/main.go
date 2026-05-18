// Package main is the entry point for OPENROLE, a retro terminal D&D 5e
// roleplaying plugin powered by Bubble Tea.
package main

import (
	"context"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/gentleman-programming/openrole/internal/types"

	tea "charm.land/bubbletea/v2"
)

// App colors - retro terminal aesthetic
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff4444"))
)

// model represents the Bubble Tea model for OPENROLE.
type model struct {
	sessionID string
	dmState   *types.DMState
	players   []*types.Player
	input     string
	output    []string
	err       error
}

// Init initializes the Bubble Tea model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			// Process command
			if m.input != "" {
				m.output = append(m.output, fmt.Sprintf("> %s", m.input))
				m.input = ""
			}
		default:
			// Accumulate input
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	case error:
		m.err = msg
	}
	return m, nil
}

// View renders the TUI view.
func (m model) View() tea.View {
	s := headerStyle.Render("OPENROLE")
	s += "\n"
	s += infoStyle.Render("Retro Terminal D&D 5e | Press Ctrl+C or Q to quit")
	s += "\n\n"

	if len(m.output) > 0 {
		start := 0
		if len(m.output) > 20 {
			start = len(m.output) - 20
		}
		for _, line := range m.output[start:] {
			s += line + "\n"
		}
		s += "\n"
	}

	s += infoStyle.Render(fmt.Sprintf("DM: %s | Session: %s",
		truncate(m.dmState.Personality.BaseDescription, 40), m.sessionID[:min(8, len(m.sessionID))]))

	if m.err != nil {
		s += "\n"
		s += errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	s += "\n"
	s += "> " + m.input
	s += " "

	return tea.NewView(s)
}

// truncate truncates a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	ctx := context.Background()

	// Initialize DM state
	sessionID := "openrole-session-001"
	dmState := types.NewDMState(sessionID)

	// Create initial player for testing
	player := types.NewPlayer("TestPlayer")
	player.CharacterID = "test-character-001"

	initialModel := model{
		sessionID: sessionID,
		dmState:   dmState,
		players:   []*types.Player{player},
		output: []string{
			"Welcome to OPENROLE!",
			fmt.Sprintf("Your DM whispers: %s", dmState.Personality.IcebreakerFacts[0]),
			"",
		},
	}

	p := tea.NewProgram(initialModel, tea.WithContext(ctx))

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start OPENROLE: %v\n", err)
		os.Exit(1)
	}
}
