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

	mux0 := http.NewServeMux()
	mux0.HandleFunc("/", ping)
	mux0.HandleFunc("/hello", hello)
	mux0.HandleFunc("/healthz", healthz)
	mux0.HandleFunc("/proxy", proxy)
	mux0.HandleFunc("/five", five)
	mux0.HandleFunc("/four09", four09)
	mux0.HandleFunc("/pp/{myParam}/five", ppMyParamFive)

	srv0 := &http.Server{
		Addr:         ":9090",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{mux0, "9090"},
	}

	mux1 := http.NewServeMux()
	mux1.HandleFunc("/", ping)
	mux1.HandleFunc("/hello", hello)
	mux1.HandleFunc("/healthz", healthz)
	mux1.HandleFunc("/proxy", proxy)
	mux1.HandleFunc("/five", five)
	mux1.HandleFunc("/four09", four09)
	mux1.HandleFunc("/pp/{myParam}/five", ppMyParamFive)

	srv1 := &http.Server{
		Addr:         ":9091",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{mux1, "9091"},
	}

	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", ping)
	mux2.HandleFunc("/hello", hello)
	mux2.HandleFunc("/healthz", healthz)
	mux2.HandleFunc("/proxy", proxy)
	mux2.HandleFunc("/five", five)
	mux2.HandleFunc("/four09", four09)
	mux2.HandleFunc("/pp/{myParam}/five", ppMyParamFive)

	srv2 := &http.Server{
		Addr:         ":9092",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      middleware{mux2, "9092"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		srv0.ListenAndServe()
	}()

	go func() {
		srv1.ListenAndServe()
	}()

	go func() {
		srv2.ListenAndServe()
	}()

	defer func() {
		if err := srv0.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the srv0 server: ", err)
		}
		if err := srv1.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the srv1 server: ", err)
		}
		if err := srv2.Shutdown(ctx); err != nil {
			fmt.Println("error when shutting down the srv2 server: ", err)
		}
	}()

	sig := <-sigs
	fmt.Println(sig)

	cancel()

	fmt.Println("service has shutdown")
}

func ping(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m := map[string]string{
		"active": "true",
		"port":   req.Header.Get("port"),
	}
	_ = json.NewEncoder(w).Encode(m)
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
	m := map[string]string{
		"status": "500",
		"port":   req.Header.Get("port"),
	}
	_ = json.NewEncoder(w).Encode(m)
}

func four09(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusConflict)
	m := map[string]string{
		"status": "409",
		"port":   req.Header.Get("port"),
	}
	_ = json.NewEncoder(w).Encode(m)
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
	mux    http.Handler
	epName string
}

func (m middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := context.WithValue(req.Context(), "user", "unknown")
	ctx = context.WithValue(ctx, "__requestStartTimer__", time.Now())
	req = req.WithContext(ctx)

	log.Printf("EP: %s Method: %s, URL: %s\n", m.epName, req.Method, req.URL.Path)
	req.Header.Set("port", m.epName)
	headers := ""
	for name, values := range req.Header {
		for _, value := range values {
			headers += fmt.Sprintf("%s: %s\t", name, value)
		}
	}
	log.Println("Request Headers: ", headers)

	m.mux.ServeHTTP(rw, req)

	start := req.Context().Value("__requestStartTimer__").(time.Time)
	fmt.Println("request duration: ", time.Now().Sub(start))
}
