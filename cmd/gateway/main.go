package main

import (
	"net"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/weebon/distributed-rate-limiter/internal/limiter"
)

func main() {
	backendURL, _ := url.Parse("http://localhost:9090")
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	rl := limiter.NewManager(5, 1)

		handler := func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr // fallback if no port present
		}
		key := host

		if !rl.Allow(key) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		proxy.ServeHTTP(w, r)
	}

	http.HandleFunc("/", handler)
	fmt.Println("Gateway running on :8080")
	http.ListenAndServe(":8080", nil)
}