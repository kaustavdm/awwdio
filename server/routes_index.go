package server // Changed package name

import (
	"fmt"
	"net/http"
)

// registerIndexRoutes registers handlers for index-related routes.
func (s *Server) registerIndexRoutes() { // Updated receiver type
	s.router.HandleFunc("/", s.handleIndex())
}

// handleIndex returns a handler function for the root path.
func (s *Server) handleIndex() http.HandlerFunc { // Updated receiver type
	return func(w http.ResponseWriter, r *http.Request) {
		// You might want to use templates or serve static files here later
		fmt.Fprintf(w, "Welcome to Awwdio!")
	}
}
