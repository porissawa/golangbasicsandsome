// cat reads each file specified on the command line and writes its contents to standard output.
// usage: cat <file1> [<file2> ...]
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	for _, file := range os.Args[1:] {
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open %s: %v", file, err)
			os.Exit(1)
		}
		// performance note, it's better to use io.Copy but this illustrates the process
		defer f.Close()

		b, err := io.ReadAll(f) // read file into memory
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v", file, err)
			os.Exit(1)
		}

		os.Stdout.Write(b) // write the contents to STDOUT
	}
}
