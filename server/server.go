package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kaustavdm/awwdio/config"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	config *config.Config
	router *http.ServeMux
	logger *slog.Logger
}

// NewServer creates a new Server instance with the given configuration and logger.
func NewServer(cfg *config.Config, logger *slog.Logger) *Server {
	s := &Server{
		config: cfg,
		router: http.NewServeMux(),
		logger: logger,
	}
	s.routes()
	return s
}

// routes sets up the HTTP routes for the server.
func (s *Server) routes() {
	s.registerIndexRoutes()
	s.registerRoomRoutes()
	s.registerTokenRoutes()
	// Add calls to other route registration functions here
}

// Start starts the HTTP server on the configured port.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.logger.Info("Starting server", "address", addr)
	return http.ListenAndServe(addr, s.router)
}
