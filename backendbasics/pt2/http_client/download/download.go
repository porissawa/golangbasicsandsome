// download is a command-line tool to download a file from a URL
// usage: download [-dir directory to save file to] [-timeout duration] url filename
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func dl() {
	dir := flag.String("dir", ".", "directory to save file")
	timeout := flag.Duration("timeout", 30*time.Second, "timeout for download")
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatal("usage: download [-dir directory to save file to] [-timeout duration] url filename")
	}
	url, filename := args[0], args[1]
	// always set a timeout when making a request
	c := http.Client{Timeout: *timeout}

	// we'll go into context (https://pkg.go.dev/context) and (https://go.dev/blog/context) later on
	//// https://eblog.fly.dev/backendbasics2.html#6-contexts-and-cancellation
	//// context is used to
	//// 1. carry metadata about a request across function boundaries
	//// 2. A ceiling on execution time (to cancel any function that does I/O like network requests, db queries and file operations)
	////    It should be the first argument for functions that do I/O. If you do cancel a context, return an error
	if err := downloadAndSave(context.TODO(), &c, url, filename, *dir); err != nil {
		log.Fatal(err)
	}
}

func downloadAndSave(ctx context.Context, c *http.Client, url, dst, dir string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: GET %q: %v", url, err)
	}
	// Do() serializes an http.Request, sends it to the server, and then deserializes the response to an http.Response
	resp, err := c.Do(req)
	// always check for errors after calling Do, they usually mean something went wrong on the network.
	if err != nil {
		return fmt.Errorf("request: %v", err)
	}
	// always close the body when you're done. Take advantage of the defer keyword and do it right after error checking the request
	//// closing the body means closing the connection to the server
	defer resp.Body.Close()

	// if Do didn't return an error, we still need to know if the request had a 200 status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response status: %s", resp.Status)
	}

	// response was succesful, let's save it to a file
	dstPath := filepath.Join(dir, dst)
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("creating file: %v", err)
	}
	// always close files you've opened in the FS
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, resp.Body); err != nil {
		return fmt.Errorf("copying response to file: %v", err)
	}

	return nil
}
