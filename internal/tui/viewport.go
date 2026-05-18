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
	"math"
	"strings"
)

// Default viewport dimensions matching classic 80x25 terminals.
const (
	ViewportWidth  = 80
	ViewportHeight = 25
)

// Border margins for centering content.
const (
	HorizontalMargin = 2
	VerticalMargin  = 2
)

// Viewport represents a rectangular region in the terminal.
type Viewport struct {
	// Position and dimensions
	X      int
	Y      int
	Width  int
	Height int

	// Actual terminal dimensions
	TermWidth  int
	TermHeight int

	// Decorative border style
	BorderStyle BorderStyle

	// Content offset within viewport (for scrolling)
	ScrollX int
	ScrollY int
}

// BorderStyleNone, BorderStyleSingle, BorderStyleDouble, BorderStyleRounded, BorderStyleHeavy
// defined in model.go to avoid duplication

// DefaultViewport creates a centered 80x25 viewport with decorative borders.
func DefaultViewport(termWidth, termHeight int) Viewport {
	vp := Viewport{
		Width:       ViewportWidth,
		Height:      ViewportHeight,
		TermWidth:   termWidth,
		TermHeight:  termHeight,
		BorderStyle: BorderStyleSingle,
		ScrollX:     0,
		ScrollY:     0,
	}

	// Center the viewport
	vp.X = (termWidth - vp.Width) / 2
	vp.Y = (termHeight - vp.Height) / 2

	// Ensure viewport doesn't go negative
	if vp.X < 0 {
		vp.X = 0
		vp.Width = termWidth
	}
	if vp.Y < 0 {
		vp.Y = 0
		vp.Height = termHeight
	}

	return vp
}

// ContentWidth returns the width of the content area (minus borders).
func (v *Viewport) ContentWidth() int {
	if v.BorderStyle == BorderStyleNone {
		return v.Width
	}
	return v.Width - 2*HorizontalMargin
}

// ContentHeight returns the height of the content area (minus borders).
func (v *Viewport) ContentHeight() int {
	if v.BorderStyle == BorderStyleNone {
		return v.Height
	}
	return v.Height - 2*VerticalMargin
}

// InnerWidth returns the width inside the border (if any).
func (v *Viewport) InnerWidth() int {
	return v.Width - 2
}

// InnerHeight returns the height inside the border (if any).
func (v *Viewport) InnerHeight() int {
	return v.Height - 2
}

// Contains checks if a point is within the viewport.
func (v *Viewport) Contains(x, y int) bool {
	return x >= v.X && x < v.X+v.Width && y >= v.Y && y < v.Y+v.Height
}

// Resize updates the viewport dimensions and re-centers if configured.
func (v *Viewport) Resize(termWidth, termHeight int) {
	v.TermWidth = termWidth
	v.TermHeight = termHeight

	// Recalculate centered position
	v.X = int(math.Max(0, float64((termWidth-v.Width)/2)))
	v.Y = int(math.Max(0, float64((termHeight-v.Height)/2)))
}

// RenderBorder renders the decorative border around the viewport.
// Returns ANSI escape sequences for positioning.
func (v *Viewport) RenderBorder() string {
	if v.BorderStyle == BorderStyleNone {
		return ""
	}

	var b strings.Builder

	topLeft := boxTopLeft
	topRight := boxTopRight
	bottomLeft := boxBottomLeft
	bottomRight := boxBottomRight
	horizontal := boxHorizontal
	vertical := string(boxVertical)

	switch v.BorderStyle {
	case BorderStyleDouble:
		topLeft = boxDoubleTopLeft
		topRight = boxDoubleTopRight
		bottomLeft = boxDoubleBottomLeft
		bottomRight = boxDoubleBottomRight
		horizontal = boxDoubleHorizontal
		vertical = string(boxDoubleVertical)
	}

	// Move to viewport position
	b.WriteString(MoveTo(v.Y+1, v.X+1))

	// Top border
	b.WriteString(topLeft)
	b.WriteString(strings.Repeat(horizontal, v.InnerWidth()))
	b.WriteString(topRight)

	// Side borders
	for i := 0; i < v.InnerHeight(); i++ {
		b.WriteString(MoveTo(v.Y+2+i, v.X+1))
		b.WriteString(vertical)
		b.WriteString(MoveTo(v.Y+2+i, v.X+v.Width))
		b.WriteString(vertical)
	}

	// Bottom border
	b.WriteString(MoveTo(v.Y+v.Height, v.X+1))
	b.WriteString(bottomLeft)
	b.WriteString(strings.Repeat(horizontal, v.InnerWidth()))
	b.WriteString(bottomRight)

	return b.String()
}

