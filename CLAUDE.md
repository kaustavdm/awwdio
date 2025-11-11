# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Awwdio is a lightweight fullstack audio and video conversation application built with Go and Svelte/SvelteKit, leveraging Twilio's Programmable Video API.

**Architecture:**
- **Backend**: Go server handling authentication, room management, and API endpoints
- **Frontend**: Svelte/SvelteKit application (to be implemented in `web/` directory)
- **Deployment**: Single compiled Go binary with embedded frontend resources

The backend uses Go's standard library wherever possible, with the Twilio SDK being the primary external dependency for video functionality.

## Common Commands

### Building and Running

**Production Build** (creates single binary with embedded frontend):
```bash
# 1. Build the frontend
cd web && npm run build && cd ..

# 2. Build the Go binary (embeds frontend resources from web/build/)
go build -o bin/awwdio
```

**Run the built binary:**
```bash
./bin/awwdio
```

**Development Mode** (run backend directly from source):
```bash
go run main.go
```

**Note**: The Go binary embeds:
- Frontend build output from `web/build/` (SvelteKit app files) using `//go:embed web/build/*`
- Static assets from `web/static/` (files served at root like favicon.ico) using `//go:embed web/static/*`
All files in `web/static/` are served at the root URL path (e.g., `web/static/favicon.ico` → `http://localhost:8080/favicon.ico`).

### Environment Setup

The application requires the following environment variables:

- `TWILIO_ACCOUNT_SID`: Your Twilio account SID (required)
- `TWILIO_API_KEY`: Twilio API key SID (required)
- `TWILIO_API_SECRET`: Twilio API secret (required)
- `PORT`: Server port (optional, defaults to 8080)

Optional configuration:
- `DEBUG`: Set to `true` to enable debug logging
- `JSON_LOGGER`: Set to `true` to use JSON-formatted logs

Quick setup:
```bash
cp sample.env .env
# Edit .env with your actual credentials
source .env
```

## Architecture

### Fullstack Structure

**Backend (Go)**:
- Standard library-first approach (minimal external dependencies)
- HTTP server with nested mux routing
- Embedded filesystem for serving compiled frontend assets

**Frontend (Svelte/SvelteKit)**:
- Lives in `web/` directory
- Build output goes to `web/build/` for embedding (gitignored)
- Static assets in `web/static/` are served at root level (e.g., favicon.ico, robots.txt)
- Communicates with backend via `/api/*` endpoints
- Uses TailwindCSS with dark/light mode support
- Integrates Twilio Video JS SDK for WebRTC

**Build & Deployment**:
- Frontend compiled to static assets
- Assets embedded into Go binary using `embed.FS`
- Single executable contains both frontend and backend

### Core Backend Components

1. **Main Application** (`main.go`):
   - Sets up HTTP server with routing
   - Configures logging (text or JSON format, with debug mode)
   - Embeds and serves frontend static files
   - Route structure:
     - `/api/*` → API endpoints
     - `/static/*` → Static assets
     - `/favicon.ico`, `/robots.txt` → Root-level resources

2. **Configuration** (`config/config.go`):
   - Loads environment variables into configuration struct
   - Provides validation for required settings
   - No external config file dependencies

3. **API Layer** (`internal/api/`):
   - Modular API structure with route registration
   - Video API module with Twilio integration
   - Uses nested mux instances for clean route organization

4. **Video Module** (`internal/api/video/`):
   - Handles Twilio access token generation (`access_token.go`)
   - Manages room creation and access (`video.go`, `room.go`)
   - Integrates with Twilio REST client

### Request Flow

1. Client makes request to server
2. Request is routed to appropriate handler in API module
3. Handler processes request, potentially communicating with Twilio
4. Response is returned to client with token or room information

## Development Principles

### Dependency Management
**Always prefer Go's standard library over third-party dependencies.** The only exception is the Twilio SDK, which is necessary for video functionality. This principle ensures:
- Minimal external dependencies
- Reduced security surface area
- Better long-term maintainability
- Smaller binary size

### Backend Development

