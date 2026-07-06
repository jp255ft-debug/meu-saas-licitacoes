package middleware

import (
	"context"
	"net/http"
	"sync"
)

// AuthProvider define a interface para validar chaves de API.
type AuthProvider interface {
	// Validate retorna o clientID se a chave for válida, ou string vazia se inválida.
	Validate(apiKey string) (string, bool)
}

// MapAuthProvider implementa AuthProvider usando um map em memória.
type MapAuthProvider struct {
	keys map[string]string
}

func NewMapAuthProvider(keys map[string]string) *MapAuthProvider {
	return &MapAuthProvider{keys: keys}
}

func (p *MapAuthProvider) Validate(apiKey string) (string, bool) {
	clientID, ok := p.keys[apiKey]
	return clientID, ok
}

// CachedDBProvider implementa AuthProvider com cache em memória + consulta ao banco.
type CachedDBProvider struct {
	mu       sync.RWMutex
	cache    map[string]string // apiKey -> clientID
	fallback func(apiKey string) (string, bool)
}

func NewCachedDBProvider(fallbackFn func(apiKey string) (string, bool)) *CachedDBProvider {
	return &CachedDBProvider{
		cache:    make(map[string]string),
		fallback: fallbackFn,
	}
}

func (p *CachedDBProvider) Validate(apiKey string) (string, bool) {
	// Tentar cache primeiro
	p.mu.RLock()
	clientID, ok := p.cache[apiKey]
	p.mu.RUnlock()
	if ok {
		return clientID, true
	}

	// Fallback para banco
	if p.fallback != nil {
		clientID, ok = p.fallback(apiKey)
		if ok {
			p.mu.Lock()
			p.cache[apiKey] = clientID
			p.mu.Unlock()
			return clientID, true
		}
	}

	return "", false
}

// APIKeyAuth retorna um middleware que valida o header X-API-Key.
// Aceita múltiplos providers para fallback.
func APIKeyAuth(providers ...AuthProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Rotas públicas não exigem autenticação
			path := r.URL.Path
			if path == "/health" || path == "/ready" || path == "/metrics" || len(path) >= 9 && path[:9] == "/swagger/" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				respondAuthError(w, "missing_api_key", "header X-API-Key é obrigatório")
				return
			}

			// Tentar cada provider em ordem
			for _, provider := range providers {
				if clientID, ok := provider.Validate(key); ok {
					r = r.WithContext(contextWithClientID(r.Context(), clientID))
					next.ServeHTTP(w, r)
					return
				}
			}

			respondAuthError(w, "invalid_api_key", "chave de API inválida")
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
