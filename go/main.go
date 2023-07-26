package main

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	SugaredLogger := logger.Sugar()

	httpPort := 9090
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		SugaredLogger.Debug("This is a debug log")
		SugaredLogger.Info("This is an info log")
		SugaredLogger.Warn("This is a warn log")
		SugaredLogger.Error("This is an error log")

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
			body, err := ioutil.ReadAll(resp.Body)
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
