package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func newProxy(s string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(s)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return proxy, nil

}

func handleProxy(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}

}

func main() {
	proxy, err := newProxy("http://httpbin.org/")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", handleProxy(proxy))
	http.ListenAndServe(":8080", nil)
}
