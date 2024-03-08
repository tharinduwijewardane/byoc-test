package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Starting server")

	helloMux := http.NewServeMux()
	helloMux.HandleFunc("/", ping)
	helloMux.HandleFunc("/hello", hello)
	helloMux.HandleFunc("/healthz", healthz)
	helloMux.HandleFunc("/proxy", proxy)

	helloSrv := &http.Server{
		Addr:         "127.0.0.1:9090",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{helloMux},
	}

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/", ping)
	adminMux.HandleFunc("/hello", hello)
	adminMux.HandleFunc("/healthz", healthz)
	adminMux.HandleFunc("/proxy", proxy)

	adminSrv := &http.Server{
		Addr:         "127.0.0.1:9091",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{adminMux},
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		helloSrv.ListenAndServe()
	}()

	go func() {
		adminSrv.ListenAndServe()
	}()

	defer func() {
		if err := helloSrv.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the main server: ", err)
		}
		if err := adminSrv.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the admin server: ", err)
		}
	}()

	sig := <-sigs
	fmt.Println(sig)

	cancel()

	fmt.Println("service has shutdown")
}

func ping(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"active\": true}")
}

func healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"healthy\": true}")
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello %s", req.URL.Query().Get("name"))
}

func proxy(w http.ResponseWriter, req *http.Request) {
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
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

type middleware struct {
	mux http.Handler
}

func (m middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := context.WithValue(req.Context(), "user", "unknown")
	ctx = context.WithValue(ctx, "__requestStartTimer__", time.Now())
	req = req.WithContext(ctx)

	m.mux.ServeHTTP(rw, req)

	start := req.Context().Value("__requestStartTimer__").(time.Time)
	fmt.Println("request duration: ", time.Now().Sub(start))
}