- When adding new API endpoints, follow the pattern in `internal/api/video/video.go` by adding handler methods and registering them in the `Register` function
- New API modules should be registered in `internal/api/api.go`
- The codebase uses Go's standard library for HTTP handling with nested mux instances for route organization
- Error handling follows Go's idiomatic approach with early returns and clear error messages
- Use `slog` (standard library) for all logging needs

### Frontend Development

**Structure:**
- Frontend source code lives in `web/` directory
- SvelteKit build outputs to `web/build/` directory (configured in `svelte.config.js`, gitignored)
- Static assets in `web/static/` are served at root level in the final binary
- Frontend communicates with backend exclusively through `/api/*` endpoints
- During development, run SvelteKit dev server separately and proxy API requests to the Go backend (configured in `vite.config.ts`)

**Implemented Pages:**
1. **Homepage** (`/`): Landing page with "Start Call" CTA button
   - Redirects to login if user not authenticated
   - Shows user info when logged in
2. **Login** (`/login`): Multi-step authentication flow
   - Email input → OTP verification → Display name (optional)
   - Uses Twilio Verify (API endpoints to be implemented)
3. **Call Setup** (`/call/[callId]/setup`): Pre-call device configuration
   - Join via web or phone options
   - Audio/video device selection
   - Live audio meter and video preview
4. **Call** (`/call/[callId]`): Active call interface
   - Participant grid with local and remote participants
   - Audio/video toggle controls
   - Twilio Video SDK integration
   - Invite link sharing

**Key Components:**
- `web/src/lib/stores/auth.ts`: User authentication state management
- `web/src/app.css`: TailwindCSS configuration with dark mode support
- `web/src/routes/+layout.svelte`: Global layout with theme toggle

**Development Workflow:**
```bash
# Install dependencies
cd web && npm install

# Run dev server (with API proxy to localhost:8080)
npm run dev

# Build for production
npm run build

# Output will be in web/build/ ready for Go embedding
```

**Before production builds:**
- Ensure frontend build completes successfully before compiling the Go binary
- The build process uses `@sveltejs/adapter-static` to generate static files
- All routes use client-side routing with fallback to `index.html`

## Claude Code Guidance

This section provides specific instructions for Claude Code when working with this repository.

### Output Token Limit Workaround

Claude Code has a per-response output token limit controlled by `CLAUDE_CODE_MAX_OUTPUT_TOKENS`. If you encounter the error:

```
API Error: Claude's response exceeded the 4096 output token maximum
```

When writing large files or making extensive changes, break the work into smaller chunks:

1. **Use multiple tool calls** instead of large single responses
2. **Write files in sections**: Create the file structure first, then fill in implementations across multiple responses
3. **Edit incrementally**: Make focused edits to specific sections rather than rewriting entire files
4. **Summarize between chunks**: After each chunk, provide a brief status update before continuing

**Example Chunking Strategy:**
```
Response 1: Create file skeleton and imports
Response 2: Implement first major function/section
Response 3: Implement second major function/section
Response 4: Add remaining functions and exports
```

This approach ensures all work is completed even with output token constraints.

### Pending Backend Implementation

The following API endpoints are referenced by the frontend but not yet implemented:

1. **Authentication Endpoints** (Twilio Verify integration needed):
   - `POST /api/auth/send-otp` - Send OTP to email via Twilio Verify
     - Request: `{ "email": "user@example.com" }`
     - Response: `{ "success": true }`
   - `POST /api/auth/verify-otp` - Verify OTP code
     - Request: `{ "email": "user@example.com", "otp": "123456" }`
     - Response: `{ "success": true, "token": "session_token" }`

2. **Room Management**:
   - Create new rooms/calls (currently using client-side UUID generation)
   - List active rooms
   - Room metadata and participant tracking

3. **Phone Integration** (Future):
   - `POST /api/call/phone` - Initiate phone call to join room
     - Twilio Voice API integration
     - SIP/PSTN bridge to video room

**Implementation Notes:**
- Follow the existing pattern in `internal/api/video/` for new modules
- Create an `internal/api/auth/` module for authentication endpoints
- Use Go's standard library for HTTP handling
- Add proper error handling and validation
- Consider session management for authenticated users
- Do not add emojis
- Add code comments for complex logic