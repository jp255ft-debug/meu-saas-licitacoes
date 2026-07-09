package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/middleware"
)

// AuthHandler lida com requisições de autenticação e validação de API Keys.
type AuthHandler struct {
	providers []middleware.AuthProvider
}

// NewAuthHandler cria uma nova instância do handler de autenticação.
func NewAuthHandler(providers []middleware.AuthProvider) *AuthHandler {
	return &AuthHandler{providers: providers}
}

// validateAPIKeyResponse representa a resposta do endpoint de validação.
type validateAPIKeyResponse struct {
	Valid    bool   `json:"valid"`
	ClientID string `json:"client_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ValidateAPIKey valida uma API Key e retorna se é válida.
func (h *AuthHandler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(ctx, "failed to decode validate request", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validateAPIKeyResponse{Valid: false, Error: "invalid request body"})
		return
	}

	if req.APIKey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validateAPIKeyResponse{Valid: false, Error: "apiKey is required"})
		return
	}

	// Tentar cada provider em ordem (MapAuthProvider, CachedDBProvider)
	for _, provider := range h.providers {
		if clientID, ok := provider.Validate(req.APIKey); ok {
			slog.InfoContext(ctx, "api key validated successfully", "client_id", clientID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(validateAPIKeyResponse{Valid: true, ClientID: clientID})
			return
		}
	}

	slog.WarnContext(ctx, "api key validation failed")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(validateAPIKeyResponse{Valid: false, Error: "invalid api key"})
}