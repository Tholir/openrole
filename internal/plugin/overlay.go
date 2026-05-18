// Package plugin provides the WASM plugin runtime for openrole using GopherJS.
// This file contains JavaScript UI overlay components for rendering in the browser.
package plugin

import (
	"context"
	"syscall/js"
)

// OverlayComponent represents a JavaScript UI component.
type OverlayComponent struct {
	id       string
	element  js.Value
	visible  bool
	onClose  func()
	renderer func(ctx context.Context, data interface{}) string
}

// OverlayManager manages all UI overlay components.
type OverlayManager struct {
	components map[string]*OverlayComponent
	container js.Value
	styles     CSSStyles
	mu         int // Protects components
}

// CSSStyles defines inline styles for overlay components.
type CSSStyles map[string]string

// DefaultOverlayStyles provides the default styling for overlays.
var DefaultOverlayStyles = CSSStyles{
	// Container styles
	"overlay-container": "position: fixed; top: 0; left: 0; width: 100%; height: 100%; z-index: 9999; display: flex; align-items: center; justify-content: center;",
	"overlay-backdrop": "position: absolute; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0, 0, 0, 0.7);",
	"overlay-content": "position: relative; background: #1a1a2e; border: 2px solid #4a4a6a; border-radius: 12px; padding: 24px; max-width: 600px; width: 90%; max-height: 80vh; overflow-y: auto; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);",

	// Character sheet specific
	"character-sheet": "font-family: 'Courier New', monospace; color: #e0e0e0;",
	"character-header": "border-bottom: 2px solid #6a6a8a; padding-bottom: 12px; margin-bottom: 16px;",
	"character-name": "font-size: 24px; font-weight: bold; color: #ffd700; text-shadow: 2px 2px 4px rgba(0,0,0,0.5);",
	"character-class": "font-size: 14px; color: #aaa; margin-top: 4px;",
	"stats-grid": "display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin: 16px 0;",
	"stat-box": "background: #2a2a4a; border: 1px solid #4a4a6a; border-radius: 8px; padding: 12px; text-align: center;",
	"stat-name": "font-size: 12px; color: #888; text-transform: uppercase;",
	"stat-value": "font-size: 28px; font-weight: bold; color: #fff;",
	"stat-modifier": "font-size: 14px; color: #7f7;",

	// HP and combat
	"hp-bar": "height: 24px; background: #333; border-radius: 12px; overflow: hidden; margin: 12px 0;",
	"hp-fill": "height: 100%; background: linear-gradient(90deg, #f44, #8b0000); transition: width 0.3s;",
	"hp-text": "text-align: center; color: #fff; font-weight: bold;",

	// Skills
	"skills-list": "margin-top: 16px;",
	"skill-item": "display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #333;",
	"skill-name": "color: #e0e0e0;",
	"skill-bonus": "color: #7f7;",

	// Vote prompt
	"vote-container": "text-align: center;",
	"vote-question": "font-size: 20px; color: #fff; margin-bottom: 24px;",
	"vote-options": "display: flex; flex-direction: column; gap: 12px;",
	"vote-option": "background: #2a2a4a; border: 2px solid #4a4a6a; border-radius: 8px; padding: 16px; cursor: pointer; transition: all 0.2s;",
	"vote-option:hover": "background: #3a3a6a; border-color: #7f7;",
	"vote-option.selected": "background: #3a5a3a; border-color: #7f7;",
	"vote-option-label": "font-size: 16px; color: #fff; font-weight: bold;",
	"vote-option-desc": "font-size: 12px; color: #aaa; margin-top: 4px;",

	// Dice result
	"dice-container": "text-align: center;",
	"dice-formula": "font-size: 18px; color: #888; margin-bottom: 16px;",
	"dice-total": "font-size: 64px; font-weight: bold; color: #ffd700; text-shadow: 3px 3px 6px rgba(0,0,0,0.5);",
	"dice-breakdown": "font-size: 14px; color: #aaa; margin-top: 12px;",
	"dice-individual": "display: inline-block; background: #2a2a4a; border: 1px solid #4a4a6a; border-radius: 4px; padding: 8px 12px; margin: 4px;",
	"dice-roll-value": "font-size: 20px; color: #fff;",

	// Close button
	"close-btn": "position: absolute; top: 12px; right: 12px; background: transparent; border: none; color: #888; font-size: 24px; cursor: pointer;",
	"close-btn:hover": "color: #fff;",
}

