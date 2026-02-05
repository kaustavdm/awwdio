# LEARNINGS.md

Technical implementation details and patterns discovered in the Awwdio codebase. Update this document as you discover new patterns or implement features.

## Code Patterns & Conventions

### Backend Patterns

**Handler Pattern** (`internal/api/video/video.go:18-30`):
```go
type Handler struct {
    config *config.Config
    client *twilio.RestClient
}

func NewHandler(cfg *config.Config) *Handler {
    return &Handler{
        config: cfg,
        client: twilio.NewRestClient(cfg.TwilioApiKey, cfg.TwilioApiSecret, cfg.TwilioAccountSID),
    }
}

func (h *Handler) Register(mux *http.ServeMux) {
    mux.HandleFunc("POST /token", h.GenerateToken)
    mux.HandleFunc("GET /room", h.GetRoom)
}
```

**Three-Tier Nested Mux Routing** (`main.go:53-62`, `internal/api/api.go:15-24`):
```go
// Main → API → Module (Auth/Video)
mainMux := http.NewServeMux()
apiMux := http.NewServeMux()
videoMux := http.NewServeMux()

videoHandler.Register(videoMux)
apiMux.Handle("/video/", http.StripPrefix("/video", videoMux))
mainMux.Handle("/api/", http.StripPrefix("/api", apiMux))
```

**Logging Setup** (`main.go:16-31`):
```go
func init() {
    level := slog.LevelInfo
    if os.Getenv("DEBUG") == "true" {
        level = slog.LevelDebug
    }

    if os.Getenv("JSON_LOGGER") == "true" {
        slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &opts)))
    } else {
        slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &opts)))
    }
}
```

**Embedded Frontend** (`main.go:34-37`):
```go
//go:embed web/build/*
var buildFS embed.FS

//go:embed web/static/*
var staticFS embed.FS
```

### Frontend Patterns

**Svelte 5 Runes Pattern** (`web/src/lib/stores/auth.ts:3-28`):
```javascript
class AuthStore {
    user = $state<User | null>(null);
    token = $state<string | null>(null);

    constructor() {
        // Load from localStorage on init
        const stored = localStorage.getItem('auth');
        if (stored) {
            const parsed = JSON.parse(stored);
            this.user = parsed.user;
            this.token = parsed.token;
        }
    }
}
```

## API Specifications

### Implemented Endpoints

**POST /api/video/token** (`internal/api/video/access_token.go`)
- Generates Twilio JWT for Video SDK
- Requires: `Authorization: Bearer <jwt_token>` header
- Request: `{ "room": "room_name" }`
- Response: `{ "token": "twilio_jwt_token_string" }`
- User identity extracted from authenticated JWT

**GET /api/video/room** (`internal/api/video/room.go:16-64`)
- Fetches room details from Twilio
- Query param: `?name=room_name`
- Response: `{ "sid": "string", "status": "string", "duration": number, ... }`

**POST /api/auth/send-otp** (`internal/api/auth/auth.go:68-107`)
- Sends OTP via Twilio Verify
- Request: `{ "email": "user@example.com" }`
- Response: `{ "success": true }`
- Requires: `TWILIO_VERIFY_SERVICE_SID` env var

**POST /api/auth/verify-otp** (`internal/api/auth/auth.go`)
- Verifies OTP code and returns JWT
- Request: `{ "channel": "email", "to": "user@example.com", "otp": "123456" }`
- Response: `{ "success": true, "token": "jwt_token" }`
- JWT contains: sub (user email), iat, exp (24h expiry)
- Signed with HS256 using JWT_SECRET

### Pending API Features

1. **Room Management**:
   - POST /api/video/room - Create new room
   - GET /api/video/rooms - List active rooms
   - DELETE /api/video/room/:sid - End room
   - GET /api/video/room/:sid/participants - Track participants

2. **Session Management**:
   - Proper JWT/session tokens (currently placeholder)
   - Token refresh mechanism
   - Session storage strategy

3. **Phone Integration** (Future):
   - POST /api/call/phone - Twilio Voice API bridge
   - SIP/PSTN to video room bridge

## Build & Deployment

### Build Script (`build.sh`)

**Functions**:
- `build_ui()`: Runs `npm run build` in web/
- `build_api()`: Runs `go build -o bin/awwdio`
- `cleanup()`: Removes bin/ and web/build/

**Full Build Process**:
```bash
./build.sh  # Builds both frontend and backend
```

### Production Build Configuration

**SvelteKit** (`web/svelte.config.js:4-10`):
```javascript
adapter: adapter({
    fallback: 'index.html',  // SPA fallback
    pages: 'build',
    assets: 'build',
    precompress: false,
    strict: true
})
```

**Vite Proxy** (`web/vite.config.ts:8-14`):
```javascript
proxy: {
    '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false
    }
}
```

## Architecture Decisions & Rationale