// RenderContent renders text content within the viewport, respecting borders.
// Text is truncated/padded to fit the content area and supports scrolling.
func (v *Viewport) RenderContent(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var b strings.Builder
	contentWidth := v.InnerWidth()
	startY := v.Y + 2 + v.ScrollY

	// Limit to visible area
	maxLines := v.InnerHeight()
	if maxLines < 0 {
		maxLines = 0
	}

	for i := 0; i < maxLines && i+v.ScrollY < len(lines); i++ {
		line := lines[i+v.ScrollY]
		// Truncate if too wide
		if len(line) > contentWidth {
			line = line[:contentWidth]
		}
		// Pad if too short
		if len(line) < contentWidth {
			line = line + strings.Repeat(" ", contentWidth-len(line))
		}

		b.WriteString(MoveTo(startY+i, v.X+2))
		b.WriteString(line)
	}

	return b.String()
}

// Scroll moves the viewport content by the given delta.
func (v *Viewport) Scroll(dx, dy int) {
	v.ScrollX = int(math.Max(0, float64(v.ScrollX+dx)))
	v.ScrollY = int(math.Max(0, float64(v.ScrollY+dy)))
}

// ScrollTo moves the viewport to a specific scroll position.
func (v *Viewport) ScrollTo(x, y int) {
	v.ScrollX = int(math.Max(0, float64(x)))
	v.ScrollY = int(math.Max(0, float64(y)))
}

// DecoratedViewport wraps a viewport with additional decorative elements.
type DecoratedViewport struct {
	*Viewport
	Title       string
	Subtitle    string
	Footer      string
	CornerChars CornerChars
	ColorScheme ColorScheme
}

// CornerChars defines custom corner characters for decoration.
type CornerChars struct {
	TL, TR, BL, BR string
}

// ColorScheme defines colors for different viewport elements.
type ColorScheme struct {
	Border      Color256
	Title       Color256
	Subtitle    Color256
	Content     Color256
	Footer      Color256
}

// DefaultColorScheme returns the classic green-on-black retro terminal colors.
func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		Border:   NewColor256(32),   // Dark green
		Title:    NewColor256(34),   // Bright green
		Subtitle: NewColor256(240), // Gray
		Content:  NewColor256(82),  // Light green
		Footer:   NewColor256(244), // Dark gray
	}
}

// NewDecoratedViewport creates a fully decorated viewport with title and borders.
func NewDecoratedViewport(termWidth, termHeight int, title string) DecoratedViewport {
	vp := DefaultViewport(termWidth, termHeight)
	return DecoratedViewport{
		Viewport:   &vp,
		Title:      title,
		CornerChars: CornerChars{
			TL: "◆",
			TR: "◇",
			BL: "◇",
			BR: "◆",
		},
		ColorScheme: DefaultColorScheme(),
	}
}

