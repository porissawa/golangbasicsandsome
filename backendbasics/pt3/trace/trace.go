package trace

import "github.com/google/uuid"

type Trace struct {
	// TraceID is unique across the lifecycle of a single 'event', regardless of how many requests it takes to complete.
	// Carried in the `X-Trace-ID` header.
	TraceID uuid.UUID
	// RequestID is unique to each request. Carried in the `X-Request-ID` header.
	RequestID uuid.UUID
}
