// Package tui provides the TUI foundation for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin built on Bubble Tea v2.
package tui

// ColorCycler cycles through colors for spell effects.
// The cycle is: red → orange → yellow → white → red
type ColorCycler struct {
	// Color cycle sequence (ANSI 256-color indices)
	colors [5]int

	// Current position in the cycle
	index int

	// Number of frames between color transitions
	interval    int
	frameCount  int
}

// Spell color cycle constants (ANSI 256-color).
const (
	SpellRed     = 196 // bright red
	SpellOrange  = 214 // orange
	SpellYellow  = 226 // yellow
	SpellWhite   = 231 // white
	SpellPink    = 205 // pink (alternate)
	SpellCyan    = 51  // cyan (alternate)
	SpellPurple  = 93  // purple (alternate)
	SpellGreen   = 46  // green
)

// NewColorCycler creates a color cycler with the default spell effect colors.
// The default cycle is: red → orange → yellow → white → red
func NewColorCycler() *ColorCycler {
	return &ColorCycler{
		colors:   [5]int{SpellRed, SpellOrange, SpellYellow, SpellWhite, SpellRed},
		index:    0,
		interval: 4, // change color every 4 frames (~7.5fps at 30fps)
		frameCount: 0,
	}
}

// NewColorCyclerWithInterval creates a color cycler with a custom interval.
func NewColorCyclerWithInterval(interval int) *ColorCycler {
	c := NewColorCycler()
	c.interval = interval
	return c
}

// Next advances the color cycle by one step.
// Call this once per animation frame.
func (c *ColorCycler) Next() {
	c.frameCount++
	if c.frameCount >= c.interval {
		c.frameCount = 0
		c.index = (c.index + 1) % len(c.colors)
	}
}

// Current returns the current color as a Color256.
func (c *ColorCycler) Current() Color256 {
	return NewColor256(c.colors[c.index])
}

// CurrentColor returns the current color as a Color256.
func (c *ColorCycler) CurrentColor() Color256 {
	return c.Current()
}

// Paint applies the current color to the given string.
func (c *ColorCycler) Paint(s string) string {
	return c.Current().Paint256(s)
}

// Reset returns the color to the start of the cycle.
func (c *ColorCycler) Reset() {
	c.index = 0
	c.frameCount = 0
}

// SetColor sets a specific color at an index in the cycle.
func (c *ColorCycler) SetColor(idx int, color int) {
	if idx >= 0 && idx < len(c.colors) {
		c.colors[idx] = color
	}
}

// SetCustomCycle sets a custom color cycle.
// The slice must have at least 1 element.
func (c *ColorCycler) SetCustomCycle(colors []int) {
	if len(colors) == 0 {
		return
	}
	copy(c.colors[:], colors)
	c.index = 0
}

// CycleSpeed sets how many frames between color transitions.
func (c *ColorCycler) CycleSpeed(frames int) {
	if frames > 0 {
		c.interval = frames
	}
}

// SpellColorCycler is a specialized cycler for spell effects.
// It has pre-configured cycles for different spell schools.
type SpellColorCycler struct {
	ColorCycler
	School SpellSchool
}

// SpellSchool defines color themes for different magic types.
type SpellSchool int

const (
	SchoolFire SpellSchool = iota
	SchoolIce
	SchoolLightning
	SchoolHoly
	SchoolDark
	SchoolArcane
)

// FireSpellColors returns the color cycle for fire spells.
func FireSpellColors() []int {
	return []int{196, 214, 226, 9, 196} // red → orange → yellow → red
}

// IceSpellColors returns the color cycle for ice/cold spells.
func IceSpellColors() []int {
	return []int{195, 189, 159, 231, 195} // cyan → blue → white → cyan
}

// LightningSpellColors returns the color cycle for lightning spells.
func LightningSpellColors() []int {
	return []int{226, 27, 21, 231, 226} // yellow → electric blue → red → white
}

// HolySpellColors returns the color cycle for holy/divine spells.
func HolySpellColors() []int {
	return []int{226, 255, 231, 229, 226} // yellow → white → green → pink
}

// DarkSpellColors returns the color cycle for dark/necromancy spells.
func DarkSpellColors() []int {
	return []int{93, 55, 129, 231, 93} // purple → dark purple → violet → white
}

// ArcaneSpellColors returns the color cycle for arcane spells.
func ArcaneSpellColors() []int {
	return []int{205, 93, 129, 231, 205} // pink → purple → magenta → white
}

// NewSpellColorCycler creates a cycler configured for a spell school.
func NewSpellColorCycler(school SpellSchool) *SpellColorCycler {
	c := &SpellColorCycler{School: school}

	switch school {
	case SchoolFire:
		c.SetCustomCycle(FireSpellColors())
	case SchoolIce:
		c.SetCustomCycle(IceSpellColors())
	case SchoolLightning:
		c.SetCustomCycle(LightningSpellColors())
	case SchoolHoly:
		c.SetCustomCycle(HolySpellColors())
	case SchoolDark:
		c.SetCustomCycle(DarkSpellColors())
	case SchoolArcane:
		c.SetCustomCycle(ArcaneSpellColors())
	default:
		// Default to fire colors
		c.SetCustomCycle(FireSpellColors())
	}

	return c
}

// LerpColor blends between two colors based on t (0.0 to 1.0).
func LerpColor(c1, c2 Color256, t float64) Color256 {
	// Convert to RGB components
	r1 := int(c1) / 36 * 51
	g1 := (int(c1) % 36) / 6 * 51
	b1 := (int(c1) % 6) * 51

	r2 := int(c2) / 36 * 51
	g2 := (int(c2) % 36) / 6 * 51
	b2 := (int(c2) % 6) * 51

	// Lerp
	r := int(float64(r1)*(1-t) + float64(r2)*t)
	g := int(float64(g1)*(1-t) + float64(g2)*t)
	b := int(float64(b1)*(1-t) + float64(b2)*t)

	// Convert back to 256 color index
	return NewColor256(16 + r/51*36 + g/51*6 + b/51)
}

// PulseColor returns a color that pulses between two values based on intensity.
// At t=0 returns c1, at t=1 returns c2, in-between returns the lerp.
func PulseColor(c1, c2 Color256, t float64) Color256 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return LerpColor(c1, c2, t)
}

// CycleDirection controls which way the color cycle moves.
type CycleDirection int

const (
	CycleForward CycleDirection = iota
	CycleBackward
)

// Reverse changes the direction of the color cycle.
func (c *ColorCycler) Reverse() {
	c.index = (len(c.colors) - c.index) % len(c.colors)
}

// Advance moves the cycler forward by n steps without rendering.
func (c *ColorCycler) Advance(steps int) {
	c.index = (c.index + steps) % len(c.colors)
}

// At returns the color at the given index without changing current position.
func (c *ColorCycler) At(idx int) Color256 {
	safeIdx := idx % len(c.colors)
	if safeIdx < 0 {
		safeIdx += len(c.colors)
	}
	return NewColor256(c.colors[safeIdx])
}

// Length returns the number of colors in the cycle.
func (c *ColorCycler) Length() int {
	return len(c.colors)
}