package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from backend!")
}

func search(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Search results here")
}

func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Upload accepted")
}

func main() {
	http.HandleFunc("/api/test", hello)
	http.HandleFunc("/api/search", search)
	http.HandleFunc("/api/upload", upload)
	fmt.Println("Backend running on :9090")
	http.ListenAndServe(":9090", nil)
}