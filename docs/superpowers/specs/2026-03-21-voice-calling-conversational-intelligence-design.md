# Voice Calling & Conversational Intelligence Design

## Overview

Add Twilio Voice calling (PSTN dial-in) and post-call Conversational Intelligence to Awwdio. The architecture is **audio-first, video-optional**: all calls default to audio, with video as an opt-in enhancement for browser participants.

**Approach**: Video-First with Phone Dial-In. Every call is a Twilio Video Group Room where audio is the primary modality. PSTN participants join via outbound Voice calls bridged into the room using TwiML `<Connect><Room>`. Post-call analysis uses the Intelligence v2 API, with Twilio Sync for real-time pipeline state and Twilio APIs as the source of truth (no local storage).

## Architecture

### Room & Call Model

Every call is a **Twilio Video Group Room** (not P2P, required for PSTN support).

**Room lifecycle**:
1. User clicks "Start Call" -> Backend creates Group Room via Video REST API
2. Returns room name + SID to frontend
3. Browser participants join via Video SDK with audio published by default, video off by default
4. Phone participants join via outbound call -> TwiML `<Connect><Room>`

**Participant modes** (all in the same room):
- **Browser (audio-only)**: Default. Joins Video room, publishes audio track only. Can toggle video on.
- **Browser (audio+video)**: Opts in to video from setup page or during call.
- **Phone (PSTN)**: Joins via outbound call. Audio only, no video capability. Cannot be dominant speaker.

**Room creation** uses Twilio Video API exclusively:
```
POST https://video.twilio.com/v1/Rooms
- UniqueName: room name
- Type: "group"
- RecordParticipantsOnConnect: true
- StatusCallback: BASE_URL/api/webhooks/room-status
```

**Constraints** (from Twilio internal docs):
- Room must exist before PSTN call connects, or call disconnects immediately
- Up to 35 PSTN participants per room
- PSTN audio always routed through US-1 region
- Cannot bridge Voice Conference with Video Room (mutually exclusive)
- PSTN participants cannot be detected as dominant speaker

### Voice Calling (PSTN Dial-In)

**New backend module**: `internal/api/voice/` following the Handler struct + Register() pattern.

**Endpoints**:

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/voice/dial` | POST | JWT | Initiate outbound call to phone number |
| `/api/voice/twiml` | POST | Twilio Signature | TwiML callback returning `<Connect><Room>` |
| `/api/voice/status` | POST | Twilio Signature | Call status webhook |

**Dial flow**:
```
Frontend (setup page "Join via Phone")
  -> POST /api/voice/dial {roomName, phoneNumber}
  -> Backend: verify room exists via Video API
  -> Backend: create outbound call via Voice REST API
      - To: user's phone number
      - From: TWILIO_PHONE_NUMBER
      - Url: BASE_URL/api/voice/twiml?room=RoomName&identity=user-identity
  -> Twilio calls the phone, phone rings
  -> When answered, Twilio fetches TwiML endpoint
  -> TwiML response:
     <Response>
       <Connect>
         <Room participantIdentity="user-identity">RoomName</Room>
       </Connect>
     </Response>
  -> Phone participant appears in the Video room
