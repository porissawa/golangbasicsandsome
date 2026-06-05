// binpatch replaces a sequence of bytes in file starting at offset with a replacement string,
// and writes the result to standard output
// Usage: binpatch <file> <offset> <replacement>
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 4 {
		fatalf("Usage: %s <file> <offset> <replacement>\n", os.Args[0])
	}

	var (
		file        = os.Args[1]
		offset, err = strconv.ParseInt(os.Args[2], 0, 64)
		replacement = os.Args[3]
	)

	if err != nil {
		fatalf("invalid offset: %v\nUsage: %s <file> <offset> <replacement>\n", err, os.Args[0])
	}

	//// open file with read and write permissions
	f, err := os.OpenFile(file, os.O_RDWR, 0)
	if err != nil {
		fatalf("open %s: %v\n", file, err)
	}
	defer f.Close()

	//// write to STDOUT until offset
	_, err = io.CopyN(os.Stdout, f, offset)
	if err != nil {
		fatalf("copy: %v\n", err)
	}

	//// write replacement to stdout
	_, err = os.Stdout.Write([]byte(replacement))
	if err != nil {
		fatalf("write: %v\n", err)
	}

	//// walk file for the length of the replacement by writing to io.Discard, a noop
	if _, err := io.CopyN(io.Discard, f, int64(len(replacement))); err != nil {
		fatalf("copy: %v\n", err)
	}

	//// copy the rest of it. We don't need to worry about offsetting the first value because the file variable's
	//// Reader implementation is already doing so internally
	_, err = io.Copy(os.Stdout, f)
	if err != nil {
		fatalf("copy: %v\n", err)
	}

	//// Notice that at no point we allocate a buffer or anything like that. The whole point to this program is to
	//// eventually write stuff to STDOUT. So we just do that from the beginning, every write goes there, with no
	//// unnecessary intermediate step
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