// NewOverlayManager creates a new overlay manager with the given container element.
func NewOverlayManager(container js.Value) *OverlayManager {
	return &OverlayManager{
		components: make(map[string]*OverlayComponent),
		container:  container,
		styles:      DefaultOverlayStyles,
	}
}

// SetStyles updates the overlay styles.
func (m *OverlayManager) SetStyles(styles CSSStyles) {
	m.styles = styles
	m.applyGlobalStyles()
}

// ShowCharacterSheet displays the character sheet overlay.
func (m *OverlayManager) ShowCharacterSheet(ctx context.Context, sheet CharacterSheet, onClose func()) {
	id := "character-sheet-" + sheet.ID

	component := &OverlayComponent{
		id:      id,
		visible: true,
		onClose: onClose,
		renderer: func(ctx context.Context, data interface{}) string {
			s := data.(CharacterSheet)
			return m.renderCharacterSheetHTML(s)
		},
	}

	m.components[id] = component
	m.renderComponent(ctx, component, sheet)
}

// ShowVotePrompt displays a voting prompt overlay.
func (m *OverlayManager) ShowVotePrompt(ctx context.Context, prompt VotePrompt, onVote func(selected []string), onClose func()) {
	id := "vote-prompt-" + prompt.ID

	component := &OverlayComponent{
		id:      id,
		visible: true,
		onClose: onClose,
		renderer: func(ctx context.Context, data interface{}) string {
			p := data.(VotePrompt)
			return m.renderVotePromptHTML(p)
		},
	}

	m.components[id] = component
	m.renderComponent(ctx, component, prompt)
}

// ShowDiceResult displays a dice roll result overlay.
func (m *OverlayManager) ShowDiceResult(ctx context.Context, result DiceResult, onClose func()) {
	id := "dice-result-" + result.RollID

	component := &OverlayComponent{
		id:      id,
		visible: true,
		onClose: onClose,
		renderer: func(ctx context.Context, data interface{}) string {
			r := data.(DiceResult)
			return m.renderDiceResultHTML(r)
		},
	}

	m.components[id] = component
	m.renderComponent(ctx, component, result)
}

// Hide removes an overlay by ID.
func (m *OverlayManager) Hide(id string) {
	if comp, ok := m.components[id]; ok {
		if comp.element.Truthy() {
			comp.element.Call("remove")
		}
		comp.visible = false
		delete(m.components, id)
	}
}

// HideAll removes all overlays.
func (m *OverlayManager) HideAll() {
	for id := range m.components {
		m.Hide(id)
	}
}

// renderComponent renders a component to the DOM.
func (m *OverlayManager) renderComponent(ctx context.Context, comp *OverlayComponent, data interface{}) {
	// Create the overlay container.
	backdrop := m.container.Get("document").Call("createElement", "div")
	backdrop.Set("className", "overlay-backdrop")
	backdrop.Get("style").Call("assign", parseStyles(m.styles["overlay-backdrop"]))

	content := m.container.Get("document").Call("createElement", "div")
	content.Set("className", "overlay-content")
	content.Get("style").Call("assign", parseStyles(m.styles["overlay-content"]))

	// Render the inner HTML.
	innerHTML := comp.renderer(ctx, data)
	content.Set("innerHTML", innerHTML)

	// Create close button.
	closeBtn := m.container.Get("document").Call("createElement", "button")
	closeBtn.Set("innerHTML", "&times;")
	closeBtn.Get("classList").Call("add", "close-btn")
	closeBtn.Get("style").Call("assign", parseStyles(m.styles["close-btn"]))
	closeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		m.Hide(comp.id)
		if comp.onClose != nil {
			comp.onClose()
		}
		return nil
	}))

	// Assemble.
	container := m.container.Get("document").Call("createElement", "div")
	container.Set("className", "overlay-container")
	container.Get("style").Call("assign", parseStyles(m.styles["overlay-container"]))
	container.Get("style").Set("position", "fixed")

	container.Call("appendChild", backdrop)
	container.Call("appendChild", content)
	content.Call("appendChild", closeBtn)

	m.container.Get("document").Call("body").Call("appendChild", container)
	comp.element = container
}

