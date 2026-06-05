// shexdump.go dumps the input as pairs of space-separated hexadecimal bytes
// with a newline after every 16 bytes.

// #example
//  //	#!usr/bin/env/bash
//  //	echo "now is the time for all good men to come to the aid of their country" | shexdump
//  //	6e 6f 77 20 69 73 20 74  68 65 20 74 69 6d 65 20
//	//	6f 66 20 61 6c 6c 20 67  6f 6f 64 20 6d 65 6e 20
//	//	74 6f 20 63 6f 6d 65 20  74 6f 20 74 68 65 20 61
//	//	69 64 20 6f 66 20 74 68  65 69 72 20 63 6f 75 6e
//	//	74 72 79 20 0a

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	// 1. choose the input source: stdin or a file
	var src io.Reader
	switch len(os.Args) {
	case 1:
		src = os.Stdin
	case 2:
		f, err := os.Open(os.Args[1])
		if err != nil {
			fatalf("open %s: %v", os.Args[1], err)
		}
		defer f.Close()
		src = f
	default:
		fatalf("Usage: %s [filename]", os.Args[0])
	}
	if err := hexdump(os.Stdout, src); err != nil {
		fatalf("hexdump: %v", err)
	}
}

// dumps the contents of r to w in a hexdump format
func hexdump(dst io.Writer, src io.Reader) error {
	// performance: small reads and writes are very inefficient. While we could write a byte at a time
	// it's much faster to read and write in chunks.
	r := bufio.NewReader(src)
	for {
		var raw [16]byte                       // read 16 bytes at a time
		encoded := make([]byte, 0, 16*(3+1)+1) // 16 bytes, 3 characters per byte, 1 space between bytes, newline at the end
		n, err := io.ReadFull(r, raw[:])
		// convert each byte in the chunk to a pair of hexadecimal digits
		const hex = "0123456789abcdef"
		if n != 0 {
			for i := range min(n, 8) {
				//// raw[i]>>4 gets the top 4 bits of the byte (high nibble), by right-shifting by 4 bits (so the lower four are replaced)
				//// raw[i]&0x0f gets the lower 4 bits (the low nibble), as the binary representation for the mask is 00001111
				//// they each end up with a number from 0 to 15 ([0:(2^4 - 1)], since we're working with 4 bits of information)
				//// which is then used as an index to lookup the value in the `hex` array
				encoded = append(encoded, hex[raw[i]>>4], hex[raw[i]&0x0f], ' ')
			}
			encoded = append(encoded, ' ')
			for i := 8; i < min(n, 16); i++ {
				encoded = append(encoded, hex[raw[i]>>4], hex[raw[i]&0x0f], ' ')
			}
			//// add the newline to end of 16 bits line
			encoded[len(encoded)-1] = '\n'
		}
		if _, err := dst.Write(encoded); err != nil {
			return err
		}
		if err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
