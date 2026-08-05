package server

import (
	"net/http"

	handlers "github.com/sahil/leasewebassignment/internal/server/handlers"
	"github.com/sahil/leasewebassignment/internal/server/middleware"
	"github.com/sahil/leasewebassignment/internal/service"
)

type Server struct {
	handler        *handlers.Handler
	authMiddleware *middleware.AuthMiddleware
	once           *http.ServeMux
}

func NewServer(s service.Service, jwtSigningKey string) *Server {
	return &Server{
		handler:        handlers.NewHandler(s),
		authMiddleware: middleware.NewAuthMiddleware(jwtSigningKey),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.routes().ServeHTTP(w, r)
}
