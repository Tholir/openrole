// Package tui provides the TUI foundation for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin built on Bubble Tea v2.
//
// This package implements:
//   - Bubble Tea model with message handling (init/update/view pattern)
//   - ANSI escape sequence utilities for colors and styling
//   - Viewport scaling with 80x25 default and decorative borders
package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
	"charm.land/lipgloss/v2"
	"github.com/gentleman-programming/openrole/internal/types"

	tea "charm.land/bubbletea/v2"
)

// Default viewport dimensions.
const (
	DefaultWidth  = 80
	DefaultHeight = 25
)

// Model represents the main Bubble Tea model for OPENROLE.
// It follows the init/update/view pattern required by Bubble Tea v2.
type Model struct {
	// Viewport dimensions
	Width  int
	Height int

	// Session state
	SessionID string
	DMState   *types.DMState
	Players   []*types.Player
	Characters []*types.Character

	// Input state
	Input      string
	InputPos   int
	History    []string
	HistoryIdx int

	// Output buffer
	Output    []string
	OutputTop int // Top line of visible output (for scrolling)

	// View state
	CurrentView  ViewState
	BorderStyle  BorderStyle
	Animating    bool

	// Animation system
	Animator *Animator

	// Error state
	Err error
}

// ViewState represents the current view mode.
type ViewState int

const (
	ViewMain ViewState = iota
	ViewCharacter
	ViewCombat
	ViewInventory
	ViewDiceRoll
	ViewMenu
)

// BorderStyle defines the visual style of borders.
type BorderStyle int

const (
	BorderStyleNone BorderStyle = iota
	BorderStyleSingle
	BorderStyleDouble
)

// Init initializes the Bubble Tea model.
// This is called once at startup.
func (m Model) Init() tea.Cmd {
	m.Width = DefaultWidth
	m.Height = DefaultHeight
	m.Output = []string{}
	m.History = []string{}
	m.HistoryIdx = -1
	m.CurrentView = ViewMain
	m.BorderStyle = BorderStyleSingle

	return nil
}

// Update handles incoming messages and updates the model.
// This is the main update loop following the Elm architecture.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case OutputMsg:
		m.Output = append(m.Output, msg.Text)
		if len(m.Output) > 100 {
			m.OutputTop = len(m.Output) - 100
		}
		return m, nil

	case ClearOutputMsg:
		m.Output = []string{}
		m.OutputTop = 0
		return m, nil

	case SetViewMsg:
		m.CurrentView = msg.View
		return m, nil

	case RollDiceMsg:
		output := fmt.Sprintf("Rolled %s -> %d (%s)", msg.Dice, msg.Result, msg.Details)
		m.Output = append(m.Output, output)
		return m, nil

	case TickMsg:
		m.Animating = true
		// Advance animation systems
		if m.Animator != nil {
			m.Animator.tick()
		}
		return m, nil

	case AnimatorStartMsg:
		if m.Animator != nil && !m.Animator.Running {
			return m, m.Animator.Start()
		}
		return m, nil

	case AnimatorStopMsg:
		if m.Animator != nil && m.Animator.Running {
			m.Animator.Stop()
		}
		return m, nil

	default:
		// Check for error type
		if err, ok := msg.(error); ok {
			m.Err = err
			return m, nil
		}
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.Key()
	keyStr := key.String()

	switch {
	case keyStr == "ctrl+c", keyStr == "q":
		return m, tea.Quit

	case keyStr == "enter":
		if m.Input != "" {
			m.History = append(m.History, m.Input)
			m.HistoryIdx = len(m.History)
			m.Output = append(m.Output, fmt.Sprintf("> %s", m.Input))
			m.Input = ""
			m.InputPos = 0
		}
		return m, nil

	case keyStr == "up":
		if len(m.History) > 0 && m.HistoryIdx > 0 {
			m.HistoryIdx--
			m.Input = m.History[m.HistoryIdx]
			m.InputPos = len(m.Input)
		}
		return m, nil

	case keyStr == "down":
		if m.HistoryIdx < len(m.History)-1 {
			m.HistoryIdx++
			m.Input = m.History[m.HistoryIdx]
		} else {
			m.HistoryIdx = len(m.History)
			m.Input = ""
		}
		m.InputPos = len(m.Input)
		return m, nil

	case keyStr == "backspace":
		if m.InputPos > 0 {
			m.Input = m.Input[:m.InputPos-1] + m.Input[m.InputPos:]
			m.InputPos--
		}
		return m, nil

	case keyStr == "delete":
		if m.InputPos < len(m.Input) {
			m.Input = m.Input[:m.InputPos] + m.Input[m.InputPos+1:]
		}
		return m, nil

	case keyStr == "left":
		if m.InputPos > 0 {
			m.InputPos--
		}
		return m, nil

	case keyStr == "right":
		if m.InputPos < len(m.Input) {
			m.InputPos++
		}
		return m, nil

	case keyStr == "home":
		m.InputPos = 0
		return m, nil

	case keyStr == "end":
		m.InputPos = len(m.Input)
		return m, nil

	case key.Text != "":
		// Handle regular character input via key.Text
		for _, r := range key.Text {
			if unicode.IsPrint(r) {
				m.Input = m.Input[:m.InputPos] + string(r) + m.Input[m.InputPos:]
				m.InputPos++
			}
		}
		return m, nil

	default:
		return m, nil
	}
}

