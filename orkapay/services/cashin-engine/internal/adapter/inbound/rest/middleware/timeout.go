package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout retorna um middleware que adiciona timeout ao contexto da requisição.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			// Channel para detectar quando a requisição termina
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Requisição completou dentro do prazo
			case <-ctx.Done():
				// Timeout ou cancelamento
				if ctx.Err() == context.DeadlineExceeded {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusGatewayTimeout)
					w.Write([]byte(`{"error":{"code":"timeout","message":"requisição excedeu o tempo limite"}}`))
				}
			}
		})
	}
}