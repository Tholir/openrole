// Package tui provides the TUI foundation for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin built on Bubble Tea v2.
package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// AnimationIntensity controls the level of animation effects.
type AnimationIntensity int

const (
	AnimationOff AnimationIntensity = iota
	AnimationLow
	AnimationHigh
)

// AnimationConfig holds animation system configuration.
type AnimationConfig struct {
	Intensity     AnimationIntensity
	ParticleCount int // number of particles per effect
	FrameRate     int // target FPS (default 30)
	FlickerHz     int // flicker frequency for text effects
}

// DefaultAnimationConfig returns the default animation settings.
func DefaultAnimationConfig() AnimationConfig {
	return AnimationConfig{
		Intensity:     AnimationLow,
		ParticleCount: 12,
		FrameRate:     30,
		FlickerHz:     12,
	}
}

// TickDuration returns the time interval between animation frames.
func (c AnimationConfig) TickDuration() time.Duration {
	return time.Second / time.Duration(c.FrameRate)
}

// Animator manages the 30fps animation frame loop using Bubble Tea tick messages.
type Animator struct {
	Config      AnimationConfig
	FrameCount  int64
	Running     bool
	Particles   []Particle
	ColorCycle  *ColorCycler
	Flicker     TextFlicker
	ticker      *time.Ticker
	done        chan struct{}
}

// NewAnimator creates a new animator with the given configuration.
func NewAnimator(cfg AnimationConfig) *Animator {
	return &Animator{
		Config:     cfg,
		FrameCount: 0,
		Running:    false,
		Particles:  make([]Particle, 0, cfg.ParticleCount),
		ColorCycle: NewColorCycler(),
		Flicker:    NewTextFlicker(cfg.FlickerHz),
	}
}

// Start begins the animation loop, returning a tea.Cmd that produces tick messages.
func (a *Animator) Start() tea.Cmd {
	if a.Running {
		return nil
	}
	a.Running = true
	a.ticker = time.NewTicker(a.TickDuration())
	a.done = make(chan struct{})

	return a.tickerLoop()
}

// tickerLoop is an internal coroutine that sends tick messages at 30fps.
func (a *Animator) tickerLoop() tea.Cmd {
	return func() tea.Msg {
		for a.Running {
			select {
			case <-a.ticker.C:
				a.FrameCount++
				a.tick()
			case <-a.done:
				a.ticker.Stop()
				return nil
			}
		}
		return nil
	}
}

// tick advances the animation state by one frame.
func (a *Animator) tick() {
	// Update particles
	for i := range a.Particles {
		a.Particles[i].Update()
	}

	// Update color cycle
	a.ColorCycle.Next()

	// Update flicker
	a.Flicker.Tick()
}

// Stop halts the animation loop.
func (a *Animator) Stop() {
	if a.Running {
		a.Running = false
		close(a.done)
	}
}

// TickDuration returns the frame interval based on config.
func (a *Animator) TickDuration() time.Duration {
	return a.Config.TickDuration()
}

// IsActive returns true if animations should be rendered based on intensity setting.
func (a *Animator) IsActive() bool {
	return a.Config.Intensity != AnimationOff
}

// IsHigh returns true if high-intensity animations are enabled.
func (a *Animator) IsHigh() bool {
	return a.Config.Intensity == AnimationHigh
}

// SpawnParticles creates a burst of particles at the given position.
func (a *Animator) SpawnParticles(x, y int, char rune, count int) {
	if !a.IsActive() || count <= 0 {
		return
	}

	actualCount := count
	if a.Config.Intensity == AnimationLow {
		actualCount = count / 2
	}
	if actualCount < 1 {
		actualCount = 1
	}

	for i := 0; i < actualCount; i++ {
		p := NewParticle(x, y, char)
		p.VelX = float64((i % 3) - 1) * 0.5    // drift left/right
		p.VelY = -float64(i%3) * 0.3           // slight upward bias
		a.Particles = append(a.Particles, p)
	}

	// Trim to max capacity
	if len(a.Particles) > a.Config.ParticleCount*2 {
		a.Particles = a.Particles[len(a.Particles)-a.Config.ParticleCount:]
	}
}

// RenderParticles returns the current particle states as colored strings.
func (a *Animator) RenderParticles() []string {
	if !a.IsActive() || len(a.Particles) == 0 {
		return nil
	}

	var lines []string
	for _, p := range a.Particles {
		if p.Alive {
			color := a.ColorCycle.CurrentColor()
			lines = append(lines, color.Paint256(string(p.Char)))
		}
	}
	return lines
}

// ClearParticles removes all active particles.
func (a *Animator) ClearParticles() {
	a.Particles = a.Particles[:0]
}

// FrameMsg is sent each animation frame.
type FrameMsg struct {
	Frame    int64
	Particles []Particle
}

// AnimatorMsg is the base message type for animator events.
type AnimatorMsg struct {
	Type    string
	Payload any
}

// AnimatorStartMsg signals the animator to begin.
type AnimatorStartMsg struct{}

// AnimatorStopMsg signals the animator to stop.
type AnimatorStopMsg struct{}

// ApplyIntensity applies animation intensity to a model.
func ApplyIntensity(m *Model, intensity AnimationIntensity) {
	if m.Animator == nil {
		m.Animator = NewAnimator(DefaultAnimationConfig())
	}
	m.Animator.Config.Intensity = intensity
}