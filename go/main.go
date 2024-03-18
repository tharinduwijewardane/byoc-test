package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
			path := data["path"]

			if len(host) == 0 {
				if h, ok := os.LookupEnv("HOST"); ok {
					host = h
				} else {
					host = "http://postman-echo.com"
				}
			}
			if len(path) == 0 {
				if a, ok := os.LookupEnv("PATH"); ok {
					path = a
				} else {
					path = "get?foo1=bar1&foo2=bar2"
				}
			}

			resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(host, "/"), strings.TrimLeft(path, "/")))
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
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
