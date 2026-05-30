// tcpupperecho serves tcp connection on port 8080, reading from each connection line-by-line and
// writing the uppercase version of each line back to the client

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	const name = "tcpupperecho"
	log.SetPrefix(name + "\t")

	port := flag.Int("p", 8080, "port to listen on")
	flag.Parse()

	// ListenTCP creates a TCP listener accepting connections on the given address.
	// TCPAddr represents the address of a TCP endpoint; it has an IP, Port and Zone, all of which are options
	// Zone only matters for IPV6; we'll ignore it for now.
	// If we omit the IP, it means we're listening on all available IP addresses; if we omit the Port, we're
	// listening on a random port.
	// We want to listen on a port specified by the user on the command-line. See
	// https://golang.org/pkg/net/#ListenTCP and https://golang.org/pkg/net/#Dial for details.
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: *port})
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	log.Printf("listening at localhost: %s", listener.Addr())

	// loop forever, accepting connections one at a time
	for {
		// Accept() blocks until a connection is made, then returns a Conn representing the connection
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go echoUpper(conn, conn)
	}
}

func echoUpper(w io.Writer, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		//// could have just called this inside of toUpper instead, no need to initialize and assign a variable for it
		text := scanner.Text()
		fmt.Fprintf(w, "%s\n", strings.ToUpper(text))
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error: %s", err)
	}
}
