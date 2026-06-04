// escapetext replaces the non-printable characters in a file and returns the replaced content. It does not modify the file
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: escapetext <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]

	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v", filename, err)
		os.Exit(1)
	}

	var buf strings.Builder
	buf.Grow(len(f))

	for _, c := range f {
		if strconv.IsPrint(rune(c)) {
			buf.WriteByte(c)
		} else {
			fmt.Fprintf(&buf, "%+q", c)
		}
	}

	fmt.Fprintf(os.Stdout, "%s\n", buf.String())
	os.Exit(0)
}