// handleMouse processes mouse input.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Mouse support can be extended for clickable UI elements
	_ = msg
	return m, nil
}

// View returns the rendered TUI view.
// This is called after each Update.
func (m Model) View() tea.View {
	var s strings.Builder

	// Calculate centering for content area
	viewport := m.calculateViewport()

	// Render header
	s.WriteString(m.renderHeader())

	// Render main content based on view state
	switch m.CurrentView {
	case ViewMain:
		s.WriteString(m.renderMainView(viewport))
	case ViewCharacter:
		s.WriteString(m.renderCharacterView(viewport))
	case ViewCombat:
		s.WriteString(m.renderCombatView(viewport))
	case ViewInventory:
		s.WriteString(m.renderInventoryView(viewport))
	case ViewDiceRoll:
		s.WriteString(m.renderDiceView(viewport))
	case ViewMenu:
		s.WriteString(m.renderMenuView(viewport))
	}

	// Render status bar
	s.WriteString(m.renderStatusBar())

	// Render input line
	s.WriteString(m.renderInputLine())

	return tea.NewView(s.String())
}

// viewport represents the centered content area.
type viewport struct {
	Left   int
	Top    int
	Width  int
	Height int
}

// calculateViewport computes the centered viewport rectangle.
func (m Model) calculateViewport() viewport {
	// Default 80x25 retro terminal dimensions
	contentWidth := DefaultWidth - 4 // 2 char padding each side
	contentHeight := DefaultHeight - 5 // header + status + input = 5 lines

	// If terminal is larger, center the 80x25 content area
	width := contentWidth
	height := contentHeight
	left := 2
	top := 2

	if m.Width > DefaultWidth {
		left = (m.Width - DefaultWidth) / 2
		width = DefaultWidth - 4
	}

	if m.Height > DefaultHeight {
		top = (m.Height - DefaultHeight) / 2
		height = DefaultHeight - 5
	}

	return viewport{
		Left:   left,
		Top:    top,
		Width:  width,
		Height: height,
	}
}

// renderHeader renders the header with decorative border.
func (m Model) renderHeader() string {
	var s strings.Builder

	// Top border line
	headerBorder := strings.Repeat(boxHorizontal, m.Width-2)
	s.WriteString(boxTopLeft)
	s.WriteString(headerBorder)
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	// Header content with title
	title := " OPENROLE "
	subtitle := " Retro Terminal D&D 5e "

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff00")).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	// Title line
	s.WriteString(boxVertical)
	s.WriteString(" ")
	s.WriteString(titleStyle.Render(title))

	// Pad to center subtitle area
	padding := m.Width - 4 - lipgloss.Width(title) - lipgloss.Width(subtitle) - 4
	s.WriteString(strings.Repeat(" ", padding/2))
	s.WriteString(subtitleStyle.Render(subtitle))
	s.WriteString(" ")
	s.WriteString(boxVertical)
	s.WriteString("\n")

	// Separator line
	s.WriteString(boxVertical)
	s.WriteString(strings.Repeat("-", m.Width-2))
	s.WriteString(boxVertical)
	s.WriteString("\n")

	return s.String()
}

