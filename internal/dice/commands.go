package dice

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// CommandHandler handles dice-related commands.
type CommandHandler struct {
	roller      *Roller
	spellDisplay *SpellDisplay
	monsterDisplay *MonsterDisplay
}

// NewCommandHandler creates a new dice command handler.
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		roller:        NewRoller(),
		spellDisplay:  NewSpellDisplay(),
		monsterDisplay: NewMonsterDisplay(),
	}
}

// RollCommandResult represents the result of a /roll command.
type RollCommandResult struct {
	Text     string
	IsError  bool
	IsHTML   bool
}

// SpellCommandResult represents the result of a /spell command.
type SpellCommandResult struct {
	Text    string
	IsError bool
}

// SpawnCommandResult represents the result of a /spawn command.
type SpawnCommandResult struct {
	Text    string
	IsError bool
}

var (
	rollCommandPattern   = regexp.MustCompile(`^\s*/roll\s+(.+)$`)
	spellCommandPattern  = regexp.MustCompile(`^\s*/spell\s+(.+)$`)
	spawnCommandPattern  = regexp.MustCompile(`^\s*/spawn\s+(.+)$`)
)

// HandleRoll processes /roll commands.
func (h *CommandHandler) HandleRoll(ctx context.Context, input string) RollCommandResult {
	matches := rollCommandPattern.FindStringSubmatch(input)
	if len(matches) < 2 {
		return RollCommandResult{
			Text:    "Usage: /roll <dice notation>\nExample: /roll 2d6+3, /roll 1d20, /roll 2d20kh1",
			IsError: true,
		}
	}

	expression := strings.TrimSpace(matches[1])
	
	// Check for advantage/disadvantage shortcuts
	lowerExpr := strings.ToLower(expression)
	if lowerExpr == "advantage" || lowerExpr == "adv" {
		result := h.roller.RollAdvantage()
		return RollCommandResult{
			Text: FormatAdvantage(result),
		}
	}
	if lowerExpr == "disadvantage" || lowerExpr == "dis" {
		result := h.roller.RollDisadvantage()
		return RollCommandResult{
			Text: FormatAdvantage(result),
		}
	}

	result, err := h.roller.ParseRoll(expression)
	if err != nil {
		return RollCommandResult{
			Text:    fmt.Sprintf("Error: %v", err),
			IsError: true,
		}
	}

	return RollCommandResult{
		Text: FormatResult(result),
	}
}

// HandleSpell processes /spell commands.
func (h *CommandHandler) HandleSpell(ctx context.Context, spellName string) SpellCommandResult {
	spellName = strings.TrimSpace(spellName)
	if spellName == "" {
		return SpellCommandResult{
			Text:    "Usage: /spell <spell name>\nExample: /spell fireball, /spell magic missile",
			IsError: true,
		}
	}

	info, err := h.spellDisplay.LookupSpell(ctx, spellName)
	if err != nil {
		return SpellCommandResult{
			Text:    fmt.Sprintf("Spell not found: %s\nError: %v", spellName, err),
			IsError: true,
		}
	}

	return SpellCommandResult{
		Text: FormatSpell(info),
	}
}

// HandleSpawn processes /spawn commands.
func (h *CommandHandler) HandleSpawn(ctx context.Context, monsterName string) SpawnCommandResult {
	monsterName = strings.TrimSpace(monsterName)
	
	// Handle random spawn
	if monsterName == "random" || monsterName == "" {
		stats, err := h.monsterDisplay.RandomMonster(ctx)
		if err != nil {
			return SpawnCommandResult{
				Text:    fmt.Sprintf("Error spawning random monster: %v", err),
				IsError: true,
			}
		}
		return SpawnCommandResult{
			Text: FormatMonster(stats),
		}
	}

	stats, err := h.monsterDisplay.LookupMonster(ctx, monsterName)
	if err != nil {
		return SpawnCommandResult{
			Text:    fmt.Sprintf("Monster not found: %s\nError: %v", monsterName, err),
			IsError: true,
		}
	}

	return SpawnCommandResult{
		Text: FormatMonster(stats),
	}
}

// ExecuteCommand parses and executes a dice command.
func (h *CommandHandler) ExecuteCommand(ctx context.Context, input string) (string, bool) {
	input = strings.TrimSpace(input)

	// Check if it's a roll command
	if rollCommandPattern.MatchString(input) {
		result := h.HandleRoll(ctx, input)
		return result.Text, result.IsError
	}

	// Check if it's a spell command
	if spellCommandPattern.MatchString(input) {
		matches := spellCommandPattern.FindStringSubmatch(input)
		if len(matches) >= 2 {
			result := h.HandleSpell(ctx, matches[1])
			return result.Text, result.IsError
		}
	}

	// Check if it's a spawn command
	if spawnCommandPattern.MatchString(input) {
		matches := spawnCommandPattern.FindStringSubmatch(input)
		if len(matches) >= 2 {
			result := h.HandleSpawn(ctx, matches[1])
			return result.Text, result.IsError
		}
	}

	return "", false
}

// IsDiceCommand checks if the input is a dice command.
func IsDiceCommand(input string) bool {
	input = strings.TrimSpace(input)
	return rollCommandPattern.MatchString(input) ||
		spellCommandPattern.MatchString(input) ||
		spawnCommandPattern.MatchString(input)
}

// GetHelp returns help text for dice commands.
func GetHelp() string {
	return `
D&D 5e Commands:
  /roll <dice>     Roll dice using notation (e.g., 2d6+3, 1d20, 2d20kh1)
                   Special: /roll advantage, /roll disadvantage
                   
  /spell <name>    Look up a D&D 5e spell by name
                   Example: /spell fireball
                   
  /spawn [name]    Spawn a monster by name, or "random" for random
                   Example: /spawn dragon, /spawn random

Dice Notation Examples:
  1d20         - Single d20
  2d6+3        - Two d6 plus 3
  1d8-1        - One d8 minus 1
  2d20kh1      - Advantage (2d20 keep highest)
  2d20kl1      - Disadvantage (2d20 keep lowest)
  4d6kh3       - Roll 4d6, keep highest 3 (for ability scores)
`
}