package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Request struct {
	//// notice the json struct tag syntax (https://go.dev/ref/spec#Struct_types)
	//// a raw string literal (https://go.dev/ref/spec#string_lit) specifying the json encoding and field name in double quotes
	Format string `json:"format"` // Format as in time.Format. If empty, use time.RFC3339
	TZ     string `json:"tz"`     // TZ as in time.LoadLocation. If empty, use time.Local
}

type Resp struct {
	//// if using omitempty, it'd look like this: `json:"time,omitempty"` (https://pkg.go.dev/encoding/json#Marshal)
	Time string `json:"time"` // no need for omitempty here, will never send a zero time
}

type Error struct {
	Error string `json:"error"` // no need for omitempty here, will never send an empty error
}

// http handler: writes current time as JSON object (`{"Time": <time>}`)
// Original implementation
// func getTime(w http.ResponseWriter, r *http.Request) {
// 	var req Request
// 	// always set the Content-Type header manually, go tends to get confused when setting it automatically
// 	w.Header().Set("Content-Type", "encoding/json")
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		// failed to decode the request body, send back a bad request status
// 		w.WriteHeader(400)
// 		json.NewEncoder(w).Encode(Error{err.Error()})
// 	}
// 	//// we've decoded the the http.Request and wrote it to the local variable req
// 	//// we no longer need the connection, so close it
// 	r.Body.Close()

// 	var tz *time.Location = time.Local
// 	if req.TZ != "" {
// 		var err error
// 		tz, err = time.LoadLocation(req.TZ)
// 		if err != nil || tz == nil {
// 			w.WriteHeader(400)
// 			json.NewEncoder(w).Encode(Error{err.Error()})
// 			return
// 		}
// 	}
// 	format := time.RFC3339
// 	if req.Format != "" {
// 		format = req.Format
// 	}

// 	resp := Resp{time.Now().In(tz).Format(format)}
// 	json.NewEncoder(w).Encode(resp)
// }

// using the helpers below
func getTime(w http.ResponseWriter, r *http.Request) {
	//// could use Request as the type for ReadJSON here
	//// keeping the anonymous struct just to show the syntax
	req, err := ReadJSON[struct{ TZ, Format string }](r.Body)
	if err != nil {
		WriteError(w, err, 400)
		return
	}
	var tz *time.Location = time.Local
	if req.TZ != "" {
		var err error
		tz, err = time.LoadLocation(req.TZ)
		if err != nil {
			WriteError(w, err, 400)
			return
		}
	}
	format := time.RFC3339
	if req.Format != "" {
		format = req.Format
	}
	WriteJSON(w, Resp{time.Now().In(tz).Format(format)})
}

var client = &http.Client{Timeout: 2 * time.Second}

func sendRequest(tz, format string) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(Request{TZ: tz, Format: format})
	log.Printf("request body: %v", body)
	req, err := http.NewRequestWithContext(context.TODO(), "GET", "http://localhost:8080", body)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	resp.Write(os.Stdout)
	resp.Body.Close()
}

func main() {
	server := http.Server{Addr: ":8080", Handler: http.HandlerFunc(getTime)}
	go server.ListenAndServe()

	sendRequest("", "")
	sendRequest("America/Los_Angeles", time.RFC3339)
	sendRequest("America/New_York", time.RFC3339)
	sendRequest("America/Denver", "^^%noic2312..323") //// appendFormat() deals with the nonsense and defaults to -0700 in the stdTZ case
	sendRequest("badtz", "")
}

// Some helpers for working with JSON

// ReadJSON reads a JSON object from an io.ReadCloser, closing the reader when it's done. It's primarily useful for reading JSON from *http.Request.Body.
func ReadJSON[T any](r io.ReadCloser) (T, error) {
	var v T                               // declare a variable of type T
	err := json.NewDecoder(r).Decode(&v)  // decode the JSON into v
	return v, errors.Join(err, r.Close()) // close the reader and return any errors.
}

// WriteJSON writes a JSON object to a http.ResponseWriter, setting the Content-Type header to application/json.
func WriteJSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// WriteError logs an error, then writes it as a JSON object in the form {"error": <error>}, setting the Content-Type header to application/json.
func WriteError(w http.ResponseWriter, err error, code int) {
	log.Printf("%d %v: %v", code, http.StatusText(code), err) // log the error; http.StatusText gets "Not Found" from 404, etc.
	w.Header().Set("Content-Type", "encoding/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Error{err.Error()})
}