// renderCharacterSheetHTML generates the HTML for a character sheet.
func (m *OverlayManager) renderCharacterSheetHTML(sheet CharacterSheet) string {
	// Build stats HTML.
	statsHTML := ""
	for _, stat := range []struct {
		name  string
		value int
	}{
		{"STR", sheet.Stats.STR},
		{"DEX", sheet.Stats.DEX},
		{"CON", sheet.Stats.CON},
		{"INT", sheet.Stats.INT},
		{"WIS", sheet.Stats.WIS},
		{"CHA", sheet.Stats.CHA},
	} {
		mod := (stat.value - 10) / 2
		modStr := ""
		if mod >= 0 {
			modStr = "+" + itoa(mod)
		} else {
			modStr = itoa(mod)
		}
		statsHTML += `<div class="stat-box">
			<div class="stat-name">` + stat.name + `</div>
			<div class="stat-value">` + itoa(stat.value) + `</div>
			<div class="stat-modifier">` + modStr + `</div>
		</div>`
	}

	// Build skills HTML.
	skillsHTML := ""
	for _, skill := range sheet.Skills {
		prof := ""
		if skill.Proficient {
			prof = " (proficient)"
		}
		if skill.Expertise {
			prof = " (expertise)"
		}
		skillsHTML += `<div class="skill-item">
			<span class="skill-name">` + skill.Name + prof + `</span>
			<span class="skill-bonus">+` + itoa(skill.Bonus) + `</span>
		</div>`
	}

	// HP percentage.
	hpPercent := 100
	if sheet.HP.Max > 0 {
		hpPercent = (sheet.HP.Current * 100) / sheet.HP.Max
		if hpPercent < 0 {
			hpPercent = 0
		}
	}

	return `
		<div class="character-sheet">
			<div class="character-header">
				<div class="character-name">` + sheet.Name + `</div>
				<div class="character-class">Level ` + itoa(sheet.Level) + ` ` + sheet.Race + ` ` + sheet.Class + `</div>
			</div>

			<div class="hp-bar">
				<div class="hp-fill" style="width: ` + itoa(hpPercent) + `%"></div>
			</div>
			<div class="hp-text">HP: ` + itoa(sheet.HP.Current) + `/` + itoa(sheet.HP.Max) + `</div>

			<div class="stats-grid">
				` + statsHTML + `
			</div>

			<div style="margin-top: 16px;">
				<strong>AC:</strong> ` + itoa(sheet.AC) + ` |
				<strong>Speed:</strong> ` + itoa(sheet.Speed) + `ft |
				<strong>Alignment:</strong> ` + sheet.Alignment + `
			</div>

			<div class="skills-list">
				<strong>Skills:</strong>
				` + skillsHTML + `
			</div>

			` + m.renderInventoryHTML(sheet.Inventory) + `

			` + m.renderNotesHTML(sheet.Notes) + `
		</div>
	`
}

// renderVotePromptHTML generates the HTML for a vote prompt.
func (m *OverlayManager) renderVotePromptHTML(prompt VotePrompt) string {
	optionsHTML := ""
	for _, opt := range prompt.Options {
		optionsHTML += `<div class="vote-option" data-option-id="` + opt.ID + `">
			<div class="vote-option-label">` + opt.Label + `</div>
			` + m.maybeRenderDescription(opt.Description) + `
		</div>`
	}

	return `
		<div class="vote-container">
			<div class="vote-question">` + prompt.Question + `</div>
			<div class="vote-options">
				` + optionsHTML + `
			</div>
		</div>
	`
}

