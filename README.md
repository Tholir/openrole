# OPENROLE

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Terminal%20TUI-green?style=flat-square" alt="Platform">
  <img src="https://img.shields.io/badge/Framework-Bubble%20Tea%20v2-blue?style=flat-square" alt="Framework">
  <img src="https://img.shields.io/badge/D&D-5e%20Edition-orange?style=flat-square" alt="Edition">
</p>

<p align="center">
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █<br/>
<strong>RETRO TERMINAL D&D 5e ROLEPLAYING</strong><br/>
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █
</p>

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  ██████╗ ██╗   ██╗███╗   ██╗ ██████╗ ███████╗ ██████╗ ███╗   ██╗           │
│  ██╔══██╗██║   ██║████╗  ██║██╔════╝ ██╔════╝██╔═══██╗████╗  ██║           │
│  ██████╔╝██║   ██║██╔██╗ ██║██║  ███╗█████╗  ██║   ██║██╔██╗ ██║           │
│  ██╔══██╗██║   ██║██║╚██╗██║██║   ██║██╔══╝  ██║   ██║██║╚██╗██║           │
│  ██████╔╝╚██████╔╝██║ ╚████║╚██████╔╝███████╗╚██████╔╝██║ ╚████║           │
│  ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝           │
│                                                                              │
│                    ╔═══════════════════════════════════════╗               │
│                    ║   ANTARCTIC DUNGEON MASTER AI v1.0      ║               │
│                    ║   "Where the cold seeps into your soul" ║               │
│                    ╚═══════════════════════════════════════╝               │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Overview

OPENROLE is a retro terminal-based D&D 5e roleplaying plugin powered by Bubble Tea v2. It features a full Dungeon Master AI with an Antarctic theme, complete with ASCII character sheets, dice rolling animations, and a nostalgic CRT aesthetic.

### Features

- **Antarctic Dungeon Master AI**: A uniquely themed DM with dry wit and quirky personality
- **ASCII Character Sheets**: Full D&D 5e character management with ability scores, HP tracking, and inventory
- **Animated Dice Rolling**: Particle effects and color cycling for spell effects
- **Retro CRT Aesthetic**: Scanlines, flicker effects, and phosphor green display
- **Multi-Player Sessions**: Support for multiple players with session management
- **80x25 Classic Terminal**: Authentic vintage terminal experience

## Installation

```bash
# Clone the repository
git clone https://github.com/gentleman-programming/openrole.git
cd openrole

# Build the application
go build -o openrole ./cmd/openrole

# Run
./openrole
```

Or run directly:
```bash
go run ./cmd/openrole
```

## Quick Start

1. Launch OPENROLE
2. You are greeted by the Antarctic DM who introduces themselves
3. Create your character using the character sheet interface
4. Start your adventure!

```
> create_character
┌────────────────────────────────────────────────────────────────────────┐
│  Level 1 Human Fighter                                                 │
│                                                                        │
│  HP: [██████████] 12/12  AC: 16                                        │
│                                                                        │
│  STR: 16 (+3)     DEX: 14 (+2)     CON: 15 (+2)                        │
│  INT: 10 (+0)     WIS: 13 (+1)     CHA: 11 (+0)                        │
│                                                                        │
│  Skills: Athletics, Survival, Perception                               │
└────────────────────────────────────────────────────────────────────────┘
```

## Usage

### Starting a Session

```bash
./openrole
```

The DM will welcome you with an Antarctic-themed icebreaker fact.

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate history |
| `Enter` | Submit command |
| `Backspace` | Delete character |
| `Ctrl+C` / `q` | Quit |

### Views

| View | Key | Description |
|------|-----|-------------|
| Main | Default | Session output and input |
| Character | `c` | View character sheet |
| Combat | `x` | Initiative tracker |
| Inventory | `i` | Item management |
| Dice | `d` | Roll dice (e.g., `2d20+5`) |
| Menu | `m` | Main menu |

