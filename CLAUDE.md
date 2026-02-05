# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**IMPORTANT**: Before starting any work, read LEARNINGS.md for detailed technical implementation patterns, API specifications, and discovered conventions. Update LEARNINGS.md with new findings as you work.

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

**Fullstack**: Single Go binary with embedded SvelteKit frontend. Frontend in `web/`, build output in `web/build/` (embedded), static assets in `web/static/` served at root.

**Backend Structure**:
- `main.go`: HTTP server with three-tier nested mux routing, slog logging, embed.FS for frontend
- `config/config.go`: Environment variable loading and validation
- `internal/api/`: Modular API with Handler struct + Register() pattern
- `internal/api/video/`: Twilio token generation and room management
- `internal/api/auth/`: OTP authentication via Twilio Verify

**Request Flow**: Client → Nested mux routing → Handler → Twilio API → JSON response

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

### Implementation Guidelines

- Follow Handler struct + Register() pattern from `internal/api/video/`
- Use standard library for HTTP handling, slog for logging
- Early return error handling with clear messages
- No emojis in code or comments
- See LEARNINGS.md for API specifications and pending features