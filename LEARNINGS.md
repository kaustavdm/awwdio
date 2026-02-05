# LEARNINGS.md

Technical patterns and implementation details for Awwdio. Update as you work.

## Key Patterns

**Backend (Go):**
- Handler struct + `NewHandler()` + `Register(mux)` pattern for API modules
- Three-tier mux routing: Main → API (`/api/`) → Module (`/auth/`, `/video/`)
- Use `slog` for logging, early return error handling
- **Standard library only** - no external deps except Twilio SDK
- JWT: `internal/api/auth/jwt.go` uses crypto/hmac + sha256 (HS256)
- Auth middleware: `internal/api/middleware/auth.go` - validates JWT, sets user in context

**Frontend (SvelteKit):**
- Svelte 5 with `$state` runes
- Auth store: `web/src/lib/stores/auth.ts`
- API helper: `web/src/lib/api.ts` - adds Bearer token, handles 401→login redirect
- Build output: `web/build/` (embedded in Go binary)

## API Endpoints

| Endpoint | Method | Auth | Request | Response |
|----------|--------|------|---------|----------|
| `/api/auth/send-otp` | POST | No | `{channel, to}` | `{success}` |
| `/api/auth/verify-otp` | POST | No | `{channel, to, otp}` | `{success, token}` |
| `/api/video/token` | POST | Yes | `{room}` | `{token}` |
| `/api/video/room` | GET | Yes | `?name=` | Room details |

## Environment Variables

**Required:**
- `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY`, `TWILIO_API_SECRET`
- `TWILIO_VERIFY_SERVICE_SID` - for OTP
- `JWT_SECRET` - min 32 chars

**Optional:**
- `PORT` (default: 8080)
- `DEBUG=true` - verbose logging
- `JSON_LOGGER=true` - JSON log format

## Build

```bash
./build.sh           # Full build (frontend + backend)
cd web && npm run build  # Frontend only
go build -o bin/awwdio   # Backend only (embeds web/build/)
```

## Adding New API Module

1. Create `internal/api/newmodule/newmodule.go`
2. Implement: `type Handler struct`, `NewHandler(cfg)`, `Register(mux)`
3. Register in `internal/api/api.go`
4. Apply auth middleware if protected: `middleware.RequireAuth(secret)(mux)`

## Pending Features

- Room management API (create, list, delete, participants)
- Token refresh mechanism
- Phone/PSTN bridge (Twilio Voice)
- Rate limiting, CORS, CSRF protection

## File Structure

```
main.go                          # Server, embeds frontend
config/config.go                 # Env var loading
internal/api/
  api.go                         # Router setup
  auth/{auth.go,jwt.go}          # OTP + JWT
  middleware/auth.go             # JWT validation
  video/{video.go,access_token.go,room.go}
web/src/
  lib/{api.ts,stores/auth.ts}    # API helper, auth state
  routes/{+page,login,call/[callId]/{+page,setup}}
```
