package dice

import (
	"context"
	"fmt"
	"strings"

	"github.com/gentleman-programming/openrole/internal/api"
)

// SpellDisplay renders spell information in TUI-friendly format.
type SpellDisplay struct {
	client *api.Open5eClient
}

// NewSpellDisplay creates a new spell display helper.
func NewSpellDisplay() *SpellDisplay {
	return &SpellDisplay{
		client: api.NewOpen5eClient(),
	}
}

// SpellInfo contains formatted spell information.
type SpellInfo struct {
	Name           string
	Level          string
	School         string
	CastingTime    string
	Range          string
	Components     string
	Duration       string
	Description    string
	HigherLevel    string
	Concentration  bool
	Ritual         bool
	Classes        string
}

// FormatSpell formats a spell for display in retro ASCII style.
func FormatSpell(info *SpellInfo) string {
	var sb strings.Builder

	sb.WriteString("┌────────────────────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ ✨ %-56s │\n", info.Name))
	sb.WriteString("├────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Level: %-53s │\n", info.Level))
	sb.WriteString(fmt.Sprintf("│ School: %-52s │\n", info.School))
	sb.WriteString("├────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Casting Time: %-47s │\n", info.CastingTime))
	sb.WriteString(fmt.Sprintf("│ Range: %-52s │\n", info.Range))
	sb.WriteString(fmt.Sprintf("│ Components: %-49s │\n", info.Components))
	sb.WriteString(fmt.Sprintf("│ Duration: %-50s │\n", info.Duration))
	
	flags := []string{}
	if info.Concentration {
		flags = append(flags, "CONCENTRATION")
	}
	if info.Ritual {
		flags = append(flags, "RITUAL")
	}
	if len(flags) > 0 {
		sb.WriteString(fmt.Sprintf("│ ⚑ %-55s │\n", strings.Join(flags, ", ")))
	}
	
	sb.WriteString("├────────────────────────────────────────────────────────────┤\n")
	sb.WriteString("│ DESCRIPTION:                                                │\n")
	
	// Word wrap description
	descLines := wordWrap(info.Description, 60)
	for _, line := range descLines {
		sb.WriteString(fmt.Sprintf("│ %-60s │\n", line))
	}
	
	if info.HigherLevel != "" {
		sb.WriteString("├────────────────────────────────────────────────────────────┤\n")
		sb.WriteString("│ AT HIGHER LEVEL:                                            │\n")
		higherLines := wordWrap(info.HigherLevel, 60)
		for _, line := range higherLines {
			sb.WriteString(fmt.Sprintf("│ %-60s │\n", line))
		}
	}
	
	sb.WriteString("├────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Classes: %-51s │\n", info.Classes))
	sb.WriteString("└────────────────────────────────────────────────────────────┘\n")
	return sb.String()
}

// wordWrap wraps text to a specified width.
func wordWrap(text string, width int) []string {
	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}
		if currentLine.Len() > 0 {
			currentLine.WriteByte(' ')
		}
		currentLine.WriteString(word)
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	return lines
}

// LookupSpell searches for a spell by name and returns formatted info.
func (s *SpellDisplay) LookupSpell(ctx context.Context, name string) (*SpellInfo, error) {
	spells, err := s.client.SearchSpells(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search spells: %w", err)
	}

	if len(spells) == 0 {
		return nil, fmt.Errorf("spell not found: %s", name)
	}

	// Return first match
	spell := spells[0]
	return s.spellToInfo(&spell), nil
}

// GetSpellBySlug fetches a specific spell by its slug.
func (s *SpellDisplay) GetSpellBySlug(ctx context.Context, slug string) (*SpellInfo, error) {
	spell, err := s.client.GetSpell(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get spell: %w", err)
	}
	return s.spellToInfo(spell), nil
}

func (s *SpellDisplay) spellToInfo(spell *api.Spell) *SpellInfo {
	components := strings.Join(spell.Components, ", ")

	levelStr := spell.LevelStr
	if spell.Ritual {
		levelStr += " (Ritual)"
	}

	classes := strings.Join(spell.Classes, ", ")

	var higherLevel string
	if len(spell.HigherLevel) > 0 {
		higherLevel = strings.Join(spell.HigherLevel, " ")
	}

	return &SpellInfo{
		Name:          spell.Name,
		Level:         levelStr,
		School:        spell.School,
		CastingTime:   spell.CastingTime,
		Range:         spell.Range,
		Components:    components,
		Duration:      spell.Duration,
		Description:   spell.Desc,
		HigherLevel:   higherLevel,
		Concentration: spell.Concentration,
		Ritual:        spell.Ritual,
		Classes:       classes,
	}
}

// ListSpellsByLevel lists all spells of a specific level.
func (s *SpellDisplay) ListSpellsByLevel(ctx context.Context, level int) ([]*SpellInfo, error) {
	response, err := s.client.ListSpells(ctx, 1)
	if err != nil {
		return nil, err
	}

	var result []*SpellInfo
	for _, spell := range response.Results {
		if spell.Level == level {
			info := s.spellToInfo(&spell)
			result = append(result, info)
		}
	}
	return result, nil
}