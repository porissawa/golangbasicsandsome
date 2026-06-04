// echo prints its arguments to standard output, separated by spaces and terminated by a newline.
// usage: echo <args...>
package main

import (
	"fmt"
	"os"
)

func main() {
	// 1. iterate over the command line arguments
	for i, arg := range os.Args[1:] {
		if i > 0 {
			fmt.Printf(" ")
		}
		// 2. Print each argument to STDOUT, separated by spaces
		fmt.Print(arg)
	}
	// 3. Terminate with newline
	fmt.Println()
}
