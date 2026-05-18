// Package tui provides the TUI foundation for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin built on Bubble Tea v2.
package tui

import (
	"math"
	"strings"
	"time"
)

// TextFlicker provides rapid ANSI bold toggle for text flicker effects.
type TextFlicker struct {
	// Flicker frequency in Hz (times per second)
	Hz int

	// Whether the flicker is currently in "on" state
	On bool

	// Internal timing
	lastTick time.Time
	interval time.Duration
}

// NewTextFlicker creates a text flicker with the given frequency in Hz.
func NewTextFlicker(hz int) TextFlicker {
	if hz <= 0 {
		hz = 12 // default to 12Hz
	}
	return TextFlicker{
		Hz:      hz,
		On:      true,
		lastTick: time.Now(),
		interval: time.Second / time.Duration(hz),
	}
}

// WithHz is a functional option to set flicker frequency.
func (f TextFlicker) WithHz(hz int) TextFlicker {
	if hz > 0 {
		f.Hz = hz
		f.interval = time.Second / time.Duration(hz)
	}
	return f
}

// Tick advances the flicker state based on elapsed time.
// Call this once per animation frame.
func (f *TextFlicker) Tick() {
	now := time.Now()
	elapsed := now.Sub(f.lastTick)
	if elapsed >= f.interval {
		f.On = !f.On
		f.lastTick = now
	}
}

// FlickerOn returns the ANSI bold-on sequence if flicker is on.
func (f *TextFlicker) FlickerOn() string {
	return boldOn
}

// FlickerOff returns the ANSI bold-off sequence.
func (f *TextFlicker) FlickerOff() string {
	return boldOff
}

// Apply applies the flicker effect to a string.
// If flicker is on, returns the string with bold enabled.
func (f *TextFlicker) Apply(s string) string {
	if f.On {
		return boldOn + s + boldOff
	}
	return s
}

// ApplyRapid applies rapid flicker to a string with custom on/off intervals.
// The onInterval and offInterval are in animation frames.
func (f *TextFlicker) ApplyRapid(s string, onFrames, offFrames int) string {
	if onFrames <= 0 {
		onFrames = 1
	}
	if offFrames <= 0 {
		offFrames = 1
	}

	totalFrames := onFrames + offFrames
	frameInCycle := int(f.Hz) % totalFrames

	if frameInCycle < onFrames {
		return boldOn + s + boldOff
	}
	return s
}

// FlickerText wraps a string with flicker styling based on current state.
func (f *TextFlicker) FlickerText(s string) string {
	return f.Apply(s)
}

// Strobe creates a rapid strobe effect (very fast flicker).
func Strobe(s string, rate int) string {
	if rate <= 0 {
		rate = 20 // 20Hz default for strobe
	}
	// Strobe is always at full intensity, alternating as fast as possible
	return boldOn + s + boldOff
}

// GhostFlicker creates a subtle ghost-like flicker effect.
// The text appears and disappears with less harsh transitions.
func (f *TextFlicker) GhostFlicker(s string, intensity float64) string {
	if intensity <= 0 {
		intensity = 0.5
	}
	if intensity > 1 {
		intensity = 1
	}

	// Ghost effect uses inverse instead of bold for softer feel
	if f.On && float64(f.Hz)/20.0 < intensity {
		return inverseOn + s + inverseOff
	}
	return s
}

// DynamicFlicker applies flicker with dynamically changing intensity.
type DynamicFlicker struct {
	BaseFlicker TextFlicker
	Intensity   float64 // 0.0 to 1.0
	Phase       float64 // 0 to 2π for wave-like intensity changes
	PhaseSpeed  float64
}

// NewDynamicFlicker creates a dynamic flicker with sinusoidal intensity.
func NewDynamicFlicker(baseHz int, phaseSpeed float64) *DynamicFlicker {
	return &DynamicFlicker{
		BaseFlicker: NewTextFlicker(baseHz),
		Intensity:   0.5,
		Phase:       0,
		PhaseSpeed:  phaseSpeed,
	}
}

