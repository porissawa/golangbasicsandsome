// torso reads the "middle" of a file, the bytes around a give offset.
// It's not the head not the tail, it's the torso.
// usage:
// torso -offset n -before [b=128] -after [a=128] -from file [-newline]
//
// if no file is given, reads from standard input
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	var offset, before, after int
	var from string
	var newline bool
	{
		flag.IntVar(&offset, "offset", -1, "offset to read from: must be specified")
		flag.IntVar(&before, "before", 128, "bytes to read before offset: will be clamped to 0")
		flag.IntVar(&after, "after", 128, "bytes to read after offset: will be clamped to 0")
		flag.StringVar(&from, "from", "", "file to read from: if empty, reads from standard input")
		flag.BoolVar(&newline, "newline", false, "append a newline to the output")
		flag.Parse()
	}

	{
		before = max(before, 0)      // can't be negative
		before = min(before, offset) // can't go past the beginning
		after = max(after, 0)
		if offset < 0 {
			fmt.Fprintf(os.Stderr, "missing or invalid -offset\n")
			os.Exit(1)
		}
	}

	start := offset - before // byte to start from
	n := before + after      // how many bytes to read
	if n == 0 {
		return
	}

	//// allocate exactly the size we want. Useful not to waste memory
	//// and because we'll write to this buffer later on,
	//// so we don't waste ops either
	buf := make([]byte, n)

	f, err := os.Open(from)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %s: %v", from, err)
		os.Exit(1)
	}
	//// using defer instead of closing on each failure + end of program
	defer f.Close()

	// Skip to first byte we want to read
	_, err = f.Seek(int64(start), io.SeekStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seek: %s: %v\n", from, err)
		os.Exit(1)
	}

	// read into memory
	n, err = io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		fmt.Fprintf(os.Stderr, "read: %s: %v", from, err)
		os.Exit(1)
	}

	// write to stdout
	_, err = os.Stdout.Write(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}

	if newline {
		fmt.Println()
	}
}
