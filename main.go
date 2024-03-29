package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"tailscale.com/net/proxymux"
	"tailscale.com/net/socks5"
	"tailscale.com/tsnet"
)

func handleProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodConnect {
			proxyConnect(w, req)
		} else if req.Method == http.MethodGet {
			proxyGet(w, req)
		} else {
			http.Error(w, "only supports CONNECT and GET", http.StatusMethodNotAllowed)
		}
	}
}

func proxyConnect(w http.ResponseWriter, r *http.Request) {
	targetConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Printf("Dial failed: %v: %v", r.Host, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer targetConn.Close()

	clientConn, clientBuf, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Println("Hijack failed:", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	if _, err = io.WriteString(clientConn, "HTTP/1.1 200 OK\r\n\r\n"); err != nil {
		log.Println("Write 200 OK failed:", err)
		return
	}

	log.Printf("CONNECT from %v to %v", r.RemoteAddr, r.Host)

	var clientSrc io.Reader = clientBuf
	if clientBuf.Reader.Buffered() == 0 {
		clientSrc = clientConn
	}

	errc := make(chan error, 1)
	go func() {
		_, err := io.Copy(clientConn, targetConn)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(targetConn, clientSrc)
		errc <- err
	}()
	err = <-errc
	if err != nil {
		log.Println("Copy failed:", err)
	}
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
	if err := http.Serve(ln, handleProxy()); err != nil {
		log.Fatal("serveHttp failed:", err)
	}
}

func serveSocks(ln net.Listener) {
	server := &socks5.Server{}
	if err := server.Serve(ln); err != nil {
		log.Fatal("serveSocks failed:", err)
	}
}

func listen(hostname, addr string) (net.Listener, func(), error) {
	if os.Getenv("TS_AUTHKEY") == "" {
		ln, err := net.Listen("tcp", addr)
		closeFn := func() {}
		return ln, closeFn, err
	} else {
		s := &tsnet.Server{
			Hostname:  hostname,
			Ephemeral: true,
		}
		closeFn := func() {
			_ = s.Close()
		}

		ln, err := s.Listen("tcp", addr)
		return ln, closeFn, err
	}
}

func main() {
	var (
		hostname = flag.String("hostname", "tailsocks", "hostname on tailnet, default is tailsocks")
		addr     = flag.String("addr", ":1080", "proxy address, default is :1080")
	)

	flag.Parse()

	ln, closeFn, err := listen(*hostname, *addr)
	if err != nil {
		log.Fatal("Listen failed:", err)
	}

	log.Println("Starting proxy server on", *addr)

	socksListener, httpListener := proxymux.SplitSOCKSAndHTTP(ln)
	go serveSocks(socksListener)
	serveHttp(httpListener)
	closeFn()
}
