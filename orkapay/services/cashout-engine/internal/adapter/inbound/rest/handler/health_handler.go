package handler

import (
	"net/http"
)

// HealthHandler expõe endpoints de health check.
type HealthHandler struct{}

// NewHealthHandler cria uma nova instância do health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health godoc
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// Ready godoc
// GET /ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}