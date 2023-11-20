package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
)

func main() {
	httpPort := 9090

	http.HandleFunc("/squareroot/", func(w http.ResponseWriter, req *http.Request) {
		numberStr := req.URL.Query().Get("number")
		if numberStr == "" {
			log.Printf("Error: number can't be empty")
			http.Error(w, "Error: number can't be empty", http.StatusBadRequest)
			return
		}
		num, err := strconv.Atoi(numberStr)
		if err != nil {
			log.Printf("Error: number must be an integer")
			http.Error(w, "Error: number must be an integer", http.StatusBadRequest)
			return
		}
		sqrt := math.Sqrt(float64(num))
		log.Printf("Square root of %d is %f", num, sqrt)
		_, _ = fmt.Fprintf(w, "Square root of %d is %f\n", num, sqrt)
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
