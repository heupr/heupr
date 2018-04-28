package ingestor

import (
	"context"
	"net/http"
	"time"
)

// Server holds assets necessary for listening to and processing GitHub events.
type Server struct {
	server   http.Server
	database dataAccess
}

// Start begins server listening.
func (s *Server) Start() {

}

// Stop provides graceful server shutdown.
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	i.server.Shutdown(ctx)
	// NOTE: Include successful log shutdown here
}
