// pardownload downloads a list of URLs in parallel, saving them to the specified directory
// It exits with a nonzero status code if any of the downloads fail, where the status code is the number of failed downloads
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
	"sync"
	"time"
)

func main() {
	var dstDir string
	// http.Client is safe for concurrent use across multiple goroutines. In general, you should reuse an http.Client
	// as much as possible. Avoid creating a new one for each request
	var client http.Client // the zero value of http.Client is a usable client
	flag.StringVar(&dstDir, "dst", "", "destination directory; defaults to current directory")
	flag.DurationVar(&client.Timeout, "timeout", 1*time.Minute, "timeout for the request")
	flag.Parse()

	// get the non-flag arguments (the ones captured above)
	src := flag.Args()
	if len(src) == 0 {
		log.Fatalf("usage: download [-dst directory to save file to] [-timeout duration] url1, url2, ..., urln")
	}
	// make the destination directory absolute so our error messages are easier to read
	dstDir, err := filepath.Abs(dstDir)
	if err != nil {
		log.Fatalf("invalid destination directory: %v", err)
	}
	// make a slice of the same length as src, so we can access it in parallel, without worrying about syncing
	//// wonder what makes a slice safe to access in parallel. I see why it's created, we want to populate the slice
	//// with a changed version of src and still need src to make the requests. Maybe that's what the author means by syncing
	//// not having to worry about what has already been read for its source url before changing it to a destination.
	//// It's one extra slice to keep in memory but simpler to reason with.
	dst := make([]string, len(src))
	for i := range src {
		// add filepaths to destination slice
		dst[i] = filepath.Join(dstDir, filepath.Base(src[i]))
	}

	// same for errors
	errs := make([]error, len(src))

	// a WaitGroup waits for a collection of goroutines to finish
	wg := new(sync.WaitGroup)
	// Add the number of goroutines we're going to wait for
	wg.Add(len(src))

	now := time.Now()

	for i := range src {
		//// this was in the original but is no longer needed since Go 1.22
		//// see https://go.dev/doc/faq#closures_and_goroutines for reference
		// i := i
		go func() {
			// tell the WaitGroup we're done. This is a simple function so we don't really need to defer, but it's a good habit
			// to get into
			defer wg.Done()
			errs[i] = downloadAndSave(context.TODO(), &client, src[i], dst[i])
		}()
	}
	// wait for the goroutines to finish
	wg.Wait()

	log.Printf("downloaded %d files in %v", len(src), time.Since(now))
	var errCount int
	for i := range errs {
		if errs[i] != nil {
			log.Printf("err: %s -> %s: %v", src[i], dst[i], errs[i])
			errCount++
		} else {
			log.Printf("ok: %s -> %s", src[i], dst[i])
		}
	}

	// if an error happened, it'll be a non-zero exit
	//// not sure how I feel about this, in my mind the error code should map to something and not be
	//// dependent on how many errors happened. That said, this behavior is described at the top of the file.
	os.Exit(errCount)
}

func downloadAndSave(ctx context.Context, c *http.Client, url, dst string) error {
	//// copying over the function from (download.go)[../download/download.go]
	//// check that one for comments.
	//// one change though: dir is no longer passed in since dst now has the complete path,
	//// which also means we're no longer creating the path in this function
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: GET %q: %v", url, err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response status: %s", resp.Status)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating file: %v", err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, resp.Body); err != nil {
		return fmt.Errorf("copying response to file: %v", err)
	}

	return nil
}