// renderMainView renders the main game view.
func (m Model) renderMainView(vp viewport) string {
	var s strings.Builder

	// Calculate available content area
	contentHeight := m.Height - 5 // header + separator + status + input

	// Border with decorative corners
	s.WriteString(boxTopLeft)
	s.WriteString(strings.Repeat(boxHorizontal, m.Width-2))
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	// Content area with output
	outputHeight := contentHeight - 3 // Leave room for separator and input hint
	for i := 0; i < outputHeight; i++ {
		s.WriteString(boxVertical)

		lineIdx := m.OutputTop + i
		if lineIdx < len(m.Output) {
			line := m.Output[lineIdx]
			// Truncate if needed
			lineWidth := lipgloss.Width(line)
			if lineWidth > m.Width-2 {
				line = truncateString(line, m.Width-2)
			}
			s.WriteString(" ")
			s.WriteString(line)
			s.WriteString(strings.Repeat(" ", m.Width-3-lineWidth))
		} else {
			s.WriteString(strings.Repeat(" ", m.Width-2))
		}

		s.WriteString(boxVertical)
		s.WriteString("\n")
	}

	return s.String()
}

// renderCharacterView renders the character sheet view.
func (m Model) renderCharacterView(vp viewport) string {
	var s strings.Builder

	// Render box top border
	s.WriteString(boxTopLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	// If we have a character, render the full character sheet
	if len(m.Characters) > 0 && m.Characters[0] != nil {
		char := m.Characters[0]
		sheet := NewCharacterSheet(char, vp.Width)
		s.WriteString(sheet.Render())
	} else {
		// No character selected - show placeholder
		s.WriteString(m.renderBoxContent("No character selected", vp.Width))
	}

	// Render box bottom border
	s.WriteString(boxBottomLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxBottomRight)
	s.WriteString("\n")

	return s.String()
}

// renderCombatView renders the combat view.
func (m Model) renderCombatView(vp viewport) string {
	var s strings.Builder

	// Render box top border
	s.WriteString(boxTopLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	if len(m.Characters) > 0 && m.Characters[0] != nil {
		char := m.Characters[0]

		// Combat header
		s.WriteString(fmt.Sprintf("%s--- COMBAT ---\n", boxVertical))

		// HP bar
		hpBar := RenderHPBarStatic(char.HP, char.MaxHP)
		s.WriteString(fmt.Sprintf("%s HP: %s %s\n", boxVertical, hpBar, boxVertical))

		// AC
		acLine := fmt.Sprintf(" AC: %d", char.AC)
		s.WriteString(fmt.Sprintf("%s %s %s\n", boxVertical, acLine, boxVertical))

		// Initiative (Dex modifier)
		initMod := char.Stats.Modifier(char.Stats.Dexterity)
		initLine := fmt.Sprintf(" Initiative: %+d", initMod)
		s.WriteString(fmt.Sprintf("%s %s %s\n", boxVertical, initLine, boxVertical))

		// Proficiency (derived from level)
		profBonus := 1 + (char.Level-1)/4
		profLine := fmt.Sprintf(" Proficiency: +%d", profBonus)
		s.WriteString(fmt.Sprintf("%s %s %s\n", boxVertical, profLine, boxVertical))

		// Speed (default 30 ft if not specified)
		speedLine := " Speed: 30 ft"
		s.WriteString(fmt.Sprintf("%s %s %s\n", boxVertical, speedLine, boxVertical))
	} else {
		s.WriteString(m.renderBoxContent("No character loaded", vp.Width))
	}

	// Render box bottom border
	s.WriteString(boxBottomLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxBottomRight)
	s.WriteString("\n")

	return s.String()
}

// renderInventoryView renders the inventory view.
func (m Model) renderInventoryView(vp viewport) string {
	var s strings.Builder

	// Render box top border
	s.WriteString(boxTopLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	// If we have a character, render the inventory display
	if len(m.Characters) > 0 && m.Characters[0] != nil {
		char := m.Characters[0]
		inv := NewInventoryDisplay(char, vp.Width)
		s.WriteString(inv.Render())
	} else {
		// No character selected - show placeholder
		s.WriteString(m.renderBoxContent("No character selected", vp.Width))
	}

	// Render box bottom border
	s.WriteString(boxBottomLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxBottomRight)
	s.WriteString("\n")

	return s.String()
}

// renderDiceView renders the dice roll view.
func (m Model) renderDiceView(vp viewport) string {
	var s strings.Builder

	// Render box top border
	s.WriteString(boxTopLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	// Header
	s.WriteString(fmt.Sprintf("%s--- DICE ROLLER ---\n", boxVertical))

	// Dice notation reference
	refLine := " Notation: XdY+Z (e.g. 2d20+5)"
	s.WriteString(fmt.Sprintf("%s%s%s\n", boxVertical, refLine, boxVertical))

	// Separator
	s.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical, boxLeftT,
		strings.Repeat(boxHorizontal, vp.Width-4), boxRightT, boxVertical))

	// Recent rolls section
	s.WriteString(fmt.Sprintf("%s--- RECENT ROLLS ---\n", boxVertical))

	if len(m.Output) == 0 {
		s.WriteString(fmt.Sprintf("%s No rolls yet%s\n", boxVertical, boxVertical))
	} else {
		// Show last 5 rolls from output that contain "Rolled"
		count := 0
		for i := len(m.Output) - 1; i >= 0 && count < 5; i-- {
			if strings.Contains(m.Output[i], "Rolled") {
				line := m.Output[i]
				if len(line) > vp.Width-4 {
					line = line[:vp.Width-4] + ".."
				}
				s.WriteString(fmt.Sprintf("%s %s %s\n", boxVertical, line, boxVertical))
				count++
			}
		}
		if count == 0 {
			s.WriteString(fmt.Sprintf("%s No rolls yet%s\n", boxVertical, boxVertical))
		}
	}

	// Help text
	s.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical, boxLeftT,
		strings.Repeat(boxHorizontal, vp.Width-4), boxRightT, boxVertical))
	s.WriteString(fmt.Sprintf("%s Type dice notation and press Enter %s\n", boxVertical, boxVertical))

	// Render box bottom border
	s.WriteString(boxBottomLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxBottomRight)
	s.WriteString("\n")

	return s.String()
}

