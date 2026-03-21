# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

**Note:** `LEARNINGS.md` and `docs/` are gitignored — do not re-add them to git.

## Project Overview

Awwdio is a fullstack audio/video conversation app built with Go and SvelteKit, using Twilio APIs for video, voice, sync, and conversational intelligence.

**Architecture:**
- **Backend**: Go server — video, voice, sync, intelligence, webhooks, auth
- **Frontend**: SvelteKit app in `web/`, built to `web/build/` (embedded in Go binary)
- **Deployment**: Single compiled Go binary with embedded frontend

Standard library only — no external deps except Twilio SDK.

## Common Commands

```bash
# Full build (frontend + backend)
./build.sh

# Frontend only
cd web && npm run build

# Backend only (embeds web/build/)
go build -o bin/awwdio

# Dev mode (backend only, no frontend embedding)
go run main.go

# Frontend dev server (proxies /api/* to localhost:8080)
cd web && npm run dev
```

## Environment Variables

**Required:**
- `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY`, `TWILIO_API_SECRET`
- `TWILIO_VERIFY_SERVICE_SID` — for OTP authentication
- `JWT_SECRET` — min 32 chars, signs auth tokens

**Optional (voice/webhook features):**
- `TWILIO_AUTH_TOKEN` — webhook signature validation
- `TWILIO_PHONE_NUMBER` — outbound PSTN calls
- `BASE_URL` — webhook callback base URL (e.g. ngrok URL)
- `TWILIO_INTELLIGENCE_SERVICE_SID` — post-call analysis
- `TWILIO_SYNC_SERVICE_SID` — pipeline state management
- `PORT` (default: 8080), `DEBUG=true`, `JSON_LOGGER=true`

```bash
cp sample.env .env && source .env
```

## Architecture

**Backend Structure:**
- `main.go` — HTTP server, three-tier nested mux, slog, embed.FS for frontend
- `config/config.go` — env var loading and validation
- `internal/api/api.go` — registers all module routes
- `internal/api/auth/{auth.go,jwt.go}` — OTP via Twilio Verify + JWT (HS256, crypto/hmac+sha256)
- `internal/api/middleware/auth.go` — JWT validation, sets user in context
- `internal/api/middleware/twilio.go` — Twilio webhook signature validation
- `internal/api/response/` — shared JSON response helper
- `internal/api/video/` — room management and token generation
- `internal/api/voice/` — outbound PSTN dialing and TwiML
- `internal/api/sync/` — Twilio Sync token
- `internal/api/intelligence/` — post-call transcript and operator results
- `internal/api/webhooks/` — room/composition/transcript/sync status callbacks

**Request Flow:** Client → nested mux → optional auth middleware → handler → Twilio API → JSON response

**Three-tier mux routing:** Main mux → `/api/` submux → module submux (`/auth/`, `/video/`, `/voice/`, etc.)

**Frontend Structure:**
- `web/src/lib/api.ts` — fetch wrapper: adds Bearer token, redirects to `/login` on 401
- `web/src/lib/stores/auth.ts` — user auth state (persisted to localStorage)
- `web/src/lib/types.ts` — shared TypeScript types
- `web/src/lib/stores/call.ts`, `sync.ts` — call and sync state
- `web/src/lib/utils/audio.ts`, `phone.ts` — audio level and phone formatting helpers
- `web/src/lib/components/` — ParticipantTile, AudioWaveform, TranscriptViewer, InsightsPanel, PipelineStepper, SyncStatusBadge, InvitePhoneModal

## API Endpoints

| Endpoint | Method | Auth | Notes |
|----------|--------|------|-------|
| `/api/auth/send-otp` | POST | No | `{channel, to}` |
| `/api/auth/verify-otp` | POST | No | `{channel, to, otp}` → `{token}` |
| `/api/video/token` | POST | Yes | `{room}` → `{token}` |
| `/api/video/room` | GET | Yes | `?name=` → room details |
| `/api/video/room` | POST | Yes | `{name}` → create room |
| `/api/voice/dial` | POST | Yes | `{roomName, phoneNumber}` |
| `/api/voice/twiml` | POST | No | TwiML response for Twilio |
| `/api/voice/status` | POST | No | Call status callback |
| `/api/sync/token` | POST | Yes | → `{token}` |
| `/api/intelligence/results/:room` | GET | Yes | Operator results |
| `/api/intelligence/sentences/:room` | GET | Yes | Transcript sentences |
| `/api/webhooks/room-status` | POST | No | Room completion trigger |
| `/api/webhooks/composition` | POST | No | Composition status |
| `/api/webhooks/transcript` | POST | No | Transcript ready |
| `/api/webhooks/sync` | POST | No | Sync document update |

## Pages

1. **`/`** — Landing, Start Call button; redirects to login if unauthenticated
2. **`/login`** — OTP flow: email/phone → verify → display name
3. **`/call/[callId]/setup`** — Two-column pre-call setup:
   - Left: name, email, headphones toggle, join button
   - Right: auto-started video preview + camera/mic/speaker device selectors
   - Headphones toggle sets `echoCancellation` constraint in getUserMedia
   - Video resolution badge reads from `videoTrack.getSettings()` after `loadedmetadata`
4. **`/call/[callId]`** — Active call: participant grid, audio/video toggles, PSTN invite
5. **`/call/[callId]/summary`** — Post-call: PipelineStepper (composing→transcribing→ready via Sync), TranscriptViewer, InsightsPanel

## Development Principles

### Backend

- **Handler pattern**: `type Handler struct` + `NewHandler(cfg)` + `Register(mux)`
- **Adding a new module**: create package → implement Handler → register in `internal/api/api.go`
- **Protecting routes**: wrap submux with `middleware.RequireAuth(secret)(submux)`
- **Webhook handlers**: validate signature via `internal/api/middleware/twilio.go`
- **All JSON responses**: use `internal/api/response` helper
- Use `slog` for logging; early return error handling with clear messages
- No external dependencies beyond Twilio SDK

### Frontend

- **Svelte 5 runes** throughout: `$state`, `$effect`, `$derived`
- **Auth store** is a Svelte writable store — use `.subscribe()` at script top level, not inside `onMount`:
  ```ts
  authStore.subscribe((value) => { user = value; });
  ```
- **API calls**: always use `web/src/lib/api.ts` helpers — they attach the Bearer token and handle 401s
- **Dark mode**: call UI pages use dark theme (`bg-twilio-gray-90`, `bg-twilio-gray-80`, etc.)
- Build with `@sveltejs/adapter-static`; all routes fall back to `index.html`

### Git Safety

- Before any `git reset` or destructive git operation, create a backup branch: `git branch backup/<name>`
- `git reset HEAD` (no `--hard`) is non-destructive — working tree is untouched
- When making multiple unrelated changes, stage and commit each logical group separately
- **Never automatically commit** planning docs, specs, or LEARNINGS.md — always wait for explicit user instruction to commit
- `docs/` and `LEARNINGS.md` are gitignored; do not stage or commit them under any circumstances

## Claude Code Guidance

### Output Token Limit

If you hit `API Error: Claude's response exceeded the 4096 output token maximum`, break work into chunks:
1. Write file skeleton and imports first
2. Implement one section per response
3. Edit incrementally rather than rewriting entire files

### Implementation Guidelines

- Follow Handler struct + Register() pattern from any existing module in `internal/api/`
- All new backend services: config field → `LoadConfig()` entry → `api.go` registration → package with Handler
- Use `internal/api/response` for all JSON responses
- No emojis in code or comments
