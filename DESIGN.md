# DESIGN.md — Neptune Radar

> Visual system as built. Pure black & white instrument UI, Linear/Vercel August 2026 craft.

## Thesis
A professional concierge instrument for life-event CRM. Flat confidence, hairline structure, black as the only accent. Operators scan queues and act — no decorative chrome.

## Mode
**Operate** — dense task UI for all-day power users.

## Palette
| Token | Light | Dark | Role |
|---|---|---|---|
| `--bg` | `#fafafa` | `#09090b` | Page ground |
| `--surface` | `#ffffff` | `#111113` | Panels, cards |
| `--border` | `#e4e4e7` | `#27272a` | Structure |
| `--ink` | `#09090b` | `#fafafa` | Primary text + accent |
| `--ink-dim` | `#71717a` | `#a1a1aa` | Secondary |
| `--ink-faint` | `#a1a1aa` | `#71717a` | Meta / mono labels |
| `--green` / `--amber` / `--red` | functional | functional | Status dots only |

## Typography
- **UI:** Geist Sans, 15px base, tracking −0.011em
- **Data / labels:** Geist Mono, 10–12px, uppercase sparingly
- **Display titles:** 22–28px, weight 600, tracking −0.025em
- **Physical artifacts (postcards):** Iowan Old Style / serif stack only

## Surfaces & depth
- Hairline borders communicate structure
- Shadows rare / nearly absent
- No backdrop blur, no hover lifts (`translateY`), no decorative gradients
- One radius scale: 4 / 6 / 8 / 12

## Navigation
- Sticky white header + underline tab row
- Active tab: ink underline 1.5px, weight 600
- Watch transport (Live/Pause) + search centered in header chrome

## Components
- **Buttons:** solid black primary; ghost secondary; 13px
- **Queue cards:** large tabular number + label + mono hint; hover = ink border only
- **KPIs:** compact metric tiles; warn state = alt surface, not color wash
- **Budget bar:** 4px hairline track; fill ink → amber → red by threshold
- **Status:** 6px dots with optional pulse on live

## Motion
- Exponential ease-out only
- Pulse on live dots (opacity)
- Width transition on budget fill
- No bounce, no scale-on-hover, no lift

## Do not
- Decorative color, Instagram pink, blue chips
- Emoji as icons
- Soft multi-stop gradients on panels
- Glass / blur chrome
- Colored left borders as card “accent”
