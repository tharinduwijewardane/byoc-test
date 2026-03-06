package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	httpPort := 9090
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"active\": true}")
	})
	http.HandleFunc("/healthz/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"healthy\": true}")
	})
	http.HandleFunc("/greeting", func(w http.ResponseWriter, req *http.Request) {
		name := req.URL.Query().Get("name")
		if name == "" {
			name = "world"
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"greeting\": \"Hello %s from internal service!\", \"timestamp\": \"%s\"}", name, time.Now().Format(time.RFC3339))
	})

	fmt.Printf("listening on %v\n", httpPort)

	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
