package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Trace middleware starts a span for every HTTP request and propagates context
func Trace(next http.Handler) http.Handler {
	tracer := otel.Tracer("") // use service global tracer
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract any existing trace context from incoming headers
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Start a new span for this HTTP request
		spanName := r.Method + " " + r.URL.Path
		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// Add standard HTTP attributes
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.Path),
			attribute.String("http.client_ip", r.RemoteAddr),
			attribute.String("http.user_agent", r.UserAgent()),
		)

		// Inject updated trace context into response headers
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))
		w.Header().Set("x-request-id", span.SpanContext().TraceID().String())

		// Pass context to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
