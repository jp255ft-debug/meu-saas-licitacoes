package middleware

import (
	"context"
	"net/http"
)

// APIKeyAuth retorna um middleware que valida o header X-API-Key.
func APIKeyAuth(validKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			r = r.WithContext(context.WithValue(r.Context(), clientIDKey, clientID))
			next.ServeHTTP(w, r)
		})
	}
}

func respondAuthError(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `"}}`))
}

type contextKey string

const clientIDKey contextKey = "client_id"

// GetClientID extrai o clientID do contexto.
func GetClientID(ctx context.Context) string {
	if v := ctx.Value(clientIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}