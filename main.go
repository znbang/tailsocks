package main

import (
	"flag"
	"fmt"
	"log"

	"tailscale.com/net/socks5"
	"tailscale.com/tsnet"
)

var (
	hostname = flag.String("h", "tailproxy", "hostname for socks5 server")
	port     = flag.Int("p", 1080, "port for socks5 server")
)

func main() {
	flag.Parse()

	s := &tsnet.Server{
		Hostname: *hostname,
	}
	defer s.Close()

	ln, err := s.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	server := &socks5.Server{}
	if err := server.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
