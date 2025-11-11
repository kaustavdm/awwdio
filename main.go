package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kaustavdm/awwdio/config"
	"github.com/kaustavdm/awwdio/internal/api"
)

//go:embed web/build/*
var buildFS embed.FS

//go:embed web/static/*
var staticFS embed.FS

// init is run before main(), so it is the ideal place to set up logging
// Note that we do not need to call init() in main() as it is automatically called by Go
func init() {
	// 0. Set up logging
	// Check if JSON_LOGGER and DEBUG environment variables are set
	_, jsonLogger := os.LookupEnv("JSON_LOGGER")
	_, debug := os.LookupEnv("DEBUG")

	// Default to log level Info
	var level slog.Level

	// If DEBUG is set, set the log level to Debug
	if debug {
		level = slog.LevelDebug
	}

	// If JSON_LOGGER is set, use JSON logging
	if jsonLogger {
		jsonh := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
		slog.SetDefault(slog.New(jsonh))
	} else {
		// Otherwise, use text logging
		texth := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
		slog.SetDefault(slog.New(texth))
	}

	slog.Info("Logger initialized", slog.Bool("json", jsonLogger), slog.Bool("debug", debug))
}

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", slog.String("error", err.Error()))
		return
	}

	// 2. Create HTTP mux (router)
	// This will act as the main router for the web server
	mux := http.NewServeMux()

	// 2.a. Set up API routes
	apiServer := api.New(cfg)
	apiMux := http.NewServeMux()
	apiServer.Register(apiMux)
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// 3. Serve static files from the "web/build" directory (SvelteKit build output)
	// 3.a. Create sub filesystem for build files
	buildFs, err := fs.Sub(buildFS, "web/build")
	if err != nil {
		slog.Error("Failed to create sub filesystem for build files", slog.String("directory", "web/build"), slog.String("error", err.Error()))
		return
	}
	// 3.b. Create file server handler for build assets
	buildFileServer := http.FileServer(http.FS(buildFs))
	mux.Handle("GET /static/", http.StripPrefix("/static/", buildFileServer))

	// 3.b.1. Serve SvelteKit's _app directory (contains JS, CSS, and other assets)
	mux.Handle("GET /_app/", buildFileServer)

	// 3.c. Serve specific files from web/static/ at root level
	// These are served directly at root (e.g., /favicon.ico, /robots.txt)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile("web/static/favicon.ico")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(data)
	})

	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile("web/static/robots.txt")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(data)
	})

	// 3.e. Serve SPA (Single Page Application) - catch-all route for client-side routing
	// This must be last so it doesn't override other routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for all routes not matched above
		// This enables client-side routing in SvelteKit
		data, err := buildFS.ReadFile("web/build/index.html")
		if err != nil {
			http.Error(w, "Frontend not found", http.StatusNotFound)
			slog.Error("Failed to read index.html", slog.String("error", err.Error()))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// 4. Set up the server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("Server starting", slog.String("port", cfg.Port))

	// 5. Start the server
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed to start", slog.Any("error", err))
	}
}