### Single Binary Deployment
- **Why**: Simplifies deployment, no CORS issues, atomic updates
- **How**: embed.FS packages frontend into Go binary
- **Trade-off**: Larger binary size (~15-20MB with frontend)

### Standard Library First
- **Why**: Stability, security, minimal dependencies
- **Implementation**: Only external dep is Twilio SDK
- **Benefits**: Reduced supply chain risk, easier auditing

### Stateless Backend
- **Why**: Horizontal scaling, simplicity
- **Implementation**: No server-side sessions (yet)
- **Challenge**: Need proper session management

### Three-Tier Mux Routing
- **Why**: Clean module separation, easy to add new modules
- **Pattern**: Main → API → Module-specific handlers
- **Benefit**: Each module owns its subroutes

## Security Considerations

### Current Implementation
- Input validation in config (`config/config.go:26-31`)
- HTTPS required for WebRTC (`getUserMedia` won't work on HTTP)
- Environment-based secrets (never committed)

### Missing Security Features
- CORS headers not configured
- No rate limiting
- No request size limits
- No CSRF protection

### Implemented Security Features
- JWT authentication with HS256 signing (`internal/api/auth/jwt.go`)
- Auth middleware for protected routes (`internal/api/middleware/auth.go`)
- Token expiration (24 hours)
- Required JWT_SECRET environment variable

### Recommendations
1. Add rate limiting middleware
2. Configure CORS for production domain
3. Add request size limits
4. Implement CSRF tokens for state-changing operations

## Known Issues & TODOs

### Critical
- None currently

### Important
- No room creation API
- No participant tracking
- Missing error recovery in build script
- No graceful shutdown handling

### Nice to Have
- Structured error responses
- Request ID tracking
- Metrics/monitoring endpoints
- Health check endpoint

## Troubleshooting

### Common Issues

**Frontend not updating**:
- Ensure `npm run build` completes in web/
- Check web/build/ exists before Go build
- Clear browser cache

**OTP not sending**:
- Verify `TWILIO_VERIFY_SERVICE_SID` is set
- Check Twilio Verify service is configured for email
- Ensure email format validation

**Video token invalid**:
- Check all Twilio env vars are set
- Verify API key has Video grants enabled
- Room name must be valid (no special chars)

**Port already in use**:
```bash
lsof -i :8080  # Find process
kill -9 <PID>  # Kill process
```

### Debug Mode
```bash
DEBUG=true JSON_LOGGER=true go run main.go
```
Shows detailed logs in JSON format for parsing.

## Development Tips

### Adding New API Module
1. Create `internal/api/newmodule/` directory
2. Implement Handler struct with NewHandler() and Register()
3. Register in `internal/api/api.go`
4. Follow error handling pattern: validate → process → respond

### Frontend API Integration
```javascript
// Use relative URLs, Vite proxy handles in dev
const response = await fetch('/api/endpoint', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
});
```

### Testing Twilio Integration
1. Use Twilio test credentials for unit tests
2. Create separate test room names
3. Clean up rooms after tests
4. Mock Twilio client for fast tests

## Configuration Reference

### Required Environment Variables
- `TWILIO_ACCOUNT_SID`: Main account identifier
- `TWILIO_API_KEY`: API key SID (starts with SK)
- `TWILIO_API_SECRET`: API key secret
- `TWILIO_VERIFY_SERVICE_SID`: Verify service for OTP
- `JWT_SECRET`: Secret key for signing JWT tokens (min 32 chars recommended)

### Optional Environment Variables
- `PORT`: Server port (default: 8080, validated 1-65535)
- `DEBUG`: Enable debug logging (true/false)
- `JSON_LOGGER`: Use JSON log format (true/false)

### Server Timeouts (`main.go:75-77`)
- ReadTimeout: 5 seconds
- WriteTimeout: 10 seconds
- IdleTimeout: 120 seconds

## File Structure Reference

```
.
├── main.go                 # Entry point, server setup
├── config/
│   └── config.go          # Environment configuration
├── internal/
│   └── api/
│       ├── api.go         # API router setup
│       ├── auth/
│       │   └── auth.go    # OTP authentication
│       └── video/
│           ├── video.go   # Handler registration
│           ├── access_token.go  # JWT generation
│           └── room.go    # Room management
├── web/
│   ├── src/
│   │   ├── lib/
│   │   │   └── stores/
│   │   │       └── auth.ts  # Auth state management
│   │   └── routes/
│   │       ├── +page.svelte          # Homepage
│   │       ├── login/+page.svelte    # Login flow
│   │       └── call/
│   │           └── [callId]/
│   │               ├── +page.svelte       # Active call
│   │               └── setup/+page.svelte # Pre-call
│   ├── static/            # Root-level assets
│   └── build/            # Compiled frontend (gitignored)
└── bin/                  # Compiled binary (gitignored)
```