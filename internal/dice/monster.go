package dice

import (
	"context"
	"fmt"
	"strings"

	"github.com/gentleman-programming/openrole/internal/api"
)

// MonsterDisplay renders monster/creature stats in TUI format.
type MonsterDisplay struct {
	client *api.Open5eClient
}

// NewMonsterDisplay creates a new monster display helper.
func NewMonsterDisplay() *MonsterDisplay {
	return &MonsterDisplay{
		client: api.NewOpen5eClient(),
	}
}

// MonsterStats contains formatted monster information.
type MonsterStats struct {
	Name             string
	Size             string
	Type             string
	Alignment        string
	ArmorClass       string
	HitPoints        string
	Speed            string
	STR              int
	DEX              int
	CON              int
	INT              int
WIS              int
	CHA              int
	ChallengeRating  string
	XP               string
	Languages        string
	DamageResistances string
	DamageImmunities  string
	ConditionImmunities string
}

// FormatMonster formats monster stats in retro ASCII style.
func FormatMonster(stats *MonsterStats) string {
	var sb strings.Builder

	sb.WriteString("┌──────────────────────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ 👹 %-57s │\n", stats.Name))
	sb.WriteString("├──────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ %-15s │ %-41s │\n", stats.Size+" "+stats.Type, stats.Alignment))
	sb.WriteString("├──────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Armor Class: %-48s │\n", stats.ArmorClass))
	sb.WriteString(fmt.Sprintf("│ Hit Points: %-49s │\n", stats.HitPoints))
	sb.WriteString(fmt.Sprintf("│ Speed: %-52s │\n", stats.Speed))
	sb.WriteString("├──────────────────────────────────────────────────────────────┤\n")
	sb.WriteString("│                    ABILITY SCORES                            │\n")
	sb.WriteString("├──────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ STR: %-3d  DEX: %-3d  CON: %-3d  INT: %-3d  WIS: %-3d  CHA: %-3d │\n",
		stats.STR, stats.DEX, stats.CON, stats.INT, stats.WIS, stats.CHA))
	sb.WriteString("├──────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Challenge: %-13s │ XP: %-33s │\n", stats.ChallengeRating, stats.XP))
	
	if stats.DamageResistances != "" {
		sb.WriteString(fmt.Sprintf("│ Damage Resistances: %-43s │\n", stats.DamageResistances))
	}
	if stats.DamageImmunities != "" {
		sb.WriteString(fmt.Sprintf("│ Damage Immunities: %-44s │\n", stats.DamageImmunities))
	}
	if stats.ConditionImmunities != "" {
		sb.WriteString(fmt.Sprintf("│ Condition Immunities: %-40s │\n", stats.ConditionImmunities))
	}
	
	sb.WriteString(fmt.Sprintf("│ Languages: %-50s │\n", stats.Languages))
	sb.WriteString("└──────────────────────────────────────────────────────────────┘\n")
	return sb.String()
}

// LookupMonster searches for a monster by name and returns formatted stats.
func (m *MonsterDisplay) LookupMonster(ctx context.Context, name string) (*MonsterStats, error) {
	monsters, err := m.client.SearchMonsters(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search monsters: %w", err)
	}

	if len(monsters) == 0 {
		return nil, fmt.Errorf("monster not found: %s", name)
	}

	// Return first match
	monster := monsters[0]
	return m.monsterToStats(&monster), nil
}

// GetMonsterBySlug fetches a specific monster by its slug.
func (m *MonsterDisplay) GetMonsterBySlug(ctx context.Context, slug string) (*MonsterStats, error) {
	monster, err := m.client.GetMonster(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster: %w", err)
	}
	return m.monsterToStats(monster), nil
}

func (m *MonsterDisplay) monsterToStats(monster *api.Monster) *MonsterStats {
	// Format armor class
	acStr := fmt.Sprintf("%d", monster.ArmorClass[0])
	if len(monster.ArmorClass) > 1 {
		acStr = fmt.Sprintf("%d, %d", monster.ArmorClass[0], monster.ArmorClass[1])
	}

	// Format hit points
	hpStr := fmt.Sprintf("%d (%s)", monster.HitPoints, monster.HitDice)

	// Format speed
	var speedParts []string
	for k, v := range monster.Speed {
		if str, ok := v.(string); ok {
			speedParts = append(speedParts, fmt.Sprintf("%s: %s", k, str))
		}
	}
	speedStr := strings.Join(speedParts, ", ")

	// Format XP
	xpStr := fmt.Sprintf("%d XP", monster.XP)

	return &MonsterStats{
		Name:              monster.Name,
		Size:              monster.Size,
		Type:              monster.Type,
		Alignment:         monster.Alignment,
		ArmorClass:        acStr,
		HitPoints:         hpStr,
		Speed:             speedStr,
		STR:               monster.Strength,
		DEX:               monster.Dexterity,
		CON:               monster.Constitution,
		INT:               monster.Intelligence,
		WIS:               monster.Wisdom,
		CHA:               monster.Charisma,
		ChallengeRating:   monster.ChallengeRating,
		XP:                xpStr,
		Languages:         monster.Languages,
		DamageResistances: strings.Join(monster.DamageResistances, ", "),
		DamageImmunities:  strings.Join(monster.DamageImmunities, ", "),
		ConditionImmunities: strings.Join(monster.ConditionImmunities, ", "),
	}
}

// RandomMonster returns a random monster from the API.
func (m *MonsterDisplay) RandomMonster(ctx context.Context) (*MonsterStats, error) {
	response, err := m.client.ListMonsters(ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to list monsters: %w", err)
	}

	if len(response.Results) == 0 {
		return nil, fmt.Errorf("no monsters found")
	}

	// Return random monster from results
	monster := response.Results[0]
	return m.monsterToStats(&monster), nil
}

// ListMonstersByCR lists monsters by challenge rating.
func (m *MonsterDisplay) ListMonstersByCR(ctx context.Context, cr string) ([]*MonsterStats, error) {
	response, err := m.client.ListMonsters(ctx, 1)
	if err != nil {
		return nil, err
	}

	var result []*MonsterStats
	for _, monster := range response.Results {
		if monster.ChallengeRating == cr {
			stats := m.monsterToStats(&monster)
			result = append(result, stats)
		}
	}
	return result, nil
}