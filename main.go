package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/things-go/go-socks5"
	"tailscale.com/tsnet"
)

var (
	hostname = flag.String("h", "tailsocks5", "hostname for socks5 server")
	port = flag.Int("p", 1080, "port for socks5 server")
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

	server := socks5.NewServer(
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
	)

	if err := server.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
