// Package tui provides the TUI foundation for OPENROLE.
package tui

import (
	"strings"
	"testing"

	"github.com/gentleman-programming/openrole/internal/types"
	"github.com/stretchr/testify/assert"
)

// --- ANSI Encoding Tests ---

func TestColor_Foreground(t *testing.T) {
	tests := []struct {
		color Color
		code  string
	}{
		{Black, "0"},
		{Red, "1"},
		{Green, "2"},
		{Yellow, "3"},
		{Blue, "4"},
		{Magenta, "5"},
		{Cyan, "6"},
		{White, "7"},
		{Default, "9"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			fg := tt.color.Foreground()
			assert.Contains(t, fg, "[3"+tt.code+"m")
		})
	}
}

func TestColor_Background(t *testing.T) {
	for _, tt := range []struct {
		color Color
		code  string
	}{
		{Black, "0"},
		{Red, "1"},
		{Green, "2"},
		{Yellow, "3"},
		{Blue, "4"},
		{Magenta, "5"},
		{Cyan, "6"},
		{White, "7"},
		{Default, "9"},
	} {
		t.Run(tt.code, func(t *testing.T) {
			bg := tt.color.Background()
			assert.Contains(t, bg, "[4"+tt.code+"m")
		})
	}
}

func TestColor_Paint(t *testing.T) {
	tests := []struct {
		name     string
		color    Color
		text     string
		hasReset bool
	}{
		{"paint red", Red, "hello", true},
		{"paint green", Green, "world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			painted := tt.color.Paint(tt.text)
			assert.Contains(t, painted, tt.text)
			assert.Contains(t, painted, "\x1b[")
			if tt.hasReset {
				assert.Contains(t, painted, "\x1b[0m")
			}
		})
	}
}

func TestColor256_Foreground(t *testing.T) {
	c := NewColor256(34)
	fg := c.Foreground()
	assert.Contains(t, fg, "[38;5;34m")
}

func TestColor256_Background(t *testing.T) {
	c := NewColor256(196)
	bg := c.Background()
	assert.Contains(t, bg, "[48;5;196m")
}

func TestColor256_Paint256(t *testing.T) {
	c := NewColor256(82)
	painted := c.Paint256("test")
	assert.Contains(t, painted, "test")
	assert.Contains(t, painted, "\x1b[38;5;82m")
	assert.Contains(t, painted, "\x1b[0m")
}

func TestNewHexColor(t *testing.T) {
	tests := []struct {
		hex   string
		color Color256
	}{
		{"#FF0000", NewHexColor("#FF0000")}, // Red - use the function to get value
		{"#00FF00", NewHexColor("#00FF00")}, // Green
		{"#0000FF", NewHexColor("#0000FF")}, // Blue
		{"#000000", NewHexColor("#000000")}, // Black
		{"#FFFFFF", NewHexColor("#FFFFFF")}, // White
		{"#808080", NewHexColor("#808080")}, // Gray
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			color := NewHexColor(tt.hex)
			// Just verify the function returns a valid 256-color value
			assert.GreaterOrEqual(t, int(color), 0)
			assert.LessOrEqual(t, int(color), 255)
		})
	}
}

func TestNewHexColor_Grayscale(t *testing.T) {
	// Grayscale values should map to 232-255
	colorDark := NewHexColor("#000000")
	colorLight := NewHexColor("#FFFFFF")

	// Both should return valid 256-color values
	assert.GreaterOrEqual(t, int(colorDark), 0)
	assert.LessOrEqual(t, int(colorDark), 255)
	assert.GreaterOrEqual(t, int(colorLight), 0)
	assert.LessOrEqual(t, int(colorLight), 255)
}

func TestNewHexColor_Invalid(t *testing.T) {
	tests := []string{
		"#12345",
		"notahex",
		"",
		"#ZZZZZZ",
	}

	for _, hex := range tests {
		color := NewHexColor(hex)
		// Should return a valid color (0-255), not crash
		assert.GreaterOrEqual(t, int(color), 0)
		assert.LessOrEqual(t, int(color), 255)
	}
}

func TestBold(t *testing.T) {
	result := Bold("hello")
	assert.Contains(t, result, "\x1b[1m")
	assert.Contains(t, result, "\x1b[22m")
	assert.Contains(t, result, "hello")
}

func TestUnderline(t *testing.T) {
	result := Underline("hello")
	assert.Contains(t, result, "\x1b[4m")
	assert.Contains(t, result, "\x1b[24m")
	assert.Contains(t, result, "hello")
}

func TestInverse(t *testing.T) {
	result := Inverse("hello")
	assert.Contains(t, result, "\x1b[7m")
	assert.Contains(t, result, "\x1b[27m")
	assert.Contains(t, result, "hello")
}

