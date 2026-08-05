package server

import (
	"net/http"

	"github.com/sahil/leasewebassignment/internal/server/handlers"
)

func (s *Server) routes() *http.ServeMux {
	if s.once != nil {
		return s.once
	}

	mux := http.NewServeMux()
	mux.Handle("/servers", handlers.Adapt(s.handler.GetServers))
	mux.Handle("/admin/upload", s.authMiddleware.Middleware(handlers.Adapt(s.handler.Upload)))

	s.once = mux
	return mux
}
