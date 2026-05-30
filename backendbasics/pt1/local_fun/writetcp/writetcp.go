// writetcp connects to a TCP server at localhost with the specified port (8080 by default)
// and forwards stdin to the server, line-by-line, until EOF is reached.
// Received lines from the server are printed to stdoud.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

//// whatever double comments (like this one) appear are mine (Pedro's)
//// mostly to give context on Go std lib functionality

func main() {
	const name = "writetcp"
	log.SetPrefix(name + "\t")

	// register the command line flags: -p specified the port to connect to
	port := flag.Int("p", 8080, "port to connect to")
	flag.Parse()

	//// net.DialTCP (and net.Dial, the general case) in here are connecting to localhost because of the params passed to it
	//// from its own function docs, DialTCP(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*net.TCPConn, error)
	//// if laddr is nil, localhost is assumed, if raddr doesn't have an ip, localhost is assumed.
	//// About the pointer to net.TCPConn. The struct is a Conn interface for TCP specifically, it implements Read and Write
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{Port: *port})
	if err != nil {
		log.Fatalf("error connecting to localhost:%d: %v", *port, err)
	}
	//// conn.RemoteAddr() returns the value of conn.fd.raddr. raddr, or Remote Address, is the address of the destination computer
	log.Printf("connected to %s: will forward stdin", conn.RemoteAddr())

	defer conn.Close()
	go func() {
		// spawn a goroutine to read incoming lines from the server and print the to stdout.
		// TCP is full-duplex, so we can read and write at the same time;
		// we just need to spawn a goroutine to do the reading.

		//// initialize a new Scanner struct, assign it to the connScanner variable.
		//// All Scanners have a default `split` function [SplitFunction type] that will advance until the next
		//// newline character. So this for loop has the `Scan` function as its condition, which loads the full line into connScanner with each
		//// iteration. `Scanner.Scan` will return false once it either reaches EOF (no error is assigned to the Scanner struct)
		//// or errors out (assigning an error, retrievable with the `Err` function)
		for connScanner := bufio.NewScanner(conn); connScanner.Scan(); {
			//// `Scanner.Text` then has the complete line up until the \n character, so print that and add a newline character
			//// as mentioned below
			fmt.Printf("%s\n", connScanner.Text()) // printf doesn't add a newline so we do it manually
			if err := connScanner.Err(); err != nil {
				log.Fatalf("error reading from %s: %v", conn.RemoteAddr(), err)
			}
		}
	}()

	// read incoming lines from stdin and forward them to the server
	// find next newline in stdin
	for stdinScanner := bufio.NewScanner(os.Stdin); stdinScanner.Scan(); {
		log.Printf("sent: %s\n", stdinScanner.Text())
		// Scanner.Bytes() returns a slice of bytes up to but not including the next \n
		// write that to the connection
		if _, err := conn.Write(stdinScanner.Bytes()); err != nil {
			log.Fatalf("error writing to %s: %v", conn.RemoteAddr(), err)
		}
		// so we add in the \n character
		if _, err := conn.Write([]byte("\n")); err != nil {
			log.Fatalf("error writing to %s: %v", conn.RemoteAddr(), err)
		}
		if stdinScanner.Err() != nil {
			log.Fatalf("error reading from %s: %v", conn.RemoteAddr(), err)
		}
	}
}