// Render renders the decorated viewport with all visual elements.
func (d *DecoratedViewport) Render(content []string) string {
	var b strings.Builder

	// Color prefix helpers
	borderColor := d.ColorScheme.Border.Foreground()
	titleColor := d.ColorScheme.Title.Foreground()
	contentColor := d.ColorScheme.Content.Foreground()

	// Render top decorative line with title
	b.WriteString(MoveTo(d.Y+1, d.X+1))
	b.WriteString(borderColor)
	b.WriteString(d.CornerChars.TL)
	b.WriteString(titleColor)
	b.WriteString(" ")
	b.WriteString(d.Title)
	b.WriteString(" ")
	b.WriteString(borderColor)

	// Fill rest of top border
	topWidth := d.Width - lipglossWidth(d.Title) - 4
	if topWidth > 0 {
		b.WriteString(strings.Repeat("▒", topWidth/2))
		b.WriteString(d.CornerChars.TR)
	}

	// Render side borders
	for i := 0; i < d.InnerHeight(); i++ {
		b.WriteString(MoveTo(d.Y+2+i, d.X+1))
		b.WriteString(borderColor)
		b.WriteString(string(boxVertical))
		b.WriteString("\x1b[0m") // Reset

		// Content line
		lineIdx := i + d.ScrollY
		if lineIdx < len(content) {
			b.WriteString(contentColor)
			line := content[lineIdx]
			if len(line) > d.InnerWidth() {
				line = line[:d.InnerWidth()]
			}
			b.WriteString(line)
			b.WriteString(strings.Repeat(" ", d.InnerWidth()-len(line)))
		} else {
			b.WriteString(strings.Repeat(" ", d.InnerWidth()))
		}

		b.WriteString(borderColor)
		b.WriteString(MoveTo(d.Y+2+i, d.X+d.Width))
		b.WriteString(string(boxVertical))
		b.WriteString("\x1b[0m")
	}

	// Render bottom border with footer
	b.WriteString(MoveTo(d.Y+d.Height, d.X+1))
	b.WriteString(borderColor)
	b.WriteString(d.CornerChars.BL)
	b.WriteString(strings.Repeat("─", d.InnerWidth()))
	b.WriteString(d.CornerChars.BR)
	b.WriteString("\x1b[0m")

	// Footer if present
	if d.Footer != "" {
		footerWidth := lipglossWidth(d.Footer)
		footerX := d.X + (d.Width-footerWidth)/2
		b.WriteString(MoveTo(d.Y+d.Height+1, footerX))
		b.WriteString(d.ColorScheme.Footer.Foreground())
		b.WriteString(d.Footer)
		b.WriteString("\x1b[0m")
	}

	return b.String()
}

// lipglossWidth returns the visual width of a string (accounts for ANSI codes).
// This is a simplified version; in production, use charmbracelet/x/ansi.
func lipglossWidth(s string) int {
	// Strip ANSI codes for width calculation
	stripped := stripANSI(s)
	return len(stripped)
}

// stripANSI removes ANSI escape sequences from a string.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape && r == 'm' {
			inEscape = false
			continue
		}
		if !inEscape {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// CenteredText returns a centered text string padded to the given width.
func CenteredText(text string, width int) string {
	padding := (width - lipglossWidth(text)) / 2
	if padding < 0 {
		padding = 0
	}
	return strings.Repeat(" ", padding) + text
}

// RenderCenteredBox renders a centered box with text content.
// This is a convenience function for simple modal-like displays.
func RenderCenteredBox(termWidth, termHeight int, title, content string) string {
	vp := DefaultViewport(termWidth, termHeight)
	lines := strings.Split(content, "\n")

	var b strings.Builder

	// Move to viewport start
	b.WriteString(MoveTo(vp.Y+1, vp.X+1))

	// Top border with title
	titleLen := lipglossWidth(title)
	sideLen := (vp.InnerWidth() - titleLen - 2) / 2

	b.WriteString(boxTopLeft)
	b.WriteString(strings.Repeat(boxHorizontal, sideLen))
	b.WriteString(" ")
	b.WriteString(title)
	b.WriteString(" ")
	b.WriteString(strings.Repeat(boxHorizontal, vp.InnerWidth()-sideLen-titleLen-2))
	b.WriteString(boxTopRight)

	// Content
	for i, line := range lines {
		if i >= vp.InnerHeight() {
			break
		}
		if lipglossWidth(line) > vp.InnerWidth() {
			line = line[:vp.InnerWidth()]
		}
		b.WriteString(MoveTo(vp.Y+2+i, vp.X+1))
		b.WriteString(boxVertical)
		b.WriteString(" ")
		b.WriteString(line)
		b.WriteString(strings.Repeat(" ", vp.InnerWidth()-lipglossWidth(line)-1))
		b.WriteString(boxVertical)
	}

	// Bottom border
	b.WriteString(MoveTo(vp.Y+vp.Height-1, vp.X+1))
	b.WriteString(boxBottomLeft)
	b.WriteString(strings.Repeat(boxHorizontal, vp.InnerWidth()))
	b.WriteString(boxBottomRight)

	return b.String()
}

// FormatDimensions returns a human-readable string of viewport dimensions.
func (v *Viewport) FormatDimensions() string {
	return fmt.Sprintf("%dx%d (terminal: %dx%d, content: %dx%d)",
		v.Width, v.Height,
		v.TermWidth, v.TermHeight,
		v.ContentWidth(), v.ContentHeight())
}