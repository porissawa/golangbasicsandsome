// dns is a simple command line tool to lookup the ip address of a host
// it prints the first ipv4 and ipv6 addresses it finds, or "none" if none are found
package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Printf("%s: usage: <host>", os.Args[0])
		log.Fatalf("Expected exactly one argument; got %d", len(os.Args)-1)
	}
	host := os.Args[1]
	ips, err := net.LookupIP(host)
	if err != nil {
		log.Fatalf("lookup ip: %s: %v", host, err)
	}
	if len(ips) == 0 {
		log.Fatalf("no ips found for %s", host) // shouldn't happen but just in case
	}
	// print the first ipv4 we find
	for _, ip := range ips {
		if ip.To4() != nil {
			fmt.Println(ip)
			goto IPV6 // what's that, Djikstra?
		}
	}
	fmt.Println("none") // we're skipping this with the goto statement if an ipv4 is found

IPV6:
	for _, ip := range ips {
		//// == nil since To4() returns nil if the ip is not and ipv4, which is what we want here
		if ip.To4() == nil {
			fmt.Println(ip) // we know at least one ip exists, no need to check for nils
			return
		}
	}
	fmt.Println("none")
}