```

**Webhook authentication**: New `ValidateTwilioSignature()` middleware in `internal/api/middleware/twilio.go`. Uses HMAC-SHA1 validation of `X-Twilio-Signature` header with `TWILIO_AUTH_TOKEN` + full request URL + sorted POST body params. Applied to `/api/voice/twiml`, `/api/voice/status`, and all `/api/webhooks/*` routes.

### Post-Call Intelligence Pipeline

**No local storage**. Twilio APIs are the source of truth. Twilio Sync provides real-time pipeline state to the frontend.

**Pipeline** (webhook-driven):

```
Room Ends
  |
  v
1. Room status webhook fires (RoomEnded event)
  |
  v
2. Backend creates Audio Composition via Video API
   POST https://video.twilio.com/v1/Compositions
   - RoomSid, AudioSources: ["*"], Format: "mp4"
   - StatusCallback: BASE_URL/api/webhooks/composition
  |
  v
3. Backend updates Sync Document: status = "composing"
  |
  v
4. Composition webhook fires -> Backend creates Transcript
   POST https://intelligence.twilio.com/v2/Transcripts
   - ServiceSid: TWILIO_INTELLIGENCE_SERVICE_SID
   - Channel: recording media URL from composition
  |
  v
5. Backend updates Sync Document: status = "transcribing"
  |
  v
6. Transcript webhook fires -> status = "ready"
  |
  v
7. Frontend fetches results from Twilio APIs via backend proxy:
   GET /v2/Transcripts/{TranscriptSid}/Sentences
   GET /v2/Transcripts/{TranscriptSid}/OperatorResults
```

### Twilio Sync for State Management

**Sync Service**: One per deployment, configured via `TWILIO_SYNC_SERVICE_SID`.

**Sync Documents**: One per room, keyed by room name. Schema:
```json
{
  "roomSid": "RMXXX",
  "roomName": "call-abc123",
  "status": "composing | transcribing | ready | failed",
  "compositionSid": "CJXXX",
  "transcriptSid": "GTXXX",
  "participants": [
    {"identity": "user@example.com", "type": "browser"},
    {"identity": "+1555...", "type": "pstn"}
  ],
  "error": null,
  "completedAt": null
}
```

Frontend subscribes to Sync Document updates via `twilio-sync` JS SDK for real-time pipeline status.

### Three Twilio Services, Three Roles

| Service | Role | Data Managed |
|---------|------|-------------|
| Video Compositions API | Creates mixed audio from room recordings | Composition media files |
| Intelligence API (v2) | Transcription + operator analysis | Transcripts, sentences, operator results |
| Twilio Sync | Real-time pipeline state + room metadata | Room-to-transcript mapping, pipeline status |

## API Endpoints

### New Voice Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/voice/dial` | POST | JWT | Initiate outbound PSTN call |
| `/api/voice/twiml` | POST | Twilio Sig | Return TwiML `<Connect><Room>` |
| `/api/voice/status` | POST | Twilio Sig | Call status updates |

### New Webhook Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/webhooks/room-status` | POST | Twilio Sig | Room lifecycle events |
| `/api/webhooks/composition` | POST | Twilio Sig | Composition completion |
| `/api/webhooks/transcript` | POST | Twilio Sig | Transcript completion |

### New Intelligence Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/intelligence/sentences/{roomName}` | GET | JWT | Proxy to Intelligence API sentences |
| `/api/intelligence/results/{roomName}` | GET | JWT | Proxy to Intelligence API operator results |

### New Sync Endpoint

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/sync/token` | GET | JWT | Generate Sync access token for frontend |

## Backend Modules

### New Modules

All follow the Handler struct + NewHandler(cfg) + Register(mux) pattern, registered in `internal/api/api.go`.

- `internal/api/voice/` — Voice calling: dial, TwiML, status handlers
- `internal/api/webhooks/` — Webhook receivers: room-status, composition, transcript
- `internal/api/intelligence/` — Intelligence result proxy endpoints
- `internal/api/sync/` — Sync token generation

### New Middleware

- `internal/api/middleware/twilio.go` — `ValidateTwilioSignature()` middleware for webhook endpoints. Uses HMAC-SHA1 validation, separate from JWT auth.

### Existing Module Changes

- `internal/api/video/` — Add room creation endpoint (POST /api/video/room) that creates Group rooms with recording enabled and status callback URL. Extend existing handler.
- `config/config.go` — Add new environment variables.

## Environment Variables

### New Required Variables

| Variable | Purpose |
|----------|---------|
| `TWILIO_PHONE_NUMBER` | Twilio number to originate outbound calls |
| `BASE_URL` | Public URL for webhook callbacks (e.g., ngrok URL in dev) |
| `TWILIO_AUTH_TOKEN` | For webhook signature validation and Sync |
| `TWILIO_INTELLIGENCE_SERVICE_SID` | Pre-configured Intelligence Service |
| `TWILIO_SYNC_SERVICE_SID` | Sync Service for pipeline state |

### Intelligence Service Setup (One-Time)

1. Create Service via Intelligence v2 API
2. Attach operators: Conversation Summary, Sentiment Analysis
3. Store Service SID in `TWILIO_INTELLIGENCE_SERVICE_SID`

## Frontend Changes

### Setup Page (`/call/[callId]/setup/+page.svelte`)

- Default join method: "Web (audio)" with mic selection
- Toggle: "Turn on camera" (off by default)
- "Join via Phone" option: enter phone number, triggers `POST /api/voice/dial`
- Camera preview is opt-in, not shown by default

### Call Page (`/call/[callId]/+page.svelte`)

- **Audio waveform per participant**: Use Web Audio API `AnalyserNode` on each participant's audio track, render frequency data as animated bars
- Participants shown as name tiles with audio waveform by default
- Video tiles appear only when participants enable camera
- Phone participants: phone icon badge, audio waveform only
- "Invite via Phone" button for adding PSTN participants mid-call

### Post-Call Summary Page (`/call/[callId]/summary/+page.svelte` — new)

- Subscribes to Sync Document for pipeline status
- Progress states: "Recording..." -> "Composing audio..." -> "Analyzing conversation..." -> Results
- Results view:
  - Call summary (from Summary operator)
  - Transcript with speaker labels and timestamps
  - Sentiment indicators per communication
  - Audio playback with transcript sync (composition media URL)

### New Dependencies

- `twilio-sync` (npm, official Twilio package v4.0.0) — frontend Sync subscriptions

### No New Go Dependencies

`twilio-go` v1.25.1 already includes `rest/sync/v1`, `rest/intelligence/v2`, `rest/voice/v1`, `rest/video/v1`.

## Error Handling

### Pipeline Failure Handling

Each webhook step can fail:
- **Composition fails**: StatusCallback includes `composition-failed` event. Update Sync Document: `status: "failed"`, `error: "composition_failed"`. Frontend shows "Analysis unavailable".
- **Transcript fails**: Same pattern — update Sync with failure reason.
- **Webhook delivery missed**: Sync Document remains in intermediate state. Backend provides manual retry endpoint or background reconciliation via Twilio API polling.

### Idempotency

Webhooks may be delivered multiple times. Each handler checks Sync Document state before acting. If composition already triggered transcription, skip duplicate webhook.

### Webhook Authentication

Two separate auth mechanisms:
- **JWT**: User-initiated API calls (browser -> backend)
- **Twilio Signature**: Webhook callbacks (Twilio -> backend). Uses `X-Twilio-Signature` HMAC-SHA1 validation. TwiML and webhook endpoints must NOT use JWT auth.

## Existing Bug Fixes (From Code Review)

These should be addressed alongside the new feature work:

1. **`user?.email` reference bug** in call page (lines 212, 218) — should be `user?.contact`
2. **Content-Type header** not set before error responses in `auth.go`
3. **Duplicate ErrorResponse type** across auth, video, and middleware modules — extract to shared package
4. **Query param inconsistency**: room handler uses `?roomName=` but LEARNINGS.md says `?name=`
5. **Room handler uses `http.Error()`** instead of JSON responses (inconsistent)
6. **No graceful shutdown** in `main.go` (missing signal handling)
7. **Unused `golang-jwt/jwt` dependency** in go.mod — remove

## IPD/OCTO Cross-Reference

Key findings from Twilio internal documentation that informed this design:

- **`docs/video/adding-programmable-voice-participants-video-rooms.mdx`**: Confirmed `<Connect><Room>` TwiML pattern, 35-participant limit, US-1 routing, room-must-exist requirement
- **`docs/video/detecting-dominant-speaker.mdx`**: PSTN participants cannot be dominant speaker
- **`docs/video/overview.mdx`**: Recording and composition pipeline details
- **`docs/platform/conversational-intelligence/`**: v3 API architecture (private beta), quickstart, integration patterns
- **OCTO Architecture Digest (Oct 2025)**: Recordings 2.0 dual-channel architecture, Voice Intelligence domain transition
- **OCTO Domain Integration Patterns**: Eventual consistency, async interactions — applied to webhook pipeline
- **Architecture domain maps**: Voice Intelligence domain definition, SLI journey patterns
- **Product rename**: Voice Intelligence -> Conversational Intelligence
