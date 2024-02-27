package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {

	go func() {
		for i := 0; i < 1000000; i++ {
			log.Println("log with sleep test: " + strconv.Itoa(i))
			time.Sleep(100 * time.Millisecond)
		}
	}()

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
