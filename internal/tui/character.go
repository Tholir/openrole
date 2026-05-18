// Package tui provides the TUI foundation for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin built on Bubble Tea v2.
package tui

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/openrole/internal/types"
)

// CharacterSheet renders a D&D 5e character sheet in ASCII art style.
type CharacterSheet struct {
	Character *types.Character
	Width     int
}

// NewCharacterSheet creates a new character sheet renderer.
func NewCharacterSheet(char *types.Character, width int) CharacterSheet {
	return CharacterSheet{
		Character: char,
		Width:     width,
	}
}

// Render renders the full character sheet.
func (cs CharacterSheet) Render() string {
	var b strings.Builder

	b.WriteString(cs.renderHeader())
	b.WriteString(cs.renderAbilityScores())
	b.WriteString(cs.renderCombatStats())
	b.WriteString(cs.renderSkills())

	return b.String()
}

// renderHeader renders the character name and basic info box.
func (cs CharacterSheet) renderHeader() string {
	var b strings.Builder
	char := cs.Character

	// Box dimensions
	width := cs.Width
	innerWidth := width - 4

	// Top border
	b.WriteString(boxTopLeft)
	b.WriteString(strings.Repeat(boxHorizontal, innerWidth))
	b.WriteString(boxTopRight)
	b.WriteString("\n")

	// Title line
	title := fmt.Sprintf(" %s ", char.Name)
	padding := strings.Repeat(" ", innerWidth-len(title))
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical, title, padding, boxVertical, boxVertical))

	// Subtitle: Class, Level, Race
	subtitle := fmt.Sprintf(" Level %d %s %s", char.Level, char.Class, char.Race)
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical,
		strings.Repeat(" ", 1),
		subtitle,
		strings.Repeat(" ", innerWidth-2-len(subtitle)),
		boxVertical))

	// HP bar
	hpBar := cs.renderHPBar()
	b.WriteString(fmt.Sprintf("%s %s %s\n", boxVertical, hpBar, boxVertical))

	// Bottom border
	b.WriteString(boxBottomLeft)
	b.WriteString(strings.Repeat(boxHorizontal, innerWidth))
	b.WriteString(boxBottomRight)
	b.WriteString("\n")

	return b.String()
}

// renderAbilityScores renders the ability score block with modifiers.
func (cs CharacterSheet) renderAbilityScores() string {
	var b strings.Builder
	char := cs.Character
	mods := char.Stats.ModifierMap()

	// Section header
	b.WriteString(fmt.Sprintf("%s--- ABILITY SCORES ---%s\n", boxVertical, boxVertical))

	// Two-column layout for abilities
	abilities := []struct {
		name  string
		score int
	}{
		{"STR", char.Stats.Strength},
		{"DEX", char.Stats.Dexterity},
		{"CON", char.Stats.Constitution},
		{"INT", char.Stats.Intelligence},
		{"WIS", char.Stats.Wisdom},
		{"CHA", char.Stats.Charisma},
	}

	// Render in pairs: left column first half, right column second half
	for i := 0; i < 3; i++ {
		left := abilities[i]
		right := abilities[i+3]

		leftMod := mods[left.name]
		rightMod := mods[right.name]

		leftLine := fmt.Sprintf("%s %2d (%+d)", left.name, left.score, leftMod)
		rightLine := fmt.Sprintf("%s %2d (%+d)", right.name, right.score, rightMod)

		padding := cs.Width - 4 - len(leftLine) - len(rightLine)
		if padding < 0 {
			padding = 0
		}

		b.WriteString(fmt.Sprintf("%s %s%s%s %s %s\n",
			boxVertical,
			leftLine,
			strings.Repeat(" ", padding/2),
			rightLine,
			strings.Repeat(" ", padding-padding/2),
			boxVertical))
	}

	// Separator
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
		boxVertical,
		boxLeftT,
		strings.Repeat(boxHorizontal, cs.Width-4),
		boxRightT,
		boxVertical))

	return b.String()
}

// renderCombatStats renders HP and AC in a compact block.
func (cs CharacterSheet) renderCombatStats() string {
	var b strings.Builder
	char := cs.Character

	b.WriteString(fmt.Sprintf("%s--- COMBAT ---\n", boxVertical))
	b.WriteString(fmt.Sprintf("%s HP: %d/%d  AC: %d%s\n",
		boxVertical, char.HP, char.MaxHP, char.AC, boxVertical))

	return b.String()
}

// renderSkills renders the skills list.
func (cs CharacterSheet) renderSkills() string {
	var b strings.Builder
	char := cs.Character

	b.WriteString(fmt.Sprintf("%s--- SKILLS ---\n", boxVertical))

	if len(char.Skills) == 0 {
		b.WriteString(fmt.Sprintf("%s No skills yet%s\n", boxVertical, boxVertical))
	} else {
		for i, skill := range char.Skills {
			line := fmt.Sprintf(" • %s", skill)
			if i < len(char.Skills)-1 {
				line += ","
			}
			b.WriteString(fmt.Sprintf("%s%s%s\n", boxVertical, line, boxVertical))
		}
	}

	// Bottom of section
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
		boxVertical,
		boxBottomLeft,
		strings.Repeat(boxHorizontal, cs.Width-4),
		boxBottomRight,
		boxVertical))

	return b.String()
}

// renderHPBar renders an ASCII HP bar using filled blocks.
func (cs CharacterSheet) renderHPBar() string {
	char := cs.Character
	if char.MaxHP <= 0 {
		return "[░░░░░░░░░░] 0/0"
	}

	percent := float64(char.HP) / float64(char.MaxHP)
	totalBars := 10
	filledBars := int(percent * float64(totalBars))
	emptyBars := totalBars - filledBars

	bar := strings.Repeat("█", filledBars) + strings.Repeat("░", emptyBars)
	return fmt.Sprintf("[%s] %d/%d", bar, char.HP, char.MaxHP)
}

// RenderHPBarStatic renders an HP bar for a given HP and maxHP.
func RenderHPBarStatic(hp, maxHP int) string {
	if maxHP <= 0 {
		return "[░░░░░░░░░░] 0/0"
	}

	percent := float64(hp) / float64(maxHP)
	totalBars := 10
	filledBars := int(percent * float64(totalBars))
	emptyBars := totalBars - filledBars

	bar := strings.Repeat("█", filledBars) + strings.Repeat("░", emptyBars)
	return fmt.Sprintf("[%s] %d/%d", bar, hp, maxHP)
}