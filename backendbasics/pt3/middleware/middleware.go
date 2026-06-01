package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"syscall"
	"time"
)

// This is what a custom function to replace Do() would look like applying the fuctionality
// we'd like to have shared across all requests. The idea is that we use it as a base for
// extracting middleware later (and show the limitations in this approach)
func DoRequest(ctx context.Context, c *http.Client, r *http.Request) (*http.Response, error) {
	// add context to request
	r = r.WithContext(ctx)
	// track execution time
	start := time.Now()
	// fun with defer: log elapsed time after function returns
	defer func() { log.Printf("request took %s", time.Since(start)) }()

	// add auth header to request
	// but how would this handle different services requiring different authorization headers?
	r = addAuthHeader(r)

	// retry logic
	var retryErrs error
	for retry := range uint(3) {
		if retry > 0 {
			//// exponential backoff
			time.Sleep(10 * time.Millisecond << retry)
		}
		resp, err := c.Do(r)
		if errors.Is(retryErrs, syscall.ECONNREFUSED) || errors.Is(retryErrs, syscall.ECONNRESET) {
			retryErrs = errors.Join(retryErrs, err)
			continue
		}
		if retryErrs != nil {
			return nil, fmt.Errorf("failed after %d retries: %w", retry, retryErrs)
		}
		switch sc := resp.StatusCode; {
		case sc >= 200 && sc < 400:
			return resp, nil //success
		case sc >= 400 && sc < 500: // 4xx status code
			return nil, fmt.Errorf("failed after %d retries: %s", retry, resp.Status)
		default:
			retryErrs = errors.Join(retryErrs, fmt.Errorf("try %d: %s", retry, resp.Status))
		}
	}
	return nil, fmt.Errorf("failed after 3 retries: %w", retryErrs)
}

// // clearly not how it should actually work, just example code
func addAuthHeader(r *http.Request) *http.Request {
	r.Header.Set("Authorization", "Token mock_token")
	return r
}