// renderMenuView renders the main menu view.
func (m Model) renderMenuView(vp viewport) string {
	var s strings.Builder

	// Render box top border
	s.WriteString(boxTopLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	// Header
	s.WriteString(fmt.Sprintf("%s--- MAIN MENU ---\n", boxVertical))

	// Menu options
	menuItems := []string{
		"[C] Character Sheet",
		"[V] Combat Tracker",
		"[I] Inventory",
		"[D] Dice Roller",
		"[M] Main View",
		"[Q] Quit",
	}

	for _, item := range menuItems {
		padding := vp.Width - 4 - len(item)
		if padding < 0 {
			padding = 0
		}
		s.WriteString(fmt.Sprintf("%s %s%s%s %s\n", boxVertical, item, strings.Repeat(" ", padding), boxVertical, boxVertical))
	}

	// Separator
	s.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical, boxLeftT,
		strings.Repeat(boxHorizontal, vp.Width-4), boxRightT, boxVertical))

	// Session info
	sessionLine := fmt.Sprintf(" Session: %s", m.SessionID[:8])
	if len(sessionLine) > vp.Width-4 {
		sessionLine = sessionLine[:vp.Width-4] + ".."
	}
	padding := vp.Width - 4 - len(sessionLine)
	if padding < 0 {
		padding = 0
	}
	s.WriteString(fmt.Sprintf("%s %s%s%s\n", boxVertical, sessionLine, strings.Repeat(" ", padding), boxVertical))

	// Render box bottom border
	s.WriteString(boxBottomLeft)
	s.WriteString(strings.Repeat(boxHorizontal, vp.Width-2))
	s.WriteString(boxBottomRight)
	s.WriteString("\n")

	return s.String()
}

// Truncate truncates a string to fit within a given width.
func truncateString(s string, maxWidth int) string {
	return ansi.Truncate(s, maxWidth, "...")
}

// renderBoxTitle renders a centered title within a box.
func (m Model) renderBoxTitle(title string, width int) string {
	var s strings.Builder
	padding := (width - lipgloss.Width(title) - 4) / 2

	s.WriteString(boxTopLeft)
	s.WriteString(strings.Repeat(boxHorizontal, padding))
	s.WriteString(" ")
	s.WriteString(title)
	s.WriteString(" ")
	s.WriteString(strings.Repeat(boxHorizontal, width-padding-lipgloss.Width(title)-5))
	s.WriteString(boxTopRight)
	s.WriteString("\n")

	return s.String()
}

