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
)

// ANSI escape sequence constants.
const (
	escape      = "\x1b"
	reset       = escape + "[0m"
	boldOn      = escape + "[1m"
	boldOff     = escape + "[22m"
	underlineOn = escape + "[4m"
	underlineOff = escape + "[24m"
	inverseOn   = escape + "[7m"
	inverseOff  = escape + "[27m"

	// Cursor control
	cursorHide    = escape + "[?25l"
	cursorShow    = escape + "[?25h"
	cursorHome    = escape + "[H"
	clearScreen   = escape + "[2J"
	clearLine     = escape + "[2K"
	clearLineRight = escape + "[0K"
)

// Standard 16 colors (ANSI escape codes).
const (
	colorBlack        = "0"
	colorRed          = "1"
	colorGreen        = "2"
	colorYellow       = "3"
	colorBlue         = "4"
	colorMagenta      = "5"
	colorCyan         = "6"
	colorWhite        = "7"
	colorDefault      = "9"

	// Bright variants
	brightBlack   = "8"
	brightRed     = "9"
	brightGreen   = "10"
	brightYellow  = "11"
	brightBlue    = "12"
	brightMagenta = "13"
	brightCyan    = "14"
	brightWhite   = "15"
)

// Foreground and background color codes.
type Color struct {
	Code string
}

// Standard colors.
var (
	Black   = Color{colorBlack}
	Red     = Color{colorRed}
	Green   = Color{colorGreen}
	Yellow  = Color{colorYellow}
	Blue    = Color{colorBlue}
	Magenta = Color{colorMagenta}
	Cyan    = Color{colorCyan}
	White   = Color{colorWhite}
	Default = Color{colorDefault}

	// Bright colors.
	BrightBlack   = Color{brightBlack}
	BrightRed     = Color{brightRed}
	BrightGreen   = Color{brightGreen}
	BrightYellow  = Color{brightYellow}
	BrightBlue    = Color{brightBlue}
	BrightMagenta = Color{brightMagenta}
	BrightCyan    = Color{brightCyan}
	BrightWhite   = Color{brightWhite}
)

// Color256 represents a 256-color palette index.
type Color256 int

// NewColor256 creates a 256-color reference (0-255).
func NewColor256(n int) Color256 {
	return Color256(n & 0xff)
}

// NewHexColor converts a hex color string (#RRGGBB) to a 256-color index.
// This uses the standard ANSI 256-color approximation.
func NewHexColor(hex string) Color256 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0
	}

	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)

	// Convert to 256 color index using ANSI approximations
	// 216 colors: 6x6x6 RGB cube (0, 51, 102, 153, 204, 255)
	// 24 grays: 24 step grayscale
	if r == g && g == b {
		// Grayscale: map 0-255 to 232-255 (24 grays)
		gray := (r*24 + 127) / 256
		if gray > 23 {
			gray = 23
		}
		return Color256(232 + gray)
	}

	// RGB cube: 0-5 indices for each component
	ri := (r*6 + 127) / 256
	gi := (g*6 + 127) / 256
	bi := (b*6 + 127) / 256
	return Color256(16 + ri*36 + gi*6 + bi)
}

// Foreground returns the ANSI escape code for foreground color.
func (c Color) Foreground() string {
	return fmt.Sprintf("%s[3%sm", escape, c.Code)
}

// Background returns the ANSI escape code for background color.
func (c Color) Background() string {
	return fmt.Sprintf("%s[4%sm", escape, c.Code)
}

// Foreground256 returns the ANSI escape code for 256-color foreground.
func (c Color256) Foreground() string {
	return fmt.Sprintf("%s[38;5;%dm", escape, c)
}

// Background256 returns the ANSI escape code for 256-color background.
func (c Color256) Background() string {
	return fmt.Sprintf("%s[48;5;%dm", escape, c)
}

// Paint applies color to a string.
func (c Color) Paint(s string) string {
	return c.Foreground() + s + reset
}

// Paint256 applies 256-color to a string.
func (c Color256) Paint256(s string) string {
	return c.Foreground() + s + reset
}

// Bold returns the ANSI escape code for bold.
func Bold(s string) string {
	return boldOn + s + boldOff
}

// Underline returns the ANSI escape code for underline.
func Underline(s string) string {
	return underlineOn + s + underlineOff
}

// Inverse returns the ANSI escape code for inverse.
func Inverse(s string) string {
	return inverseOn + s + inverseOff
}

// Box drawing characters for borders.
const (
	boxTopLeft     = "┌"
	boxTopRight    = "┐"
	boxBottomLeft  = "└"
	boxBottomRight = "┘"
	boxHorizontal  = "─"
	boxVertical    = "│"
	boxCross       = "┼"
	boxTopT        = "┬"
	boxBottomT     = "┴"
	boxLeftT       = "├"
	boxRightT      = "┤"

	// Double-line box drawing
	boxDoubleTopLeft     = "╔"
	boxDoubleTopRight    = "╗"
	boxDoubleBottomLeft  = "╚"
	boxDoubleBottomRight = "╝"
	boxDoubleHorizontal  = "═"
	boxDoubleVertical    = "║"
)

