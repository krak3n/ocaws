package ocawstest

import (
	"go.opencensus.io/trace"
)

// Consistent IDs for testing
var (
	DefaultTraceID = trace.TraceID([16]byte{97, 98, 99, 100, 101, 102, 103, 104, 105, 103, 107, 108, 109, 110, 111, 113})
	DefaultSpanID  = trace.SpanID([8]byte{97, 98, 99, 100, 101, 102, 103, 104})
)

// NewTestIDGenerator constructs a new trace ID generator for testing so we can
// assert on consistent trace ids
func NewTestIDGenerator() *TestIDGenerator {
	return &TestIDGenerator{
		TraceID: DefaultTraceID,
		SpanID:  DefaultSpanID,
	}
}

// TestIDGenerator implements the trace.IDGenerator interface
type TestIDGenerator struct {
	TraceID trace.TraceID
	SpanID  trace.SpanID
}

// NewTraceID returns the trace id
func (t *TestIDGenerator) NewTraceID() [16]byte {
	return t.TraceID
}

// NewSpanID returns the span id
func (t *TestIDGenerator) NewSpanID() [8]byte {
	return t.SpanID
}
