package hooks

import (
	"context"
	"fmt"

	"github.com/malakagl/kart-challenge/pkg/constants"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// TraceHook injects traceID from context into every log event
type TraceHook struct {
	ctx context.Context
}

func NewTraceHook(ctx context.Context) *TraceHook {
	return &TraceHook{ctx: ctx}
}

func (h TraceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	if h.ctx != nil {
		span := trace.SpanFromContext(h.ctx)
		if spanCtx := span.SpanContext(); spanCtx.IsValid() {
			e.Str("traceId", spanCtx.TraceID().String())
			e.Str("spanId", spanCtx.SpanID().String())

			if parentSpanID := h.ctx.Value(constants.ParentSpanId); parentSpanID != nil {
				e.Str("parentSpanId", fmt.Sprintf("%v", parentSpanID))
			}
		}
	}
}
