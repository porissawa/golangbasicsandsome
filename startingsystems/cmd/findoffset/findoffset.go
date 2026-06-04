// findoffset.go is a command line tool that finds the offset of the first occurrence of a string in a file
// and prints it to stdout
package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	// 1. Parse command line arguments
	// the operating system provides command line arguments to your program. os.Args[0] is the name of the program,
	// and the rest are the 'real' aguments
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Fprintf(os.Stderr, "Usage: findoffset <filename> <string> [optional]<occurrence>\n")
		os.Exit(1)
	}

	filepath, pattern := os.Args[1], os.Args[2]

	//// check if the optional arguments where passed in and overwrite defaults
	targetOccurrence, foundTimes := 1, 0
	if len(os.Args) == 4 {
		ocArg, err := strconv.Atoi(os.Args[3])
		if err != nil || ocArg == 0 {
			fmt.Fprintf(os.Stderr, "Failed to parse occurrence argument, it must be a non-zero number\n")
			os.Exit(1)
		}
		targetOccurrence = ocArg
	}

	// 2. read the file into memory
	// it's inneficient to read the entire file into memory, but it's simple and works well for small files
	b, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v", filepath, err) // Human-readable debug info should go to STDERR
		os.Exit(1)
	}

	//// deal with optional arg
	if targetOccurrence > 0 {
		// 3. compare the bytes in the file to the bytes in the string, one-by-one
		//// len(b) - len(pattern) because if the first letter in the pattern isn't the same as the len - pattern, no
		//// need to keep on reading until all bytes are read, we know the pattern can't be there
		for i := 0; i < len(b)-len(pattern); i++ {
			for j := range pattern { //byte-by-byte comparison
				// no match, continue at next offset
				//// b[i+j] because the i is offsetting what has already been read, j is the index for the substring we're looking for
				if b[i+j] != pattern[j] {
					break
				}

				//// we didn't NOT match above and break out of the loop and we've also walked through
				//// the complete length of the pattern, so we have a match
				if j == len(pattern)-1 {
					foundTimes++
					if targetOccurrence == foundTimes {
						fmt.Fprintf(os.Stdout, "%d\n", i)
						os.Exit(0)
					}
				}
			}
		}
	}

	if targetOccurrence < 0 {
		//// if user passed a negative number, we do the lookup in reverse
		//// but we're safe to convert the target to a positive now for the comparison in the inner loop
		targetOccurrence = targetOccurrence * -1
		for i := len(b) - len(pattern); i > 0; i-- {
			for j := range pattern {
				//// we're moving backwards through the source but still compare it moving forwards
				if b[i+j] != pattern[j] {
					break
				}

				//// this one is still the same
				if j == len(pattern)-1 {
					foundTimes++
					if targetOccurrence == foundTimes {
						fmt.Fprintf(os.Stdout, "%d\n", i)
						os.Exit(0)
					}
				}
			}
		}
	}

	// 4. No match, exit 1
	//// or we somehow got in a state that targetOccurrence == 0, which shouldn't be possible anyway
	os.Exit(1)
}