// Tick advances the dynamic flicker state.
func (d *DynamicFlicker) Tick() {
	d.BaseFlicker.Tick()
	d.Phase += d.PhaseSpeed
	// Use sine wave for smooth intensity oscillation between 0.2 and 1.0
	d.Intensity = 0.4*math.Sin(d.Phase) + 0.6
}

// Apply applies the dynamic flicker to a string.
func (d *DynamicFlicker) Apply(s string) string {
	if d.BaseFlicker.On && d.Intensity > 0.3 {
		return boldOn + s + boldOff
	}
	return s
}

// FlickerString applies flicker to multiple strings (e.g., a line of text).
func FlickerString(lines []string, hz int) string {
	f := NewTextFlicker(hz)
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(f.Apply(line))
		b.WriteString("\n")
	}
	return b.String()
}

// IntermittentFlicker creates an effect that flickers in bursts.
type IntermittentFlicker struct {
	TextFlicker
	BurstLength   int // frames per burst
	SilentLength  int // frames between bursts
	BurstCount    int
	FramesInBurst int
	IsBurst       bool
}

// NewIntermittentFlicker creates an intermittent flicker pattern.
func NewIntermittentFlicker(burstLen, silentLen int) IntermittentFlicker {
	if burstLen <= 0 {
		burstLen = 10
	}
	if silentLen <= 0 {
		silentLen = 30
	}
	return IntermittentFlicker{
		BurstLength:  burstLen,
		SilentLength: silentLen,
		BurstCount:   0,
		FramesInBurst: 0,
		IsBurst:      false,
	}
}

// Tick advances the intermittent flicker.
func (f *IntermittentFlicker) Tick() {
	f.TextFlicker.Tick()

	if f.IsBurst {
		f.FramesInBurst++
		if f.FramesInBurst >= f.BurstLength {
			f.IsBurst = false
			f.FramesInBurst = 0
		}
	} else {
		f.FramesInBurst++
		if f.FramesInBurst >= f.SilentLength {
			f.IsBurst = true
			f.FramesInBurst = 0
			f.BurstCount++
		}
	}
}

// Apply applies intermittent flicker - only shows bold during bursts.
func (f *IntermittentFlicker) Apply(s string) string {
	if f.IsBurst && f.TextFlicker.On {
		return boldOn + s + boldOff
	}
	return s
}

// FlickerRate controls how fast the flicker alternates.
type FlickerRate int

const (
	FlickerSlow FlickerRate = 6  // 6 Hz - subtle
	FlickerMedium FlickerRate = 12 // 12 Hz - normal
	FlickerFast FlickerRate = 20 // 20 Hz - rapid
	FlickerStrobe FlickerRate = 30 // 30 Hz - near constant flicker
)

// RateFromHz converts a Hz value to a FlickerRate.
func RateFromHz(hz int) FlickerRate {
	switch {
	case hz <= 6:
		return FlickerSlow
	case hz <= 12:
		return FlickerMedium
	case hz <= 20:
		return FlickerFast
	default:
		return FlickerStrobe
	}
}

// FlickerModifier applies additional visual effects to flickering text.
type FlickerModifier struct {
	Base     TextFlicker
	Color    Color256
	Duration int // total flicker duration in frames
	Frames   int
}

// NewFlickerModifier creates a flicker with a specific duration.
func NewFlickerModifier(hz int, color Color256, duration int) *FlickerModifier {
	return &FlickerModifier{
		Base:     NewTextFlicker(hz),
		Color:    color,
		Duration: duration,
		Frames:   0,
	}
}

// Tick advances the flicker modifier.
func (m *FlickerModifier) Tick() {
	m.Base.Tick()
	m.Frames++
}

// IsActive returns true if the flicker effect is still running.
func (m *FlickerModifier) IsActive() bool {
	return m.Frames < m.Duration
}

// Apply applies the flicker modifier to a string.
func (m *FlickerModifier) Apply(s string) string {
	if !m.IsActive() {
		return s
	}
	flickered := m.Base.Apply(s)
	return m.Color.Paint256(flickered)
}

// RemainingFrames returns how many frames are left in the effect.
func (m *FlickerModifier) RemainingFrames() int {
	return m.Duration - m.Frames
}