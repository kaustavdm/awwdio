# Awwdio

A lightweight fullstack audio (and video conversation) application built with Go and SvelteKit, leveraging Twilio's Programmable Video, Twilio Programmable Audio, and Twilio Verify.

## Architecture

**Backend (Go)**:
- Standard library-first approach with minimal external dependencies
- HTTP server with nested mux routing
- Embedded filesystem for serving frontend assets
- Single binary deployment

**Frontend (SvelteKit)**:
- TypeScript + Svelte 5
- TailwindCSS with Twilio Paste colors
- Twilio Video JS SDK integration
- Static build embedded into Go binary
- Build output: `web/build/` (embedded in binary)
- Static assets: `web/static/` (served at root, embedded in binary)

## Prerequisites

- **Go** 1.22+ ([installation guide](https://go.dev/doc/install))
- **Node.js** 18+ and npm ([installation guide](https://nodejs.org/))
- **Twilio Account** with:
  - Account SID
  - API Key and Secret ([create one](https://www.twilio.com/console/project/api-keys))

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/kaustavdm/awwdio.git
cd awwdio
```

### 2. Configure Environment Variables

Copy the sample environment file and add your Twilio credentials:

```bash
cp sample.env .env
# Edit .env with your actual credentials
source .env
```

**Required Environment Variables:**

- `TWILIO_ACCOUNT_SID`: Your Twilio account SID (find in [Twilio Console](https://www.twilio.com/console))
- `TWILIO_API_KEY`: API key SID (create at [API Keys](https://www.twilio.com/console/project/api-keys))
- `TWILIO_API_SECRET`: API key secret
- `PORT`: Server port (default: `8080`)

**Optional Environment Variables:**

- `DEBUG`: Set to `true` to enable debug logging
- `JSON_LOGGER`: Set to `true` for JSON-formatted logs

### 3. Install Frontend Dependencies

```bash
cd web
npm install
cd ..
```

## Development

### Quick Start with Build Script

The easiest way to work with the project is using the included `build.sh` script:

```bash
# Run backend in development mode
./build.sh dev

# In another terminal, run frontend dev server
cd web && npm run dev
```

### Running Backend Only (Development Mode)

Run the Go server directly from source:

```bash
# Using the build script (recommended)
./build.sh dev

# Or manually
go run main.go
```

The server will start on `http://localhost:8080` (or the port specified in `PORT`).

### Running Frontend Development Server

For frontend development with hot reload:

```bash
cd web
npm run dev
```

The frontend dev server runs on `http://localhost:5173` and proxies API requests to `http://localhost:8080`.

**Make sure the backend is running separately** when using the frontend dev server.

### Full Development Workflow

#### Using Build Script (Recommended)

1. **Terminal 1** - Start backend:
   ```bash
   ./build.sh dev
   ```

2. **Terminal 2** - Start frontend dev server:
   ```bash
   cd web
   npm run dev
   ```

3. Open `http://localhost:5173` for hot-reloading frontend development

#### Manual Approach

1. **Terminal 1** - Start backend:
   ```bash
   source .env
   go run main.go
   ```

2. **Terminal 2** - Start frontend dev server:
   ```bash
   cd web
   npm run dev
   ```

3. Open `http://localhost:5173` for hot-reloading frontend development

## Building

### Quick Start with Build Script

The easiest way to build the project:

```bash
# Full production build
./build.sh build

# Or build components separately
./build.sh build-ui     # Frontend only
./build.sh build-api    # Backend only

# Clean build artifacts
./build.sh cleanup

# Clean rebuild
./build.sh cleanup && ./build.sh build
```

### Production Build (Manual)

Build a single executable with embedded frontend:

```bash
# 1. Build frontend
cd web && npm run build && cd ..

# 2. Build Go binary with embedded frontend
go build -o bin/awwdio

# 3. Run the binary
./bin/awwdio
```

The binary embeds all frontend assets and serves everything from a single process.

### Build Steps Breakdown

#### Frontend Only

```bash
# Using build script
./build.sh build-ui

# Or manually
cd web
npm run build
```

Output: `web/build/` (embedded by Go binary)

#### Backend Only

```bash
# Using build script
./build.sh build-api

# Or manually
go build -o bin/awwdio
```

Output: `bin/awwdio` (10MB binary with embedded frontend)

## Running

### Development Mode

```bash
source .env
go run main.go
```

### Production Mode

```bash
./bin/awwdio
```

The application serves both frontend and API from the same port (default: 8080):

- Frontend: `http://localhost:8080/`
- API: `http://localhost:8080/api/*`

## Development Notes

- **Hot Reload**: Use `npm run dev` in the `web/` directory for frontend hot reload
- **API Changes**: Restart `go run main.go` when modifying backend code
- **Full Rebuild**: Run `cd web && npm run build && cd .. && go build -o bin/awwdio` after frontend changes for production
- **Design System**: Uses [Twilio Paste](https://paste.twilio.design/) color palette

## Troubleshooting

**Port already in use:**
```bash
lsof -ti:8080 | xargs kill -9
```

**Frontend not updating:**
```bash
# Using build script (recommended)
./build.sh cleanup && ./build.sh build

# Or manually
find web/build -mindepth 1 ! -name '.gitkeep' -delete  # Clean build output
rm -rf web/.svelte-kit                                  # Clean SvelteKit cache
cd web && npm run build && cd ..
go build -o bin/awwdio
```

**API endpoints not working:**
- Ensure environment variables are set (`source .env`)
- Check Twilio credentials are valid
- Verify `TWILIO_API_KEY` matches `TWILIO_ACCOUNT_SID`

## License

[MIT](LICENSE).
