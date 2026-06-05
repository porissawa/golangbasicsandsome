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

// Usage: shexdump [-offset] [-columns] [-column-width] [-squeeze] [-ascii] [filename]

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

type Options struct {
	columns, columnWidth          int
	offset, squeeze, ascii, canon bool
}

func main() {
	var opts Options

	{
		//// read flags + assign default vals
		flag.BoolVar(&opts.offset, "offset", false, "show offset in start of line")
		flag.BoolVar(&opts.ascii, "ascii", false, "show ascii representation of hex as a line suffix")
		flag.BoolVar(&opts.canon, "canon", false, "combines offset and ascii")
		flag.IntVar(&opts.columns, "columns", 2, "how many columns should be shown")
		flag.IntVar(&opts.columnWidth, "columnWidth", 8, "how wide each column is")
		flag.Parse()
	}

	//// more to adhere to the spec than anything
	if (opts.offset && opts.canon) || (opts.ascii && opts.canon) {
		fatalf("flag -canon cannot be used with -ascii or - offset")
	}

	// 1. choose the input source: stdin or a file
	var src io.Reader
	switch flag.NArg() {
	case 0:
		src = os.Stdin
	default:
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer f.Close()
		src = f
	}

	if err := hexdump(os.Stdout, src, opts); err != nil {
		fatalf("hexdump: %v\n", err)
	}
}

// dumps the contents of r to w in a hexdump format
func hexdump(dst io.Writer, src io.Reader, opts Options) error {
	// performance: small reads and writes are very inefficient. While we could write a byte at a time
	// it's much faster to read and write in chunks.
	r := bufio.NewReader(src)

	currOffset := 0
	currHexOffset := ""
	//// how many should I add to the buffer will be decided by these flags
	var offset, ascii int
	const hex = "0123456789abcdef"
	// default: 2 * 8 = 16
	hexCharPerLine := opts.columns * opts.columnWidth
	optsOffsetOrCanon := opts.offset || opts.canon
	optsAsciiOrCanon := opts.ascii || opts.canon

	if optsAsciiOrCanon {
		ascii = 16 + 4
	}

	for {
		raw := make([]byte, hexCharPerLine) // read buffer

		if optsOffsetOrCanon {
			currHexOffset = fmt.Sprintf("%.8x", currOffset)
			// "0x" + hex + " | "
			offset = 2 + len(currHexOffset) + 3
		}

		encoded := make([]byte, 0, hexCharPerLine*3+1+offset+ascii) // colums * columnWidth bytes, 3 characters per byte, 1 space between bytes, newline at the end

		n, err := io.ReadFull(r, raw[:])
		if n != 0 {
			if optsOffsetOrCanon {
				encoded = append(encoded, '0', 'x')
				for _, v := range currHexOffset {
					encoded = append(encoded, byte(v))
				}
				encoded = append(encoded, ' ', '|', ' ')
			}

			for i := range opts.columns {
				lastIdx := opts.columnWidth * (i + 1)
				initIdx := opts.columnWidth * i
				for j := initIdx; j < min(lastIdx, hexCharPerLine); j++ {
					//// raw[i]>>4 gets the top 4 bits of the byte (high nibble), by right-shifting by 4 bits (so the lower four are replaced)
					//// raw[i]&0x0f gets the lower 4 bits (the low nibble), as the binary representation for the mask is 00001111
					//// they each end up with a number from 0 to 15 ([0:(2^4 - 1)], since we're working with 4 bits of information)
					//// which is then used as an index to lookup the value in the `hex` array
					encoded = append(encoded, hex[raw[j]>>4], hex[raw[j]&0x0f], ' ')
				}
				encoded = append(encoded, ' ')
			}

			if optsAsciiOrCanon {
				encoded = append(encoded, '|', ' ')
				for _, v := range raw {
					encoded = append(encoded, v)
				}
				encoded = append(encoded, ' ', '|')
			}
			//// add the newline to end of 16 bits line
			encoded[len(encoded)-1] = '\n'
		}
		currOffset += hexCharPerLine
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