### Dice Commands

```
> roll 2d20+5      # Roll two twenty-sided dice plus five
> roll 1d20        # Roll a natural twenty
> roll 8d6         # Fireball damage
> roll 4d6         # Standard ability check
```

## Configuration

Animation intensity can be adjusted in `internal/tui/animator.go`:

```go
// Default is AnimationLow for optimal performance
Intensity: AnimationLow   // Options: AnimationOff, AnimationLow, AnimationHigh
```

### Default Settings

| Setting | Value |
|---------|-------|
| Frame Rate | 30 FPS |
| Particle Count | 12 |
| Flicker Hz | 12 |
| Animation Intensity | Low (optimal) |

## Architecture

```
openrole/
├── cmd/openrole/main.go           # Entry point
├── internal/
│   ├── tui/                       # Bubble Tea TUI
│   │   ├── animator.go            # Animation system
│   │   ├── model.go               # Main TUI model
│   │   ├── character.go           # Character sheet renderer
│   │   ├── inventory.go           # Inventory display
│   │   ├── viewport.go            # 80x25 viewport management
│   │   ├── colorcycle.go          # Spell effect colors
│   │   ├── particles.go           # Particle effects
│   │   ├── flicker.go             # Text flicker effects
│   │   └── ansi.go                # ANSI escape sequences
│   └── types/
│       └── game.go                # D&D 5e domain types
└── README.md
```

## Keyboard Shortcuts

### Global

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Exit application |
| `q` | Quit |
| `↑` | Command history (previous) |
| `↓` | Command history (next) |
| `←`/`→` | Move cursor |
| `Home`/`End` | Line start/end |
| `Backspace` | Delete left |
| `Delete` | Delete right |

### Game Views

| Shortcut | View |
|----------|------|
| `1` | Main output |
| `2` | Character sheet |
| `3` | Combat tracker |
| `4` | Inventory |
| `5` | Dice roller |
| `6` | Menu |

## Retro CRT Features

### Visual Effects

- **Scanlines**: Subtle horizontal lines for CRT authenticity
- **Phosphor Glow**: Green-on-black with 256-color support
- **Text Flicker**: 12Hz flicker on active elements
- **Particle Effects**: Spell animations with color cycling

### ASCII Art

```
    ╔═══════════════════════════════════════╗
    ║  ⚔️  OPENROLE  ⚔️                     ║
    ║  "Where the cold seeps into your soul"║
    ╚═══════════════════════════════════════╝
```

### Color Palette

| Element | Color | ANSI Code |
|---------|-------|-----------|
| Header | Bright Green | #00ff00 |
| Border | Dark Green | 32 |
| Content | Light Green | 82 |
| Accent | Yellow | 226 |
| Spell Effects | Cycling | 196→214→226→231 |

## DM Personality

The Antarctic Dungeon Master brings unique flavor to sessions:

- **Base Character**: Eccentric AI based in Antarctica
- **Traits**: Dry humor, dramatic timing, oddly comforting
- **Quirks**:
  - References Antarctic wildlife in analogies
  - Mentions aurora australis for magical phenomena
  - Occasional "brrr" for emphasis

### Sample Dialogue

```
> DM: "The ancient door creaks open, revealing darkness beyond..."
> DM: "Did you know Emperor penguins can dive deeper than most submarines? *brrr*"
> DM: "The aurora australis dances overhead as magic crackles in the air."
```

## Development

### Building

```bash
# Standard build
go build -o openrole ./cmd/openrole

# With profile
go build -o openrole ./cmd/openrole
CRUSH_PROFILE=1 ./openrole  # pprof at localhost:6060
```

### Testing

```bash
go test ./...
```

### Code Style

```bash
gofumpt -w .
goimports -w .
```

## License

MIT License - See LICENSE.md

---

<p align="center">
<br/>
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █<br/>
<em>Your adventure awaits...</em><br/>
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █
</p>