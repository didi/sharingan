package tlog

import (
	"bytes"
	"context"
)

type Tracer interface {
	Format() string
}

var NewTracer = func(data map[string]string) *tracer {
	return &tracer{data: data}
}

type tracer struct {
	data map[string]string
}

func (t *tracer) Format() string {
	var buf bytes.Buffer
	for k, v := range t.data {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(v)
		buf.WriteString("||")
	}
	return buf.String()
}

const tracerKey = "::tracer::"

func TraceContext(ctx context.Context, t Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, t)
}
