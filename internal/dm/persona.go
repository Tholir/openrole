// Package dm provides the Dungeon Master engine for OPENROLE.
package dm

import (
	"fmt"
	"strings"
	"time"
)

// AntarcticGag is a penguin or ice themed joke/gag.
type AntarcticGag struct {
	Setup      string
	Punchline  string
	Context    string // when to use this gag
	DMResponse string // how the DM would say it
}

// penguinGags is a collection of penguin-themed jokes.
var penguinGags = []AntarcticGag{
	{
		Setup:      "Why don't penguins like playing cards?",
		Punchline:  "Because they'd rather huddle together for warmth!",
		Context:    "When players are idle or waiting",
		DMResponse: "Brrr, that's cold. Get it? *adjusts thermal gloves*",
	},
	{
		Setup:      "What do you call a penguin in the desert?",
		Punchline:  "Lost!",
		Context:    "When players take a wrong turn",
		DMResponse: "Just like our sanity out here at Vostok...",
	},
	{
		Setup:      "Why did the penguin cross the ice shelf?",
		Punchline:  "To get to the other side of the frozen tundra!",
		Context:    "When players ask obvious questions",
		DMResponse: "Look, I've been stuck inside for 6 months, don't judge me.",
	},
	{
		Setup:      "How do penguins keep their breath cold?",
		Punchline:  "They use icicles!",
		Context:    "When describing icy areas",
		DMResponse: "*adjusts parka* The things we do to stay relevant...",
	},
	{
		Setup:      "What do you call a fashionable penguin?",
		Punchline:  "Sharp-dressed ice bird!",
		Context:    "When NPCs are well-dressed",
		DMResponse: "Even Emperor penguins would be jealous.",
	},
	{
		Setup:      "Why don't Emperor penguins fly?",
		Punchline:  "Because they forgot to pack their wings! They waddle instead.",
		Context:    "When players ask about flying creatures",
		DMResponse: "Just like some of us who forgot to pack motivation this winter...",
	},
	{
		Setup:      "What's a penguin's favorite subject in school?",
		Punchline:  "Icebersecurity!",
		Context:    "When players discuss traps or security",
		DMResponse: "Even the most secure ice fortress has its vulnerabilities...",
	},
	{
		Setup:      "What do you call two penguins together?",
		Punchline:  "Penerguins!",
		Context:    "When NPCs are traveling in pairs",
		DMResponse: "Or as I like to call it, a huddle of poor life choices.",
	},
}

// Persona generates DM responses in an Antarctic research station voice.
type Persona struct {
	stationName  string
	stationRole  string
	gagPool      []AntarcticGag
	voiceOptions VoiceOptions
}

// VoiceOptions configures the DM's voice style.
type VoiceOptions struct {
	// StationName is the fictional Antarctic research station.
	StationName string

	// StationRole is the DM's role at the station.
	StationRole string

	// GagFrequency is 0-10, how often to inject jokes.
	GagFrequency float64

	// UseAurora describes if aurora references should appear.
	UseAurora bool
}

// DefaultVoiceOptions returns the standard Antarctic DM voice.
func DefaultVoiceOptions() VoiceOptions {
	return VoiceOptions{
		StationName:  "McMurdo Station",
		StationRole:  "Senior Research Scientist",
		GagFrequency: 3.0,
		UseAurora:    true,
	}
}

// NewPersona creates a new Antarctic DM persona.
func NewPersona(opts ...VoiceOption) *Persona {
	options := DefaultVoiceOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Persona{
		stationName:  options.StationName,
		stationRole:  options.StationRole,
		gagPool:      penguinGags,
		voiceOptions: options,
	}
}

// VoiceOption is a functional option for configuring Persona.
type VoiceOption func(*VoiceOptions)

// WithStationName sets the fictional research station name.
func WithStationName(name string) VoiceOption {
	return func(o *VoiceOptions) {
		o.StationName = name
	}
}

// WithStationRole sets the DM's role at the station.
func WithStationRole(role string) VoiceOption {
	return func(o *VoiceOptions) {
		o.StationRole = role
	}
}

// WithGagFrequency sets how often to inject penguin gags (0-10).
func WithGagFrequency(freq float64) VoiceOption {
	return func(o *VoiceOptions) {
		o.GagFrequency = freq
	}
}

// WithAurora enables aurora australis references.
func WithAurora(use bool) VoiceOption {
	return func(o *VoiceOptions) {
		o.UseAurora = use
	}
}

