// Package tui provides the TUI foundation for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin built on Bubble Tea v2.
package tui

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/openrole/internal/types"
)

// InventoryDisplay renders the character inventory with weight tracking.
type InventoryDisplay struct {
	Character *types.Character
	Width     int
}

// NewInventoryDisplay creates a new inventory display renderer.
func NewInventoryDisplay(char *types.Character, width int) InventoryDisplay {
	return InventoryDisplay{
		Character: char,
		Width:     width,
	}
}

// Render renders the full inventory display.
func (id InventoryDisplay) Render() string {
	var b strings.Builder

	b.WriteString(id.renderHeader())
	b.WriteString(id.renderWeightStatus())
	b.WriteString(id.renderItemList())
	b.WriteString(id.renderControls())

	return b.String()
}

// renderHeader renders the inventory header box.
func (id InventoryDisplay) renderHeader() string {
	var b strings.Builder
	char := id.Character

	width := id.Width
	innerWidth := width - 4

	// Top border
	b.WriteString(boxTopLeft)
	b.WriteString(strings.Repeat(boxHorizontal, innerWidth))
	b.WriteString(boxTopRight)
	b.WriteString("\n")

	// Title
	title := " INVENTORY "
	padding := strings.Repeat(" ", innerWidth-len(title))
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical, title, padding, boxVertical, boxVertical))

	// Character name
	nameLine := fmt.Sprintf(" Owner: %s", char.Name)
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical,
		strings.Repeat(" ", 1),
		nameLine,
		strings.Repeat(" ", innerWidth-2-len(nameLine)),
		boxVertical))

	return b.String()
}

// renderWeightStatus renders the weight tracking and encumbrance.
func (id InventoryDisplay) renderWeightStatus() string {
	var b strings.Builder
	char := id.Character

	weight := char.InventoryWeight()
	maxWeight := char.EncumbranceMax()
	status, percent := char.EncumbranceStatus()

	weightLine := fmt.Sprintf(" Weight: %.1f/%.1f lbs", weight, maxWeight)
	statusLine := fmt.Sprintf(" [%s]", status)

	// Calculate bar
	totalBars := 10
	filledBars := int(float64(totalBars) * percent / 100)
	if percent > 100 {
		filledBars = totalBars
	}
	bar := strings.Repeat("█", filledBars) + strings.Repeat("░", totalBars-filledBars)

	b.WriteString(fmt.Sprintf("%s %s %s %s %s\n", boxVertical, weightLine, statusLine, bar, boxVertical))
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n", boxVertical, boxLeftT,
		strings.Repeat(boxHorizontal, id.Width-4),
		boxRightT, boxVertical))

	return b.String()
}

// renderItemList renders the list of inventory items.
func (id InventoryDisplay) renderItemList() string {
	var b strings.Builder
	char := id.Character

	// Section header
	b.WriteString(fmt.Sprintf("%s--- ITEMS (%d) ---%s\n", boxVertical, len(char.Inventory), boxVertical))

	if len(char.Inventory) == 0 {
		b.WriteString(fmt.Sprintf("%s Empty%s\n", boxVertical, boxVertical))
	} else {
		for _, item := range char.Inventory {
			equipped := ""
			if item.Equipped {
				equipped = " [EQUIPPED]"
			}
			qty := ""
			if item.Quantity > 1 {
				qty = fmt.Sprintf(" x%d", item.Quantity)
			}
			itemLine := fmt.Sprintf(" • %s%s%s (%.1f lbs)", item.Name, qty, equipped, item.Weight)

			// Truncate if too wide
			maxWidth := id.Width - 4
			if len(itemLine) > maxWidth {
				itemLine = itemLine[:maxWidth]
			}

			b.WriteString(fmt.Sprintf("%s%s%s\n", boxVertical, itemLine, boxVertical))
		}
	}

	return b.String()
}

// renderControls renders the available inventory controls.
func (id InventoryDisplay) renderControls() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s--- ACTIONS ---\n", boxVertical))
	b.WriteString(fmt.Sprintf("%s [E] Equip  [U] Unequip  [D] Drop  [A] Add %s\n", boxVertical, boxVertical))
	b.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
		boxVertical,
		boxBottomLeft,
		strings.Repeat(boxHorizontal, id.Width-4),
		boxBottomRight,
		boxVertical))

	return b.String()
}

// FormatEncumbrance formats a character's encumbrance for display.
func FormatEncumbrance(char *types.Character) string {
	status, percent := char.EncumbranceStatus()
	return fmt.Sprintf("%s (%.0f%%)", status, percent)
}

// ItemByName finds an item in inventory by name (case-insensitive).
func ItemByName(char *types.Character, name string) *types.Item {
	name = strings.ToLower(name)
	for i := range char.Inventory {
		if strings.ToLower(char.Inventory[i].Name) == name {
			return &char.Inventory[i]
		}
	}
	return nil
}

// EquipItem marks an item as equipped.
func EquipItem(char *types.Character, itemName string) bool {
	item := ItemByName(char, itemName)
	if item == nil {
		return false
	}
	item.Equipped = true
	return true
}

// UnequipItem marks an item as not equipped.
func UnequipItem(char *types.Character, itemName string) bool {
	item := ItemByName(char, itemName)
	if item == nil {
		return false
	}
	item.Equipped = false
	return true
}

// AddItem adds an item to the inventory.
func AddItem(char *types.Character, item types.Item) {
	// Check if item with same name exists
	for i := range char.Inventory {
		if strings.ToLower(char.Inventory[i].Name) == strings.ToLower(item.Name) {
			char.Inventory[i].Quantity += item.Quantity
			return
		}
	}
	// If it doesn't exist, add new item
	item.ID = fmt.Sprintf("item-%d", len(char.Inventory)+1)
	char.Inventory = append(char.Inventory, item)
}

// RemoveItem removes an item from the inventory by name.
func RemoveItem(char *types.Character, itemName string) bool {
	name := strings.ToLower(itemName)
	for i := range char.Inventory {
		if strings.ToLower(char.Inventory[i].Name) == name {
			char.Inventory = append(char.Inventory[:i], char.Inventory[i+1:]...)
			return true
		}
	}
	return false
}