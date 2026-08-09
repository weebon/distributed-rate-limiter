package main

import (
	"os"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/weebon/distributed-rate-limiter/internal/store"
)

func main() {
	backendURL, _ := url.Parse("http://localhost:9090")
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	rl := store.NewRedisLimiter("localhost:6379", 5, 1)
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		allowed, err := rl.Allow(ctx, host)
		if err != nil {
			http.Error(w, "Internal error checking rate limit", http.StatusInternalServerError)
			return
		}
		if !allowed {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		proxy.ServeHTTP(w, r)
	}

	http.HandleFunc("/", handler)
	 
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Gateway running on :" + port)
	http.ListenAndServe(":"+port, nil)
}