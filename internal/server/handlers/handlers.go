package handlers

import "github.com/sahil/leasewebassignment/internal/service"

type Handler struct {
	Service service.Service
}

func NewHandler(s service.Service) *Handler {
	return &Handler{Service: s}
}
