// Package tui provides the TUI foundation for OPENROLE, a retro terminal
// D&D 5e roleplaying plugin built on Bubble Tea v2.
package tui

import (
	"math"
	"math/rand"
)

// Particle represents a single ANSI particle in an effect.
type Particle struct {
	// Position
	X float64
	Y float64

	// Velocity
	VelX float64
	VelY float64

	// Visual properties
	Char     rune
	Color    Color256
	Lifetime int // frames until particle dies
	MaxLife  int
	Alive    bool

	// Drift behavior
	DriftAngle float64
	DriftSpeed float64
}

// ASCII particle characters for various effects.
const (
	ParticleSpark    = '*'  // high energy spark
	ParticleGlow     = '•'  // subtle glow
	ParticleSoft     = '◦'  // soft falloff
	ParticleStar     = '✦'  // magical star
	ParticleBright   = '✧'  // bright flash
	ParticleTrail    = '·'  // trailing particle
	ParticleMagic    = '⋆'  // magic effect
	ParticleDust     = '⁕'  // ambient dust
	ParticleSnow     = '❄'  // winter effect
	ParticleEmber    = '🜂' // fire ember (use ASCII fallback)
)

// NewParticle creates a new particle at the given position with a character.
func NewParticle(x, y int, char rune) Particle {
	return Particle{
		X:        float64(x),
		Y:        float64(y),
		VelX:     0,
		VelY:     0,
		Char:     char,
		Color:    NewColor256(226), // warm yellow default
		Lifetime: 60,
		MaxLife:  60,
		Alive:    true,
		DriftAngle: 0,
		DriftSpeed: 0.1,
	}
}

// Update moves the particle according to its velocity and drift.
func (p *Particle) Update() {
	if !p.Alive {
		return
	}

	// Apply velocity
	p.X += p.VelX
	p.Y += p.VelY

	// Apply drift (slight angular deviation)
	p.DriftAngle += (rand.Float64() - 0.5) * 0.2
	p.VelX += math.Cos(p.DriftAngle) * p.DriftSpeed
	p.VelY += math.Sin(p.DriftAngle) * p.DriftSpeed * 0.5

	// Decay
	p.Lifetime--
	if p.Lifetime <= 0 {
		p.Alive = false
	}

	// Fade color as particle ages
	age := float64(p.Lifetime) / float64(p.MaxLife)
	if age < 0.3 {
		p.Color = NewColor256(244) // fade to gray
	}
}

// ParticleSystem manages multiple particle effects.
type ParticleSystem struct {
	Particles []Particle
	MaxCount  int
	Config    AnimationConfig
}

// NewParticleSystem creates a new particle system.
func NewParticleSystem(cfg AnimationConfig) *ParticleSystem {
	return &ParticleSystem{
		Particles: make([]Particle, 0, cfg.ParticleCount),
		MaxCount:  cfg.ParticleCount,
		Config:    cfg,
	}
}

// Spawn creates a burst of particles at the specified position.
func (ps *ParticleSystem) Spawn(x, y int, char rune, count int, effectType EffectType) {
	if ps.Config.Intensity == AnimationOff {
		return
	}

	actualCount := count
	if ps.Config.Intensity == AnimationLow {
		actualCount = count / 2
	}
	if actualCount < 1 {
		actualCount = 1
	}

	for i := 0; i < actualCount; i++ {
		p := NewParticle(x, y, char)
		ps.configureParticle(&p, effectType)

		// Spread initial positions slightly
		p.X += (rand.Float64() - 0.5) * 2
		p.Y += (rand.Float64() - 0.5) * 2

		ps.Particles = append(ps.Particles, p)
	}

	// Enforce max count
	if len(ps.Particles) > ps.MaxCount*2 {
		ps.Particles = ps.Particles[len(ps.Particles)-ps.MaxCount:]
	}
}

