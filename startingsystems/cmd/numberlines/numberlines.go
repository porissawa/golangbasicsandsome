// numberlines counts the number of lines in the file and returns it to the user
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: numberlines <filename>")
		os.Exit(1)
	}

	filepath := os.Args[1]
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	fileContent, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", filepath, err)
		os.Exit(1)
	}

	var lineCount = 1

	for _, c := range fileContent {
		if c == '\n' {
			lineCount++
		}
	}

	fmt.Fprintf(os.Stdout, "%d\n", lineCount)
	os.Exit(0)
}
