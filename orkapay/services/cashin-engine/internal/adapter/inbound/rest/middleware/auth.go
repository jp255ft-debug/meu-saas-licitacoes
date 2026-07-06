package middleware

import (
	"context"
	"net/http"
)

// APIKeyAuth retorna um middleware que valida o header X-API-Key.
// validKeys é um map de chave -> clientID.
func APIKeyAuth(validKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Health check endpoints não exigem autenticação
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				respondAuthError(w, "missing_api_key", "header X-API-Key é obrigatório")
				return
			}

			clientID, ok := validKeys[key]
			if !ok {
				respondAuthError(w, "invalid_api_key", "chave de API inválida")
				return
			}

			// Adicionar clientID ao contexto
			r = r.WithContext(contextWithClientID(r.Context(), clientID))
			next.ServeHTTP(w, r)
		})
	}
}

// respondAuthError envia uma resposta de erro de autenticação padronizada.
func respondAuthError(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `"}}`))
}

// contextKey é um tipo para chaves de contexto.
type contextKey string

const clientIDKey contextKey = "client_id"

// contextWithClientID adiciona o clientID ao contexto.
func contextWithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}

// GetClientID extrai o clientID do contexto.
func GetClientID(ctx context.Context) string {
	if v := ctx.Value(clientIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
