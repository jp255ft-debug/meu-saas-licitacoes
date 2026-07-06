package middleware

import (
	"net/http"
	"time"
)

// ---------------------------------------------------------------------------
// Tracing middleware - Instrumentação para OpenTelemetry
//
// Este middleware adiciona tracing distribuído às requisições HTTP.
// Para ativar, adicione as seguintes dependências ao go.mod:
//
//   go.opentelemetry.io/otel
//   go.opentelemetry.io/otel/attribute
//   go.opentelemetry.io/otel/trace
//   go.opentelemetry.io/otel/propagation
//   go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
//   go.opentelemetry.io/otel/sdk/trace
//
// E descomente o código abaixo.
// ---------------------------------------------------------------------------

// var tracer = otel.Tracer("orkapay/cashout")

// func TracingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
//
// 		// Extrai contexto de propagação (headers de tracing)
// 		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
//
// 		// Inicia span para esta requisição
// 		ctx, span := tracer.Start(ctx, r.URL.Path,
// 			trace.WithAttributes(
// 				attribute.String("http.method", r.Method),
// 				attribute.String("http.url", r.URL.String()),
// 				attribute.String("http.host", r.Host),
// 				attribute.String("http.user_agent", r.UserAgent()),
// 			))
// 		defer span.End()
//
// 		// Wrap response writer para capturar status code
// 		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
//
// 		start := time.Now()
// 		next.ServeHTTP(wrapped, r.WithContext(ctx))
// 		duration := time.Since(start)
//
// 		// Adiciona atributos da resposta ao span
// 		span.SetAttributes(
// 			attribute.Int("http.status_code", wrapped.statusCode),
// 			attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
// 		)
//
// 		// Marca como erro se status >= 500
// 		if wrapped.statusCode >= 500 {
// 			span.SetAttributes(attribute.Bool("error", true))
// 		}
// 	})
// }

// responseWriter wrapper para capturar o status code HTTP.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ---------------------------------------------------------------------------
// Função auxiliar para inicializar o tracer (chamar no main.go)
//
// Para usar, descomente e adicione as dependências OpenTelemetry.
//
// func InitTracer(ctx context.Context, serviceName string) func() {
// 	exporter, err := otlptracehttp.New(ctx,
// 		otlptracehttp.WithInsecure(),
// 		otlptracehttp.WithEndpoint("localhost:4318"),
// 	)
// 	if err != nil {
// 		log.Printf("failed to create exporter: %v", err)
// 		return func() {}
// 	}
//
// 	tp := sdktrace.NewTracerProvider(
// 		sdktrace.WithBatcher(exporter),
// 		sdktrace.WithResource(resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceName(serviceName),
// 		)),
// 	)
//
// 	otel.SetTracerProvider(tp)
// 	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
// 		propagation.TraceContext{},
// 		propagation.Baggage{},
// 	))
//
// 	return func() {
// 		if err := tp.Shutdown(ctx); err != nil {
// 			log.Printf("failed to shutdown tracer: %v", err)
// 		}
// 	}
// }
// ---------------------------------------------------------------------------

// TracingMiddleware é um middleware não-operacional que serve como placeholder
// para quando o tracing não está ativo. Substitua pela implementação real
// quando as dependências OpenTelemetry estiverem disponíveis.
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		_ = duration
	})
}