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

//go:embed web/static/*
var static embed.FS

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

	// 3. Serve static files from the "web/static" directory
	// 3.a. Create sub filesystem for static files
	staticFs, err := fs.Sub(static, "web/static")
	if err != nil {
		slog.Error("Failed to create sub filesystem for static files", slog.String("directory", "web/static"), slog.String("error", err.Error()))
		return
	}
	// 3.b. Create file server handler
	fileServer := http.FileServer(http.FS(staticFs))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// 3.c. Serve favicon and robots.txt from root
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := static.ReadFile("web/static/favicon.ico")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(data)
	})

	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		data, err := static.ReadFile("web/static/robots.txt")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
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