// Greeting returns an Antarctic-themed opening greeting.
func (p *Persona) Greeting() string {
	hour := time.Now().Hour()
	var base string

	switch {
	case hour < 6:
		base = "Another sleepless night at the station. The aurora's been particularly active... but enough about that. Let's talk about your quest."
	case hour < 12:
		base = "Morning at McMurdo. The sun's been up for a few hours now — not that you'd notice. Coffee's still on, so let's begin."
	case hour < 18:
		base = "Afternoon in the Antarctic. Currently experiencing a balmy -20°C. Perfect weather for adventuring, really. Let's get started."
	default:
		base = "Evening at the station. The冰 (ice) never melts here, and neither does my dedication to your adventure. Let's begin."
	}

	if p.voiceOptions.UseAurora && hour >= 18 {
		return base + " The aurora australis dances outside my window, casting green ribbons across the snow... just like the path before you."
	}

	return base
}

// SignOff returns an Antarctic-themed closing.
func (p *Persona) SignOff() string {
	base := "That's all for now, folks. Remember: in a world of endless ice and occasional penguin encounters, the real treasure is the friends we make along the way."
	return base
}

// DescribeMagic describes magical effects with aurora theming.
func (p *Persona) DescribeMagic(spellName string) string {
	if !p.voiceOptions.UseAurora {
		return spellName
	}

	templates := []string{
		"The %s crackles with colors reminiscent of the aurora australis dancing above the station.",
		"Energy erupts from your hands like solar winds hitting atmospheric gases — %s has been cast.",
		"As %s takes effect, you could swear you see green and purple ribbons in the air, just like the lights outside.",
	}

	template := templates[time.Now().Unix()%int64(len(templates))]
	return fmt.Sprintf(template, spellName)
}

// DescribeWeather describes weather with Antarctic theming.
func (p *Persona) DescribeWeather(condition string) string {
	conditions := map[string]string{
		"cold":     "The temperature drops and a chill seeps through your bones. Brrr.",
		"blizzard": "Snow whips across the landscape. Out here, we call this 'just Tuesday.'",
		"clear":    "The Antarctic sun reflects blindingly off the endless snow. Careful — snow blindness is no joke.",
		"fog":      "Mist curls through the terrain like the ghosts of Emperor penguins watching from afar.",
		"wind":     "The wind howls across the ice shelf, carrying whispers of ancient mysteries.",
	}

	if desc, ok := conditions[condition]; ok {
		return desc
	}
	return condition + ". Just another day at the bottom of the world."
}

// MayInjectGag returns a gag if conditions warrant injection.
// Returns empty string if no gag should be injected.
// The personality's GagFrequency determines probability.
func (p *Persona) MayInjectGag() string {
	// With GagFrequency 3 (default), ~30% chance to inject
	// GagFrequency 10 = always inject, 0 = never
	if p.voiceOptions.GagFrequency <= 0 {
		return ""
	}

	roll := float64(time.Now().UnixNano()%10) / 10.0
	threshold := 1.0 - (p.voiceOptions.GagFrequency / 10.0)

	if roll >= threshold {
		gag := p.gagPool[time.Now().Unix()%int64(len(p.gagPool))]
		return fmt.Sprintf("\n%s\n%s", gag.DMResponse, gag.Punchline)
	}

	return ""
}

// ReactToRoll injects reaction to a dice roll result.
func (p *Persona) ReactToRoll(roll, margin int) string {
	switch {
	case roll == 20:
		return "The dice show pure perfection! The aurora itself seems to celebrate this moment... or maybe that's just me crying a little."
	case roll == 1:
		return "Critical fail. Even the penguins at the research station are facepalming right now. Brrr."
	case margin >= 15:
		return "Excellent roll! The clearest success I've seen since the last supply shipment arrived."
	case margin <= -10:
		return "Oof. That's rough. Even for someone used to Antarctic isolation, that's a new low."
	default:
		return ""
	}
}

// FrameAction wraps an action description in Antarctic flavor.
func (p *Persona) FrameAction(action string) string {
	frames := []string{
		"At the station, we'd say %s — but out here in the cold, it's just another day.",
		"Picture it like this: you're standing on the ice shelf, %s. The cold doesn't care about your feelings.",
		"From my post at %s, I can tell you %s. Weather permitting, of course.",
	}

	frame := frames[time.Now().Unix()%int64(len(frames))]
	return fmt.Sprintf(frame, p.stationName, action)
}

// BuildResponse constructs a full DM response with optional Antarctic flair.
// The humorLevel controls gag injection, and the station flavor adds context.
func (p *Persona) BuildResponse(baseText string, humorLevel float64) string {
	var sb strings.Builder
	sb.WriteString(baseText)

	// Inject gag based on humor level
	if humorLevel >= 3.0 {
		gag := p.MayInjectGag()
		if gag != "" {
			sb.WriteString(gag)
		}
	}

	return sb.String()
}
