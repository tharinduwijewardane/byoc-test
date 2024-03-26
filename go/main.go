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

	mux1 := http.NewServeMux()
	mux1.HandleFunc("/", ping)
	mux1.HandleFunc("/hello", hello)
	mux1.HandleFunc("/healthz", healthz)
	mux1.HandleFunc("/proxy", proxy)
	mux1.HandleFunc("/five", five)
	mux1.HandleFunc("/pp/{myParam}/five", ppMyParamFive)

	srv1 := &http.Server{
		Addr:         ":9091",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{mux1},
	}

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", ping)
	mux2.HandleFunc("/hello", hello)
	mux2.HandleFunc("/healthz", healthz)
	mux2.HandleFunc("/proxy", proxy)
	mux2.HandleFunc("/five", five)
	mux2.HandleFunc("/pp/{myParam}/five", ppMyParamFive)

	srv2 := &http.Server{
		Addr:         ":9092",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{mux2},
	}

	mux3 := http.NewServeMux()
	mux3.HandleFunc("/", ping)
	mux3.HandleFunc("/hello", hello)
	mux3.HandleFunc("/healthz", healthz)
	mux3.HandleFunc("/proxy", proxy)
	mux3.HandleFunc("/five", five)
	mux3.HandleFunc("/pp/{myParam}/five", ppMyParamFive)

	srv3 := &http.Server{
		Addr:         ":9093",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{mux3},
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		srv1.ListenAndServe()
	}()

	go func() {
		srv2.ListenAndServe()
	}()

	go func() {
		srv3.ListenAndServe()
	}()

	defer func() {
		if err := srv1.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the srv1 server: ", err)
		}
		if err := srv2.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the srv2 server: ", err)
		}
		if err := srv3.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the srv3 server: ", err)
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

func five(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "{\"stauts\": 500}")
}

func ppMyParamFive(w http.ResponseWriter, req *http.Request) {
	urlPath := req.URL.Path

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "{\"stauts\": 500, \"path\": \"%s\"}", urlPath)
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

	log.Printf("Method: %s, URL: %s\n", req.Method, req.URL.Path)
	log.Println("Request Headers:")
	for name, values := range req.Header {
		for _, value := range values {
			log.Printf("%s: %s\n", name, value)
		}
	}

	m.mux.ServeHTTP(rw, req)

	start := req.Context().Value("__requestStartTimer__").(time.Time)
	fmt.Println("request duration: ", time.Now().Sub(start))
}