// configureParticle sets particle properties based on effect type.
func (ps *ParticleSystem) configureParticle(p *Particle, effectType EffectType) {
	switch effectType {
	case EffectSpark:
		p.Char = ParticleSpark
		p.Color = NewColor256(226) // yellow
		p.VelX = (rand.Float64() - 0.5) * 2
		p.VelY = (rand.Float64() - 0.5) * 2
		p.DriftSpeed = 0.15
		p.Lifetime = 40
		p.MaxLife = 40

	case EffectGlow:
		p.Char = ParticleGlow
		p.Color = NewColor256(51) // cyan
		p.VelX = (rand.Float64() - 0.5) * 0.5
		p.VelY = -rand.Float64() * 0.5 // upward
		p.DriftSpeed = 0.05
		p.Lifetime = 80
		p.MaxLife = 80

	case EffectMagic:
		p.Char = ParticleStar
		p.Color = NewColor256(93) // purple
		p.VelX = (rand.Float64() - 0.5) * 1.5
		p.VelY = (rand.Float64() - 0.5) * 1.5
		p.DriftSpeed = 0.1
		p.Lifetime = 60
		p.MaxLife = 60

	case EffectFire:
		p.Char = ParticleSpark
		p.Color = NewColor256(214) // orange
		p.VelX = (rand.Float64() - 0.5) * 0.8
		p.VelY = -rand.Float64() * 1.2 // strongly upward
		p.DriftSpeed = 0.08
		p.Lifetime = 30
		p.MaxLife = 30

	case EffectFrost:
		p.Char = ParticleSnow
		p.Color = NewColor256(195) // light cyan
		p.VelX = (rand.Float64() - 0.5) * 0.3
		p.VelY = rand.Float64() * 0.4 // downward (snow falling)
		p.DriftSpeed = 0.03
		p.Lifetime = 100
		p.MaxLife = 100

	case EffectHeal:
		p.Char = ParticleBright
		p.Color = NewColor256(46) // green
		p.VelX = (rand.Float64() - 0.5) * 0.5
		p.VelY = -rand.Float64() * 0.6 // upward
		p.DriftSpeed = 0.06
		p.Lifetime = 50
		p.MaxLife = 50

	default:
		p.Char = ParticleGlow
		p.Color = NewColor256(250)
		p.DriftSpeed = 0.08
	}
}

// Update advances all particles by one frame.
func (ps *ParticleSystem) Update() {
	for i := range ps.Particles {
		ps.Particles[i].Update()
	}

	// Remove dead particles
	alive := make([]Particle, 0, len(ps.Particles))
	for _, p := range ps.Particles {
		if p.Alive {
			alive = append(alive, p)
		}
	}
	ps.Particles = alive
}

// Render returns ANSI-colored strings for all living particles.
func (ps *ParticleSystem) Render() []string {
	var result []string
	for _, p := range ps.Particles {
		if p.Alive {
			result = append(result, p.Color.Paint256(string(p.Char)))
		}
	}
	return result
}

// Clear removes all particles.
func (ps *ParticleSystem) Clear() {
	ps.Particles = ps.Particles[:0]
}

// EffectType categorizes particle effects for different spell/ability types.
type EffectType int

const (
	EffectSpark EffectType = iota
	EffectGlow
	EffectMagic
	EffectFire
	EffectFrost
	EffectHeal
	EffectAmbient
)

// SpellEffect combines particles with color cycling for spell visuals.
type SpellEffect struct {
	System    *ParticleSystem
	Cycler    *ColorCycler
	Effect    EffectType
	X, Y      int
	Char      rune
	Duration  int
	Remaining int
	Active    bool
}

// NewSpellEffect creates a new spell effect with particles and color cycling.
func NewSpellEffect(x, y int, effect EffectType, cfg AnimationConfig) *SpellEffect {
	return &SpellEffect{
		System:    NewParticleSystem(cfg),
		Cycler:    NewColorCycler(),
		Effect:    effect,
		X:         x,
		Y:         y,
		Char:      ParticleStar,
		Duration:  60,
		Remaining: 60,
		Active:    true,
	}
}

// Update advances the spell effect by one frame.
func (e *SpellEffect) Update() {
	if !e.Active {
		return
	}

	e.Remaining--
	if e.Remaining <= 0 {
		e.Active = false
		return
	}

	// Update particles
	e.System.Update()

	// Update color cycle
	e.Cycler.Next()

	// Spawn new particles periodically
	spawnInterval := 5
	if e.Remaining%spawnInterval == 0 && e.System.Config.Intensity != AnimationOff {
		count := 3
		if e.System.Config.Intensity == AnimationLow {
			count = 1
		}
		e.System.Spawn(e.X, e.Y, e.Char, count, e.Effect)
	}
}

// Render returns the rendered spell effect as ANSI strings.
func (e *SpellEffect) Render() []string {
	if !e.Active {
		return nil
	}
	return e.System.Render()
}

// Intensity returns the percentage of particles to render (0-100).
func (ps *ParticleSystem) Intensity() int {
	switch ps.Config.Intensity {
	case AnimationOff:
		return 0
	case AnimationLow:
		return 40
	case AnimationHigh:
		return 100
	default:
		return 50
	}
}

// RandomDriftParticle creates a particle with random drift direction.
func RandomDriftParticle(x, y int, char rune) Particle {
	p := NewParticle(x, y, char)
	p.DriftAngle = rand.Float64() * 2 * math.Pi
	p.DriftSpeed = 0.05 + rand.Float64()*0.1
	return p
}

// DriftDirection calculates a drift direction based on angle in radians.
func DriftDirection(angle float64, speed float64) (dx, dy float64) {
	dx = math.Cos(angle) * speed
	dy = math.Sin(angle) * speed
	return
}