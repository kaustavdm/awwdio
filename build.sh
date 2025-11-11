#!/bin/bash

# Awwdio Build Script
# Provides convenience commands for building and running the application

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Helper function to print colored messages
print_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to build frontend
build_ui() {
    print_info "Building frontend..."
    cd web
    npm run build
    cd ..
    print_success "Frontend built successfully → web/build/"
}

# Function to build backend
build_api() {
    print_info "Building Go binary..."
    go build -o bin/awwdio
    print_success "Backend built successfully → bin/awwdio ($(du -h bin/awwdio | cut -f1))"
}

# Function to do full build
build() {
    print_info "Starting full build..."
    build_ui
    build_api
    print_success "Full build complete! Run with: ./bin/awwdio"
}

# Function to clean build artifacts
cleanup() {
    print_info "Cleaning build artifacts..."

    # Clean frontend build (keep .gitkeep)
    if [ -d "web/build" ]; then
        # Remove all files except .gitkeep
        find web/build -mindepth 1 ! -name '.gitkeep' -delete
        print_success "Cleaned web/build/ (preserved .gitkeep)"
    fi

    if [ -d "web/.svelte-kit" ]; then
        rm -rf web/.svelte-kit
        print_success "Removed web/.svelte-kit/"
    fi

    # Clean backend build
    if [ -f "bin/awwdio" ]; then
        rm -f bin/awwdio
        print_success "Removed bin/awwdio"
    fi

    # Clean node_modules (optional, commented out)
    # if [ -d "web/node_modules" ]; then
    #     rm -rf web/node_modules
    #     print_success "Removed web/node_modules/"
    # fi

    print_success "Cleanup complete!"
}

# Function to run in development mode
dev() {
    print_info "Starting development mode..."
    print_warning "Make sure your .env file is configured!"

    # Check if .env exists
    if [ ! -f ".env" ]; then
        print_error ".env file not found!"
        echo ""
        echo "Create .env from sample:"
        echo "  cp sample.env .env"
        echo "  # Edit .env with your Twilio credentials"
        echo "  source .env"
        exit 1
    fi

    # Source .env
    set -a  # automatically export all variables
    source .env
    set +a

    print_success "Environment loaded"
    print_info "Starting backend server..."
    print_warning "Press Ctrl+C to stop"
    echo ""
    echo "Backend:  http://localhost:${PORT:-8080}"
    echo "Frontend: Run 'cd web && npm run dev' in another terminal for hot reload"
    echo ""

    go run main.go
}

# Function to show help
show_help() {
    cat << EOF
Awwdio Build Script

Usage: ./build.sh [command]

Commands:
  build       Build both frontend and backend (production)
  build-ui    Build frontend only
  build-api   Build backend only
  cleanup     Clean all build artifacts
  dev         Run in development mode (backend only)
  help        Show this help message

Examples:
  # Full production build
  ./build.sh build

  # Build and run
  ./build.sh build && ./bin/awwdio

  # Development workflow (2 terminals)
  # Terminal 1:
  ./build.sh dev
  # Terminal 2:
  cd web && npm run dev

  # Clean rebuild
  ./build.sh cleanup && ./build.sh build

EOF
}

# Main script logic
case "${1:-}" in
    build)
        build
        ;;
    build-ui)
        build_ui
        ;;
    build-api)
        build_api
        ;;
    cleanup|clean)
        cleanup
        ;;
    dev)
        dev
        ;;
    help|--help|-h)
        show_help
        ;;
    "")
        print_error "No command specified"
        echo ""
        show_help
        exit 1
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
