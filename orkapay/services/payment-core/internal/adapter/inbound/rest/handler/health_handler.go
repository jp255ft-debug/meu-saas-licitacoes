package handler

import (
	"encoding/json"
	"net/http"
)

// HealthHandler expõe endpoints de health check.
type HealthHandler struct{}

// NewHealthHandler cria um novo handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health godoc
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Ready godoc
// GET /ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// Para o MVP, sempre retorna ready
	// Em produção, verificar conexões com DB, Kafka, etc.
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}