// renderBoxContent renders a single line of content inside a box border.
func (m Model) renderBoxContent(content string, width int) string {
	var s strings.Builder
	innerWidth := width - 4

	s.WriteString(boxVertical)
	s.WriteString(" ")
	s.WriteString(content)
	s.WriteString(strings.Repeat(" ", innerWidth-len(content)))
	s.WriteString(" ")
	s.WriteString(boxVertical)
	s.WriteString("\n")

	return s.String()
}

// renderStatusBar renders the status bar at the bottom of the viewport.
func (m Model) renderStatusBar() string {
	var s strings.Builder

	// Separator
	s.WriteString(boxVertical)
	s.WriteString(strings.Repeat("-", m.Width-2))
	s.WriteString(boxVertical)
	s.WriteString("\n")

	// Status content
	s.WriteString(boxVertical)
	s.WriteString(" ")

	// Session info
	sessionInfo := fmt.Sprintf("Session: %s", m.SessionID[:8])
	s.WriteString(sessionInfo)

	// DM info
	dmInfo := ""
	if m.DMState != nil {
		dmInfo = fmt.Sprintf(" | DM: %s", m.DMState.Personality.BaseDescription[:30]+"...")
	}
	padding := m.Width - 2 - lipgloss.Width(sessionInfo) - lipgloss.Width(dmInfo) - 2
	s.WriteString(dmInfo)
	s.WriteString(strings.Repeat(" ", padding))

	s.WriteString(boxVertical)
	s.WriteString("\n")

	return s.String()
}

// renderInputLine renders the input prompt line.
func (m Model) renderInputLine() string {
	var s strings.Builder

	// Input line border
	s.WriteString(boxBottomLeft)
	s.WriteString(strings.Repeat(boxHorizontal, m.Width-2))
	s.WriteString(boxBottomRight)

	// Input prompt on new line (for better visibility)
	s.WriteString("\n")
	s.WriteString("> ")
	s.WriteString(m.Input)

	// Cursor positioning hint (simple blinking underscore effect via View)
	// Actual cursor position is tracked in InputPos

	return s.String()
}

// NewModel creates a new initialized Model with default values.
func NewModel(sessionID string, dmState *types.DMState) Model {
	return Model{
		Width:      DefaultWidth,
		Height:     DefaultHeight,
		SessionID:  sessionID,
		DMState:    dmState,
		Players:    []*types.Player{},
		Characters: []*types.Character{},
		Input:      "",
		InputPos:   0,
		History:    []string{},
		HistoryIdx: -1,
		Output:     []string{},
		OutputTop:  0,
		CurrentView: ViewMain,
		BorderStyle: BorderStyleSingle,
		Animating:  false,
		Animator:   NewAnimator(DefaultAnimationConfig()),
		Err:        nil,
	}
}

// AddPlayer adds a player to the session.
func (m *Model) AddPlayer(player *types.Player) {
	m.Players = append(m.Players, player)
}

// AddCharacter adds a character to the session.
func (m *Model) AddCharacter(char *types.Character) {
	m.Characters = append(m.Characters, char)
}

// SetView changes the current view state.
func (m *Model) SetView(view ViewState) {
	m.CurrentView = view
}

// AddOutput adds a line to the output buffer.
func (m *Model) AddOutput(text string) {
	m.Output = append(m.Output, text)
	if len(m.Output) > 100 {
		m.OutputTop = len(m.Output) - 100
	}
}

// --- Message types for the model ---

// OutputMsg adds a line to the output buffer.
type OutputMsg struct {
	Text string
}

// ClearOutputMsg clears the output buffer.
type ClearOutputMsg struct{}

// SetViewMsg changes the current view.
type SetViewMsg struct {
	View ViewState
}

// RollDiceMsg initiates a dice roll animation/result.
type RollDiceMsg struct {
	Dice    string // e.g., "2d20+5"
	Result  int
	Details string
}

// TickMsg is the animation tick message.
type TickMsg struct{}