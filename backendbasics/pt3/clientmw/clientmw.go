package clientmw

import (
	"dbs_and_more/ctxutil"
	"dbs_and_more/trace"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// RoundTripFunc is an adapter to allow the use of ordinary functions as RoundTrippers, a-la http.HandlerFunc
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements the RoundTripper interface by calling f(r)
func (f RoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// assert that RoundTripFunc implements http.RoundTripper at compile time
var _ http.RoundTripper = RoundTripFunc(nil)

// we'll use this helper function to log the beginning and end of each middleware. No need for this in the real world
// but it should help you understand what's going on
func logExec(name string) func() {
	//// notice how this gets used in the functions below. We call it with a defer and an immediate invocation
	//// so we get this first print line and then defer the returned func, which itself defers as well
	log.Printf("middleware: begin %s", name)
	return func() {
		defer log.Printf("middleware: end %s", name)
	}
}

// TimeRequest returns a RoundTripFunc that logs the duration of the request
// This is the simplest type of middleware, it just wraps the RoundTripper in a closure and accepts
// no other arguments.
func TimeRequest(rt http.RoundTripper) RoundTripFunc {
	return func(r *http.Request) (*http.Response, error) {
		// for demonstration purposes, we'll add these logs to each middleware. Don't do this in production!
		defer logExec("TimeRequest")()

		start := time.Now()
		// call next middleware, or http.DefaultTransport.RoundTrip if this is the last middleware
		resp, err := rt.RoundTrip(r)

		if err != nil {
			log.Printf("%s %s: errored after %s", r.Method, r.URL, time.Since(start))
			return nil, err
		}

		log.Printf("%s %s: %d %s in %s", r.Method, r.URL, resp.StatusCode, http.StatusText(resp.StatusCode), time.Since(start))
		return resp, nil
	}
}

// Here we have one that injects dependencies too
// RetryOn5xx returns a RoundTripFunc that retries the request up to n times if the server returns a 5xx status code.
// It will use exponential backoff: first retry will be after wait, second after 2*wait, third after 4*wait etc.
func RetryOn5xx(rt http.RoundTripper, wait time.Duration, tries int) RoundTripFunc {
	// validate arguments OUTSIDE of the closure, so that it only happens once
	if tries <= 1 {
		panic("tries must be > 1")
	}
	if wait <= 0 {
		panic("wait must be > 0")
	}

	return func(r *http.Request) (*http.Response, error) {
		defer logExec("RetryOn5xx")()
		var retryErrs error

		for retry := range tries {
			if retry > 0 {
				time.Sleep(wait << retry)
			}
			resp, err := rt.RoundTrip(r)
			if errors.Is(retryErrs, syscall.ECONNREFUSED) || errors.Is(retryErrs, syscall.ECONNRESET) {
				retryErrs = errors.Join(retryErrs, err)
				continue
			}
			if retryErrs != nil {
				return nil, fmt.Errorf("failed after %d retries: %w", retry, retryErrs)
			}
			switch sc := resp.StatusCode; {
			case sc >= 200 && sc < 400:
				return resp, nil
			case sc >= 400 && sc < 500:
				return nil, fmt.Errorf("failed after %d retries: %s", retry, resp.Status)
			default: // 1xx, 5xx or unknown status code
				retryErrs = errors.Join(retryErrs, fmt.Errorf("try %d: %s", retry, resp.Status))
			}
		}
		return nil, fmt.Errorf("failed after 3 retries: %w", retryErrs)
	}
}

// Trace applies the `X-Trace-ID` and `X-Request-ID` headers to requests, generating the first if needed
func Trace(rt http.RoundTripper) RoundTripFunc {
	return func(r *http.Request) (*http.Response, error) {
		defer logExec("Trace")()

		// does the request already have a trace? If so, use it. Otherwise, generate a new one
		traceID, err := uuid.Parse(r.Header.Get("X-Trace-ID"))
		if err != nil {
			traceID = uuid.New()
		}

		// build the trace. It's a small struct, so we put it directly in the context and don't bother
		// with a pointer
		trace := trace.Trace{TraceID: traceID, RequestID: uuid.New()}

		// add the trace to context. Retrieve with ctxutil.Value[Trace](ctx)
		ctx := ctxutil.WithValue(r.Context(), trace)

		// add context to request
		r = r.WithContext(ctx)

		// add trace id and request id to headers
		r.Header.Set("X-Trace-ID", trace.TraceID.String())
		r.Header.Set("X-Request-ID", trace.RequestID.String())

		// call next middleware, or http.DefaultTransport.RoundTrip if this is the last middleware
		return rt.RoundTrip(r)
	}
}

// Log returns a RoundTripFnc that logs the request duration and status code. It uses the trace from the context as a prefix,
// if it exists. See Trace in this package and servermw.Log for the server-side implementation
// Log supersedes TimeRequest
func Log(rt http.RoundTripper) RoundTripFunc {
	return func(r *http.Request) (*http.Response, error) {
		defer logExec("Log")()
		var prefix string

		trace, ok := ctxutil.Value[trace.Trace](r.Context())
		if ok {
			prefix = fmt.Sprintf("%s %s: [%s %s]: ", r.Method, r.URL, trace.TraceID, trace.RequestID)
		} else {
			prefix = fmt.Sprintf("%s %s: ", r.Method, r.URL)
		}

		//// The implementation in the article implies that log is received as an argument *log.Logger, but it also instantiates
		//// a logger here. If it's supposed to be a singleton, I can see passing it in from the argument and instantiating one
		//// on program start. Leaving this line here and removing from the params though, don't know how we will wire up the middlewares
		//// just yet.
		logger := log.New(os.Stderr, prefix, log.LstdFlags|log.Lshortfile)
		ctx := ctxutil.WithValue(r.Context(), logger) // add logger to context; retrieve with ctxutil.Value[log.Logger](ctx)
		r = r.WithContext(ctx)                        // add context to request

		start := time.Now()
		resp, err := rt.RoundTrip(r) // call next middleware, or http.DefaultTransport.RoundTrip if this is the last middleware
		if err != nil {
			logger.Printf("errored after %s: %s", time.Since(start), err)
			return nil, err
		}
		logger.Printf("%d %s in %s", resp.StatusCode, http.StatusText(resp.StatusCode), time.Since(start))
		return resp, nil
	}
}
