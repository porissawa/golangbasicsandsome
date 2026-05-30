// sendreq sends a request to the specified host, port and path, and prints the response to stdout
// flags: -host, -port, -path, -method

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	host, path, method string
	port               int
)

func main() {
	flag.StringVar(&method, "method", "GET", "HTTP method to use")
	flag.StringVar(&host, "host", "localhost", "host to connect to")
	flag.IntVar(&port, "port", 8080, "port to connect to")
	flag.StringVar(&path, "path", "/", "path to request")
	flag.Parse()

	// ResolveTCPAddr is a slightly more convenient way of creating a TCPAddr.
	// Now that we know how to do it by hand using net.LookupIP, we can use this instead.
	//// Check dns/dns.go to see the manual version
	ip, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, ip)
	if err != nil {
		panic(err)
	}

	log.Printf("connected to %s (@ %s)", host, conn.RemoteAddr())

	defer conn.Close()

	var reqfields = []string{
		fmt.Sprintf("%s %s HTTP/1.1", method, path),
		"Host: " + host,
		"User-Agent: httpget", //// Note the Title-Case for header keys
		"",                    // empty line to terminate the headers

		// body would go here if there was any
	}
	// e.g. for a request to http://google.com/
	// GET / HTTP/1.1
	// Host: google.com
	// User-Agent: httpget
	//

	//// build the request above as an actual string
	// windows-style line endings are part of the specification
	request := strings.Join(reqfields, "\r\n") + "\r\n"

	//// send the request by writing to the connection
	conn.Write([]byte(request))
	log.Printf("sent request:\n%s", request)

	for scanner := bufio.NewScanner(conn); scanner.Scan(); {
		line := scanner.Bytes()
		if _, err := fmt.Fprintf(os.Stdout, "%s\n", line); err != nil {
			log.Printf("error writing to connection: %s", err)
		}
		if scanner.Err() != nil {
			log.Printf("error reading from connection: %s", err)
			return
		}
	}
}
