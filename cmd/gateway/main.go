package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/weebon/distributed-rate-limiter/internal/config"
	"github.com/weebon/distributed-rate-limiter/internal/store"
)

func main() {
	backendAddr := os.Getenv("BACKEND_ADDR")
   if backendAddr == "" {
	backendAddr = "http://localhost:9090"
	}
   backendURL, _ := url.Parse(backendAddr)
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	cfg := config.NewDefaultConfig()
 redisAddr := os.Getenv("REDIS_ADDR")
if redisAddr == "" {
	redisAddr = "localhost:6379"
}
tokenLimiter := store.NewRedisLimiter(redisAddr)
slidingLimiter := store.NewRedisSlidingWindow(redisAddr)

	algo := os.Getenv("ALGO")
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		routeLimit := cfg.ForRoute(r.URL.Path)
		key := host + ":" + r.URL.Path // per-client AND per-route

		var allowed bool
		var checkErr error

		if algo == "sliding_window" {
			allowed, checkErr = slidingLimiter.Allow(ctx, key, routeLimit.Limit, routeLimit.WindowSecs)
		} else {
			allowed, checkErr = tokenLimiter.Allow(ctx, key, routeLimit.Capacity, routeLimit.RefillRate)
		}

		if checkErr != nil {
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