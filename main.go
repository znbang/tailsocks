package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"tailscale.com/net/proxymux"
	"tailscale.com/net/socks5"
	"tailscale.com/tsnet"
)

type proxyServer struct {
}

func (p *proxyServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodConnect {
		proxyConnect(w, req)
	} else if req.Method == http.MethodGet {
		proxyGet(w, req)
	} else {
		http.Error(w, "only supports CONNECT and GET", http.StatusMethodNotAllowed)
	}
}

func proxyConnect(w http.ResponseWriter, r *http.Request) {
	targetConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Printf("Dial failed: %v: %v", r.Host, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)

	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Fatal("Unable to hijack connection")
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Fatal("Hijack failed:", err)
	}

	log.Printf("CONNECT from %v to %v", r.RemoteAddr, r.Host)

	go tunnelConn(targetConn, clientConn)
	go tunnelConn(clientConn, targetConn)
}

func tunnelConn(dst io.WriteCloser, src io.ReadCloser) {
	_, _ = io.Copy(dst, src)
	_ = dst.Close()
	_ = src.Close()
}

func proxyGet(w http.ResponseWriter, r *http.Request) {
	target, err := url.Parse(r.URL.Scheme + "://" + r.URL.Host)
	if err != nil {
		log.Println("Parse failed:", err)
		return
	}

	log.Printf("GET from %v to %v", r.RemoteAddr, r.Host)

	p := httputil.NewSingleHostReverseProxy(target)
	p.ServeHTTP(w, r)
}

func serveHttp(ln net.Listener) {
	proxy := &proxyServer{}

	if err := http.Serve(ln, proxy); err != nil {
		log.Fatal("serveHttp failed:", err)
	}
}

func serveSocks(ln net.Listener) {
	server := &socks5.Server{}
	if err := server.Serve(ln); err != nil {
		log.Fatal("serveSocks failed:", err)
	}
}

func main() {
	var (
		hostname = flag.String("hostname", "tailsocks", "hostname on tailnet, default is tailsocks")
		addr     = flag.String("addr", ":1080", "proxy address, default is :1080")
	)

	flag.Parse()

	s := &tsnet.Server{
		Hostname: *hostname,
	}
	defer s.Close()

	ln, err := s.Listen("tcp", *addr)
	if err != nil {
		log.Fatal("Listen failed:", err)
	}

	log.Println("Starting proxy server on", *addr)

	socksListener, httpListener := proxymux.SplitSOCKSAndHTTP(ln)
	go serveSocks(socksListener)
	serveHttp(httpListener)
}