// renderDiceResultHTML generates the HTML for dice results.
func (m *OverlayManager) renderDiceResultHTML(result DiceResult) string {
	diceHTML := ""
	for _, roll := range result.Rolls {
		diceHTML += `<div class="dice-individual">
			<div class="dice-roll-value">` + itoa(roll.Result) + `</div>
		</div>`
	}

	modifierStr := ""
	if result.Modifiers > 0 {
		modifierStr = " + " + itoa(result.Modifiers)
	} else if result.Modifiers < 0 {
		modifierStr = " - " + itoa(-result.Modifiers)
	}

	return `
		<div class="dice-container">
			<div class="dice-formula">` + result.RollID + `</div>
			<div class="dice-total">` + itoa(result.Total) + `</div>
			<div class="dice-breakdown">
				` + joinInts(result.Individual, " + ") + modifierStr + `
			</div>
			<div style="margin-top: 16px;">
				` + diceHTML + `
			</div>
		</div>
	`
}

// renderInventoryHTML generates HTML for inventory items.
func (m *OverlayManager) renderInventoryHTML(inventory []Item) string {
	if len(inventory) == 0 {
		return ""
	}

	itemsHTML := ""
	for _, item := range inventory {
		eq := ""
		if item.Equipped {
			eq = " [Equipped]"
		}
		itemsHTML += `<div class="skill-item">
			<span class="skill-name">` + item.Name + ` x` + itoa(item.Quantity) + eq + `</span>
		</div>`
	}

	return `<div class="skills-list">
		<strong>Inventory:</strong>
		` + itemsHTML + `
	</div>`
}

// renderNotesHTML generates HTML for character notes.
func (m *OverlayManager) renderNotesHTML(notes string) string {
	if notes == "" {
		return ""
	}
	return `<div style="margin-top: 16px;">
		<strong>Notes:</strong>
		<p>` + notes + `</p>
	</div>`
}

// maybeRenderDescription returns description HTML if non-empty.
func (m *OverlayManager) maybeRenderDescription(desc string) string {
	if desc == "" {
		return ""
	}
	return `<div class="vote-option-desc">` + desc + `</div>`
}

// applyGlobalStyles injects global CSS styles.
func (m *OverlayManager) applyGlobalStyles() {
	style := m.container.Get("document").Call("createElement", "style")
	style.Set("innerHTML", m.buildGlobalCSS())
	m.container.Get("document").Call("head").Call("appendChild", style)
}

// buildGlobalCSS generates the global CSS string.
func (m *OverlayManager) buildGlobalCSS() string {
	css := ""
	for selector, style := range m.styles {
		if selector == "overlay-container" || selector == "overlay-backdrop" ||
			selector == "overlay-content" || selector == "close-btn" {
			continue // These are applied inline
		}
		css += "." + selector + " { " + style + " }\n"
	}
	return css
}

// parseStyles converts a style string to a JavaScript object.
func parseStyles(s string) js.Value {
	// Styles are applied directly via setProperty calls
	// This is a simplified implementation
	return js.ValueOf(s)
}

// itoa converts an int to string.
func itoa(i int) string {
	if i < 0 {
		return "-" + uitoa(uint(-i))
	}
	return uitoa(uint(i))
}

// uitoa converts an unsigned int to string.
func uitoa(val uint) string {
	if val == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + val - q*10)
		i--
		val = q
	}
	buf[i] = byte('0' + val)
	return string(buf[i:])
}

// joinInts joins integers with a separator.
func joinInts(nums []int, sep string) string {
	if len(nums) == 0 {
		return ""
	}
	result := itoa(nums[0])
	for i := 1; i < len(nums); i++ {
		result += sep + itoa(nums[i])
	}
	return result
}