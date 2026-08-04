# Neptune Growth Organism

Agentic operating system that turns Meet Neptune (meetneptune.com) from a
destination people visit into a system that shows up when a couple’s world
changes — **celebrate first, align second, dual counsel last**.

## Thesis

```
Scout → Detective → Concierge → Risk Sentinel → Counselor Bridge → Learn
```

Agents propose. Policy cages. Humans approve anything customer-facing.

## Hard guarantees

1. **No pitch on relationship-risk** — unfollow/state-change never becomes a prenup pitch
2. **Celebrate first** — first touch is congratulations only (postcard / soft note)
3. **Human in the loop** — pending actions require operator approve
4. **Dual counsel path** — handoffs land in app.meetneptune.com/chat (both partners get lawyers)

## Surfaces

| Surface | Path |
|---------|------|
| Operator briefing | Today → Growth organism strip |
| Full organism | More → Organism (`#/organism`) |
| API | `GET /api/organism` |
| Closed-loop webhooks | `POST /api/webhooks/neptune` |
| Postcard QR | Kit postcard HTML → celebrate deep link |

## Postcard deep link

Printed cards include a QR into Meet Neptune chat with:

`utm_source=neptune_radar&utm_medium=postcard&utm_campaign=celebrate_first`

Analytics separates celebrate mail from soft-invite handoffs (`utm_medium=handoff`).

## Close the loop

Meet Neptune product events (chat_started, consult_booked, closed_won, closed_lost)
flow back via webhook → funnel_events → Organism yield board (by market + source).

## Env

- `NEPTUNE_CHAT_BASE_URL` — default `https://app.meetneptune.com/chat`
- `NEPTUNE_WEBHOOK_SECRET` — webhook auth for funnel events
