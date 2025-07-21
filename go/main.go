package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
	http.HandleFunc("/hello/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello %s", req.URL.Query().Get("name"))
	})
	http.HandleFunc("/print-body", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "{\"error\": \"Method not allowed. Use POST.\"}")
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\": \"Failed to read request body\"}")
			return
		}

		// Print the body to console/logs
		log.Printf("Received POST body: %s", string(body))

		// Return confirmation with the body content
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"message\": \"Body received and printed\", \"body_length\": %d}", len(body))
	})
	http.HandleFunc("/five", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"status\": 500}")
	})
	http.HandleFunc("/four09", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "{\"status\": 409}")
	})
	http.HandleFunc("/proxy/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			decoder := json.NewDecoder(req.Body)
			var data map[string]string
			err := decoder.Decode(&data)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
			}

			host := data["host"]
			args := data["args"]

			if len(host) == 0 {
				host = "http://postman-echo.com"
			}
			if len(args) == 0 {
				args = "get?foo1=bar1&foo2=bar2"
			}

			resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(host, "/"), strings.TrimLeft(args, "/")))
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(body)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
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
