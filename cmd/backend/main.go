package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from backend!")
}

func main() {
	http.HandleFunc("/api/test", hello)
	fmt.Println("Backend running on :9090")
	http.ListenAndServe(":9090", nil)
}