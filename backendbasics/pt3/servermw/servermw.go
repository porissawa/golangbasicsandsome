package servermw

import (
	"dbs_and_more/ctxutil"
	"dbs_and_more/trace"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// Trace returns a middleware that injects a trace into the request context,
// picking up the trace id from the request header if it exists, or generating a new
// one if it doesn't. See clientmw.Trace for the client-side implementation
func Trace(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		traceID, err := uuid.Parse(r.Header.Get("X-Trace-Id"))
		if err != nil {
			traceID = uuid.New()
		}
		reqID, err := uuid.Parse(r.Header.Get("X-Request-Id"))
		if err != nil {
			reqID = uuid.New()
		}

		// add trace to context, and add context to request
		trace := trace.Trace{TraceID: traceID, RequestID: reqID}
		ctx = ctxutil.WithValue(ctx, trace)
		r = r.WithContext(ctx)

		// serve the request using the populated context
		h.ServeHTTP(w, r)
	}
}

// Log returns a middleware that injects a logger into a request context. See clientmw.Log for the client-side implementation
// It uses a trace from the context as a prefix, if it exists. For most servers, use a structured logger instead
// (slog from the stdlib, for instance); that API is out of scope for this
func Log(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trace, ok := ctxutil.Value[trace.Trace](r.Context())
		var prefix string
		if ok {
			prefix = fmt.Sprintf("%s %s: [%s %s]: ", r.Method, r.URL, trace.TraceID, trace.RequestID)
		} else {
			prefix = fmt.Sprintf("%s %s:", r.Method, r.URL)
		}
		//// yeah, confirmed here that clientmw.Log should do this
		logger := log.New(os.Stderr, prefix, log.LstdFlags)
		ctx := ctxutil.WithValue(r.Context(), logger)
		r = r.Clone(ctx)
		h.ServeHTTP(w, r)
	}
}

////
// Intercepting Writes to the Response
////

// RecordingResponseWriter is an http.ResponseWriter that keeps track of the status code and total body bytes
// written to it.
type RecordingResponseWriter struct {
	RW         http.ResponseWriter
	StatusCode int // first status code written to the response writer
	Bytes      int // total bytes written
}

// WriteHeader sets the status code, if it hasn't been set already.
func (w *RecordingResponseWriter) WriteHeader(statusCode int) {
	if w.StatusCode == 0 { // first status code written; track it
		w.StatusCode = statusCode
	}
	w.RW.WriteHeader(statusCode) // write to underlying response writer
}

// Header just returns the underlying response writer's header.
func (w *RecordingResponseWriter) Header() http.Header { return w.RW.Header() }

// Write writes the given bytes to the underlying response writer, setting the status code to 200 if it hasn't been set already.
func (w *RecordingResponseWriter) Write(b []byte) (int, error) {
	if w.StatusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.RW.Write(b) // write to underlying response writer
	w.Bytes += n            // update total bytes written
	return n, err
}

// RecordResponse returns a middleware that records the response status code and total bytes written to the response
func RecordResponse(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rrw := &RecordingResponseWriter{RW: w}
		start := time.Now()
		h.ServeHTTP(rrw, r)
		elapsed := time.Since(start)
		logger, ok := ctxutil.Value[*log.Logger](r.Context())
		if !ok {
			// fall back to the default logger
			log.Printf(
				"%s %s: %d %s: %d bytes in %s",
				r.Method, r.URL, rrw.StatusCode, http.StatusText(rrw.StatusCode), rrw.Bytes, elapsed,
			)
			return
		}
		logger.Printf("%d %s: %d bytes in %s", rrw.StatusCode, http.StatusText(rrw.StatusCode), rrw.Bytes, elapsed)
	}
}

////
// Panic recovery
////

func Recovery(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recover from panic
			if err := recover(); err != nil {
				stack := debug.Stack()
				logger, ok := ctxutil.Value[*log.Logger](r.Context())
				if !ok { // use the default logger
					log.Printf("%s %s: panic: %v\n%s", r.Method, r.URL, err, stack)
				} else { // use the logger from the context
					logger.Printf("panic: %v\n%s", err, stack)
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("internal server error"))
			}
		}()
		h.ServeHTTP(w, r)
	}
}
