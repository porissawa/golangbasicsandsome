package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func basicGetExample() {
	ctx := context.TODO() // Always use a context. context.TODO if you don't know which one
	const method = "GET"
	const url = "https://eblog.fly.dev/index.html"
	var body io.Reader = nil // Empty body
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "deflate") // this is valid, Header.Add() works with a map[string][]string
	req.Header.Set("User-Agent", "eblog/1.0")    // Set() will set the value to a map[string][]string{value}, a single element list
	req.Header.Set("some-key", "a value")        // will be canonicalized to Some-Key (Add does the same)
	req.Header.Set("SOMe-keY", "somevalue")      // will overwrite the one above since it's using Set() and not Add()
	req.Write(os.Stdout)                         // serializes the request to the provided io.Writer
}

func getExampleWithQueryStringParams() {
	const method = "GET"
	v := make(url.Values)
	v.Add("q", `"of Emrakul"`) // note this uses the raw string syntax (`) to avoid having to escape the double quotes
	v.Add("order", "released")
	v.Add("dir", "asc")
	const path = "https://scryfall.com/search"
	dst := path + "?" + v.Encode() // Encode() escapes the values for us. Remember to add the "?" separator
	req, err := http.NewRequestWithContext(context.TODO(), method, dst, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Write(os.Stdout)
	// Since no User-Agent was manually added, Go automatically adds Go-http-client/1.1 for us
}

func postExample() {
	ctx := context.TODO()
	//// just changing the body not to be nil
	var body io.Reader = strings.NewReader("hello, world!")
	//// and the method, of course
	const method = "POST"
	const url = "https://eblog.fly.dev/index.html"
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("%v", req)
}