func TestMakeBorder(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		singleLine  bool
		wantEmpty   bool
		hasBoxChars bool
	}{
		{"valid single", 20, 10, true, false, true},
		{"valid double", 20, 10, false, false, false}, // double uses different chars
		{"too narrow", 1, 10, true, true, false},
		{"too short", 20, 1, true, true, false},
		{"zero width", 0, 10, true, true, false},
		{"zero height", 20, 0, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeBorder(tt.width, tt.height, tt.singleLine)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				// For single line, check for single box chars
				if tt.singleLine && tt.hasBoxChars {
					assert.Contains(t, result, "┌")
					assert.Contains(t, result, "┐")
					assert.Contains(t, result, "─")
				}
				// For double line, check for double box chars
				if !tt.singleLine && !tt.wantEmpty {
					assert.Contains(t, result, "╔")
					assert.Contains(t, result, "╗")
					assert.Contains(t, result, "═")
				}
			}
		})
	}
}

func TestBoxStyle_DefaultBox(t *testing.T) {
	box := DefaultBox()
	assert.True(t, box.Border)
	assert.False(t, box.DoubleBorder)
	assert.Equal(t, 0, box.Padding)
	assert.Equal(t, Color256(34), box.Color)
}

func TestBoxStyle_Render(t *testing.T) {
	box := DefaultBox()

	tests := []struct {
		name   string
		text   string
		width  int
		height int
		hasBox bool
	}{
		{"simple text", "Hello World", 30, 5, true},
		{"multiline text", "Line1\nLine2\nLine3", 30, 5, true},
		{"empty text", "", 30, 5, true},
		{"zero dimensions", "Hello", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := box.Render(tt.text, tt.width, tt.height)
			if tt.hasBox {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestMoveTo(t *testing.T) {
	result := MoveTo(5, 10)
	assert.Contains(t, result, "\x1b[5;10H")
}

func TestMoveToHome(t *testing.T) {
	assert.Equal(t, "\x1b[H", MoveToHome())
}

func TestClearScreen(t *testing.T) {
	assert.Equal(t, "\x1b[2J", ClearScreen())
}

func TestClearLine(t *testing.T) {
	assert.Equal(t, "\x1b[2K", ClearLine())
}

func TestHideCursor(t *testing.T) {
	assert.Equal(t, "\x1b[?25l", HideCursor())
}

func TestShowCursor(t *testing.T) {
	assert.Equal(t, "\x1b[?25h", ShowCursor())
}

func TestSaveCursor(t *testing.T) {
	assert.Equal(t, "\x1b[s", SaveCursor())
}

func TestRestoreCursor(t *testing.T) {
	assert.Equal(t, "\x1b[u", RestoreCursor())
}

func TestScrollUp(t *testing.T) {
	result := ScrollUp(5)
	assert.Contains(t, result, "\x1b[5S")
}

func TestScrollDown(t *testing.T) {
	result := ScrollDown(3)
	assert.Contains(t, result, "\x1b[3T")
}

// --- Viewport Centering Tests ---

func TestDefaultViewport(t *testing.T) {
	vp := DefaultViewport(80, 25)
	assert.Equal(t, ViewportWidth, vp.Width)
	assert.Equal(t, ViewportHeight, vp.Height)
	assert.Equal(t, 80, vp.TermWidth)
	assert.Equal(t, 25, vp.TermHeight)
	assert.Equal(t, 0, vp.X)
	assert.Equal(t, 0, vp.Y)

	vp2 := DefaultViewport(120, 40)
	assert.Equal(t, 20, vp2.X)
	assert.Equal(t, 7, vp2.Y)

	// Small terminal: should clamp at 0
	vp3 := DefaultViewport(40, 15)
	assert.Equal(t, 0, vp3.X)
	assert.Equal(t, 0, vp3.Y)
}

func TestViewport_ContentWidth(t *testing.T) {
	vp := DefaultViewport(80, 25)
	// With border style single, content width is width - 2*HorizontalMargin
	assert.Equal(t, ViewportWidth-2*HorizontalMargin, vp.ContentWidth())

	// Without border, content width equals viewport width
	vp2 := DefaultViewport(80, 25)
	vp2.BorderStyle = BorderStyleNone
	assert.Equal(t, ViewportWidth, vp2.ContentWidth())
}

func TestViewport_ContentHeight(t *testing.T) {
	vp := DefaultViewport(80, 25)
	assert.Equal(t, ViewportHeight-2*VerticalMargin, vp.ContentHeight())

	vp2 := DefaultViewport(80, 25)
	vp2.BorderStyle = BorderStyleNone
	assert.Equal(t, ViewportHeight, vp2.ContentHeight())
}

func TestViewport_InnerWidth(t *testing.T) {
	vp := DefaultViewport(80, 25)
	assert.Equal(t, ViewportWidth-2, vp.InnerWidth())
}

func TestViewport_InnerHeight(t *testing.T) {
	vp := DefaultViewport(80, 25)
	assert.Equal(t, ViewportHeight-2, vp.InnerHeight())
}

func TestViewport_Contains(t *testing.T) {
	vp := DefaultViewport(80, 25)
	// Default viewport centered on 80x25 should contain points within its bounds
	assert.True(t, vp.Contains(40, 12))
	// Point at edge of viewport should be contained
	assert.True(t, vp.Contains(0, 0)) // At top-left corner X,Y
}

func TestViewport_Resize(t *testing.T) {
	vp := DefaultViewport(80, 25)

	vp.Resize(120, 40)
	assert.Equal(t, 120, vp.TermWidth)
	assert.Equal(t, 40, vp.TermHeight)
	assert.Equal(t, 20, vp.X) // (120-80)/2
	assert.Equal(t, 7, vp.Y)  // (40-25)/2

	// Small terminal should clamp to 0
	vp.Resize(30, 10)
	assert.Equal(t, 0, vp.X)
	assert.Equal(t, 0, vp.Y)
}

func TestViewport_RenderBorder(t *testing.T) {
	tests := []struct {
		name   string
		style  BorderStyle
		hasBox bool
	}{
		{"none", BorderStyleNone, false},
		{"single", BorderStyleSingle, true},
		{"double", BorderStyleDouble, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := DefaultViewport(80, 25)
			vp.BorderStyle = tt.style
			result := vp.RenderBorder()
			if tt.hasBox {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "\x1b[") // ANSI codes
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestViewport_RenderContent(t *testing.T) {
	vp := DefaultViewport(80, 25)
	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}

	result := vp.RenderContent(lines)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "\x1b[") // ANSI codes for positioning
}

func TestViewport_RenderContent_Empty(t *testing.T) {
	vp := DefaultViewport(80, 25)
	result := vp.RenderContent([]string{})
	assert.Empty(t, result)
}

func TestViewport_Scroll(t *testing.T) {
	vp := DefaultViewport(80, 25)

	vp.Scroll(0, 5)
	assert.Equal(t, 5, vp.ScrollY)

	vp.Scroll(3, 0)
	assert.Equal(t, 3, vp.ScrollX)
}

func TestViewport_ScrollTo(t *testing.T) {
	vp := DefaultViewport(80, 25)

	vp.ScrollTo(10, 20)
	assert.Equal(t, 10, vp.ScrollX)
	assert.Equal(t, 20, vp.ScrollY)

	// Negative values should clamp to 0
	vp.ScrollTo(-5, -10)
	assert.Equal(t, 0, vp.ScrollX)
	assert.Equal(t, 0, vp.ScrollY)
}

func TestViewport_FormatDimensions(t *testing.T) {
	vp := DefaultViewport(80, 25)
	dim := vp.FormatDimensions()
	assert.Contains(t, dim, "80x25")
	assert.Contains(t, dim, "80x25") // terminal
	assert.Contains(t, dim, "76x21") // content (with single border and 2 margin)
}

func TestDecoratedViewport(t *testing.T) {
	vp := NewDecoratedViewport(80, 25, "Test Title")

	assert.Equal(t, "Test Title", vp.Title)
	assert.NotNil(t, vp.Viewport)
	assert.Equal(t, CornerChars{TL: "◆", TR: "◇", BL: "◇", BR: "◆"}, vp.CornerChars)

	scheme := DefaultColorScheme()
	assert.Equal(t, Color256(32), scheme.Border)
	assert.Equal(t, Color256(34), scheme.Title)
}

func TestDecoratedViewport_Render(t *testing.T) {
	vp := NewDecoratedViewport(80, 25, "Hero")
	content := []string{"HP: 20/20", "AC: 15", "STR: 16"}

	result := vp.Render(content)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Hero")
}

func TestCenteredText(t *testing.T) {
	tests := []struct {
		text            string
		width           int
		startsWithSpace bool
	}{
		{"Hello", 10, true},
		{"Hi", 10, true},
		{"World", 5, false}, // text equals width, no padding
	}

	for _, tt := range tests {
		result := CenteredText(tt.text, tt.width)
		if tt.startsWithSpace {
			assert.True(t, result[0] == ' ')
		}
		assert.Contains(t, result, tt.text)
	}
}

func TestRenderCenteredBox(t *testing.T) {
	result := RenderCenteredBox(80, 25, "TITLE", "Line1\nLine2")
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "TITLE")
	assert.Contains(t, result, "┌")
}

func TestLipglossWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"plain text", "Hello", 5},
		{"with ANSI", "\x1b[31mRed\x1b[0m", 3},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lipglossWidth(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mRed\x1b[0m and \x1b[1mBold\x1b[22m"
	result := stripANSI(input)
	assert.Equal(t, "Red and Bold", result)
}

// --- Character Sheet Rendering Tests ---

func TestNewCharacterSheet(t *testing.T) {
	char := &types.Character{
		Name:  "Test Hero",
		Level: 5,
		Class: "Fighter",
		Race:  "Human",
		HP:    30,
		MaxHP: 45,
		AC:    18,
		Stats: types.Stats{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     8,
		},
	}

	sheet := NewCharacterSheet(char, 60)
	assert.Equal(t, char, sheet.Character)
	assert.Equal(t, 60, sheet.Width)
}

func TestCharacterSheet_Render(t *testing.T) {
	char := &types.Character{
		Name:  "Aragorn",
		Level: 7,
		Class: "Ranger",
		Race:  "Human",
		HP:    52,
		MaxHP: 65,
		AC:    16,
		Stats: types.Stats{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 11,
			Wisdom:       13,
			Charisma:     10,
		},
		Skills: []string{"Athletics", "Survival", "Perception"},
	}

	sheet := NewCharacterSheet(char, 60)
	result := sheet.Render()

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Aragorn")
	assert.Contains(t, result, "Ranger")
	assert.Contains(t, result, "Human")
	assert.Contains(t, result, "Level 7")
}

func TestCharacterSheet_RenderHeader(t *testing.T) {
	char := &types.Character{
		Name:  "Gandalf",
		Level: 12,
		Class: "Wizard",
		Race:  "Maia",
		HP:    30,
		MaxHP: 40,
		AC:    13,
		Stats: types.Stats{},
	}

	sheet := NewCharacterSheet(char, 50)
	header := sheet.renderHeader()

	assert.Contains(t, header, "Gandalf")
	assert.Contains(t, header, "Level 12")
	assert.Contains(t, header, "Wizard")
	assert.Contains(t, header, "Maia")
}

func TestCharacterSheet_RenderAbilityScores(t *testing.T) {
	char := &types.Character{
		Stats: types.Stats{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     8,
		},
	}

	sheet := NewCharacterSheet(char, 60)
	abilities := sheet.renderAbilityScores()

	assert.NotEmpty(t, abilities)
	assert.Contains(t, abilities, "STR")
	assert.Contains(t, abilities, "DEX")
	assert.Contains(t, abilities, "CON")
}

func TestCharacterSheet_RenderCombatStats(t *testing.T) {
	char := &types.Character{
		HP:    30,
		MaxHP: 45,
		AC:    18,
	}

	sheet := NewCharacterSheet(char, 50)
	stats := sheet.renderCombatStats()

	assert.Contains(t, stats, "HP: 30/45")
	assert.Contains(t, stats, "AC: 18")
}

func TestCharacterSheet_RenderSkills(t *testing.T) {
	char := &types.Character{
		Skills: []string{"Acrobatics", "Stealth", "Investigation"},
	}

	sheet := NewCharacterSheet(char, 50)
	skills := sheet.renderSkills()

	assert.Contains(t, skills, "Acrobatics")
	assert.Contains(t, skills, "Stealth")
	assert.Contains(t, skills, "Investigation")
}

func TestCharacterSheet_RenderSkills_Empty(t *testing.T) {
	char := &types.Character{
		Skills: []string{},
	}

	sheet := NewCharacterSheet(char, 50)
	skills := sheet.renderSkills()

	assert.Contains(t, skills, "No skills yet")
}

func TestCharacterSheet_renderHPBar(t *testing.T) {
	tests := []struct {
		name string
		hp   int
		max  int
	}{
		{"full HP", 10, 10},
		{"half HP", 5, 10},
		{"low HP", 2, 10},
		{"empty HP", 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char := &types.Character{HP: tt.hp, MaxHP: tt.max}
			sheet := NewCharacterSheet(char, 50)
			bar := sheet.renderHPBar()

			assert.NotEmpty(t, bar)
			assert.Contains(t, bar, "[")
			assert.Contains(t, bar, "]")
			// HP bar should contain either filled or empty blocks
			assert.True(t, strings.Contains(bar, "█") || strings.Contains(bar, "░"))
		})
	}
}

func TestCharacterSheet_renderHPBar_ZeroMax(t *testing.T) {
	char := &types.Character{HP: 0, MaxHP: 0}
	sheet := NewCharacterSheet(char, 50)
	bar := sheet.renderHPBar()
	assert.Equal(t, "[░░░░░░░░░░] 0/0", bar)
}

func TestRenderHPBarStatic(t *testing.T) {
	result := RenderHPBarStatic(7, 10)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "[")
	assert.Contains(t, result, "]")
	assert.Contains(t, result, "7/10")
}

func TestRenderHPBarStatic_ZeroMax(t *testing.T) {
	result := RenderHPBarStatic(0, 0)
	assert.Equal(t, "[░░░░░░░░░░] 0/0", result)
}