// MakeBorder creates a border string with the given dimensions.
func MakeBorder(width, height int, singleLine bool) string {
	if width < 2 || height < 2 {
		return ""
	}

	topLeft := boxTopLeft
	topRight := boxTopRight
	bottomLeft := boxBottomLeft
	bottomRight := boxBottomRight
	horizontal := boxHorizontal
	vertical := boxVertical

	if !singleLine {
		topLeft = boxDoubleTopLeft
		topRight = boxDoubleTopRight
		bottomLeft = boxDoubleBottomLeft
		bottomRight = boxDoubleBottomRight
		horizontal = boxDoubleHorizontal
		vertical = boxDoubleVertical
	}

	var b strings.Builder
	b.Grow(width*height + height)

	// Top border
	b.WriteString(topLeft)
	b.WriteString(strings.Repeat(horizontal, width-2))
	b.WriteString(topRight)
	b.WriteString("\n")

	// Middle rows
	for i := 0; i < height-2; i++ {
		b.WriteString(vertical)
		b.WriteString(strings.Repeat(" ", width-2))
		b.WriteString(vertical)
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString(bottomLeft)
	b.WriteString(strings.Repeat(horizontal, width-2))
	b.WriteString(bottomRight)

	return b.String()
}

// BoxStyle defines styling for a text box.
type BoxStyle struct {
	Border      bool
	DoubleBorder bool
	Padding     int
	Color       Color256
}

// DefaultBox returns a default box style.
func DefaultBox() BoxStyle {
	return BoxStyle{
		Border: true,
		Color:  NewColor256(34), // Green for retro feel
	}
}

// Render renders text within a box of the given dimensions.
func (s BoxStyle) Render(text string, width, height int) string {
	if width <= 0 || height <= 0 {
		return text
	}

	innerWidth := width - 2*s.Padding - 2
	innerHeight := height - 2*s.Padding - 2

	lines := strings.Split(text, "\n")
	var rendered strings.Builder

	// Top border
	if s.Border {
		h := boxHorizontal
		if s.DoubleBorder {
			h = boxDoubleHorizontal
		}
		border := strings.Repeat(h, innerWidth)
		if s.DoubleBorder {
			rendered.WriteString(boxDoubleTopLeft)
		} else {
			rendered.WriteString(boxTopLeft)
		}
		rendered.WriteString(border)
		if s.DoubleBorder {
			rendered.WriteString(boxDoubleTopRight)
		} else {
			rendered.WriteString(boxTopRight)
		}
		rendered.WriteString("\n")
	}

	// Content lines
	for i := 0; i < innerHeight && i < len(lines); i++ {
		if s.Border {
			if s.DoubleBorder {
				rendered.WriteString(string(boxDoubleVertical))
			} else {
				rendered.WriteString(string(boxVertical))
			}
			rendered.WriteString(strings.Repeat(" ", s.Padding))
		}

		line := lines[i]
		if len(line) > innerWidth {
			line = line[:innerWidth]
		}
		rendered.WriteString(line)
		rendered.WriteString(strings.Repeat(" ", innerWidth-len(line)))

		if s.Border {
			rendered.WriteString(strings.Repeat(" ", s.Padding))
			if s.DoubleBorder {
				rendered.WriteString(string(boxDoubleVertical))
			} else {
				rendered.WriteString(string(boxVertical))
			}
		}
		rendered.WriteString("\n")
	}

	// Bottom border
	if s.Border {
		h := boxHorizontal
		if s.DoubleBorder {
			h = boxDoubleHorizontal
		}
		border := strings.Repeat(h, innerWidth)
		if s.DoubleBorder {
			rendered.WriteString(boxDoubleBottomLeft)
		} else {
			rendered.WriteString(boxBottomLeft)
		}
		rendered.WriteString(border)
		if s.DoubleBorder {
			rendered.WriteString(boxDoubleBottomRight)
		} else {
			rendered.WriteString(boxBottomRight)
		}
	}

	return rendered.String()
}

// Position represents a cursor position.
type Position struct {
	Row    int
	Column int
}

// MoveTo moves the cursor to the specified position (1-indexed).
func MoveTo(row, col int) string {
	return fmt.Sprintf("%s[%d;%dH", escape, row, col)
}

// MoveToHome moves cursor to home position (1,1).
func MoveToHome() string {
	return cursorHome
}

// ClearScreen clears the entire screen.
func ClearScreen() string {
	return clearScreen
}

// ClearLine clears the current line.
func ClearLine() string {
	return clearLine
}

// HideCursor hides the cursor.
func HideCursor() string {
	return cursorHide
}

// ShowCursor shows the cursor.
func ShowCursor() string {
	return cursorShow
}

// SaveCursor saves the current cursor position.
func SaveCursor() string {
	return fmt.Sprintf("%s[s", escape)
}

// RestoreCursor restores the saved cursor position.
func RestoreCursor() string {
	return fmt.Sprintf("%s[u", escape)
}

// ScrollUp scrolls the viewport up by n lines.
func ScrollUp(lines int) string {
	return fmt.Sprintf("%s[%dS", escape, lines)
}

// ScrollDown scrolls the viewport down by n lines.
func ScrollDown(lines int) string {
	return fmt.Sprintf("%s[%dT", escape, lines)
}