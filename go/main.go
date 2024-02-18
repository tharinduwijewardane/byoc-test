package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {

	stopAt := time.Now().Add(10 * time.Minute)
	for {
		log.Println("log flooding test: logging current time: " + time.Now().String())
		if time.Now().After(stopAt) {
			break
		}
	}

	httpPort := 9090
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, "{\"haha\": true}")
	})
	fmt.Printf("listening on %v\n", httpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), http.DefaultServeMux)
	if err != nil {
		log.Fatal(err)
	}
}
