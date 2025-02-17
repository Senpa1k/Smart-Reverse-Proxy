package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

type Proxy struct {
	proxy        *httputil.ReverseProxy
	blockedSites map[string]bool
}

func fillTheMap() map[string]bool {
	blockedSites := make(map[string]bool)
	file, err := os.Open("blockedSites.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		blockedSites[scanner.Text()] = true
	}
	return blockedSites
}

func newProxy() *Proxy {
	return &Proxy{
		proxy: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "https"
				req.URL.Host = req.Host
			},
		},
		blockedSites: fillTheMap(),
	}
}

func ProxyHTTPS(rw http.ResponseWriter, req *http.Request, p *Proxy) {
	str := ""
	for _, st := range req.Host {
		if st == ':' {
			break
		}
		str += string(st)
	}
	if p.blockedSites[str] {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("Forbidden"))
		return
	}
	fmt.Printf("Connected to %s\n", req.Host)
	serverCon, err := net.DialTimeout("tcp", req.Host, 10*time.Second)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	rw.WriteHeader(http.StatusOK)
	clientCon, _, err := rw.(http.Hijacker).Hijack()
	if err != nil {
		return
	}

	go transfer(clientCon, serverCon)
	go transfer(serverCon, clientCon)

}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func ProxyHTTP(rw http.ResponseWriter, req *http.Request, p *Proxy) {
	if p.blockedSites[req.Host] {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("Forbidden"))
		return
	}
	p.proxy.ServeHTTP(rw, req)
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "CONNECT" {
		ProxyHTTPS(rw, req, p)
	} else {
		ProxyHTTP(rw, req, p)
	}
}
func main() {
	proxy := newProxy()
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}))

}
