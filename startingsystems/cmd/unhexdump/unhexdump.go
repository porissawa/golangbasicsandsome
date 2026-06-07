// unhexdump reverses the process of hexdump, converting a hexdump back into a file.
// it expects pairs of whitespace-separated hexadecimal bytes
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	var src io.Reader
	switch len(os.Args) {
	case 1:
		src = os.Stdin
	case 2:
		f, err := os.Open(os.Args[1])
		if err != nil {
			fatalf("%v\n", err)
		}
		defer f.Close()
		src = f
	default:
		fatalf("Usage: %s [filename]", os.Args[0])
	}

	if err := unhexdump(os.Stdout, src); err != nil {
		fatalf("unhexdump: %v", err)
	}
}

func unhexdump(w io.Writer, r io.Reader) error {
	// we'll use buffered readers and writers here to reduce the number
	// of syscalls and allocations
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	for i := 0; scanner.Scan(); i++ {
		b := scanner.Bytes()
		//// 01 is 1, 10 is two, 11 is three, 100 four in binary. So if the least significant bit of the byte
		//// is 1, it means the length is odd and we can know that without a modulo operation, just comparing
		//// two bits
		if len(b)&1 == 1 {
			return fmt.Errorf("odd number of digits at position %d (%q)", i, b)
		}
		//// we know the length is even so we step in twos and read upper and lower 4-bit sections
		for i := 0; i < len(b); i += 2 {
			high, ok := unhex(b[i])
			if !ok {
				return fmt.Errorf("bad hex %x '%c' at position %d", b[i], b[i], i)
			}
			low, ok := unhex(b[i+1])
			if !ok {
				return fmt.Errorf("bad hex %x '%c' at position %d", b[i+1], b[i+1], i+1)
			}
			//// this WriteByte param looks like this:
			//// high = 1111 -> high<<<4 = 11110000
			//// low = 1010 -> high<<<4 | low = 11111010
			if err := bw.WriteByte(high<<4 | low); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func unhex(b byte) (byte, bool) {
	switch {
	case '0' <= b && b <= '9':
		return b - '0', true
	case 'a' <= b && b <= 'f':
		return b - 'a' + 10, true
	case 'A' <= b && b <= 'F':
		return b - 'A' + 10, true
	default:
		return 0, false
	}
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
