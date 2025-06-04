package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyz"

func floodLogs(resWriter http.ResponseWriter, request *http.Request) {

	stopAt := time.Now().Add(2 * time.Minute)
	for {
		log.Println("log flooding test: logging current time: " + time.Now().String())
		if time.Now().After(stopAt) {
			break
		}
	}
}

func GenerateLogsOfSize(resWriter http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	bytesPerEntry, err := strconv.Atoi(query.Get("bytesPerEntry"))
	if err != nil || bytesPerEntry <= 10 {
		http.Error(resWriter, "Invalid bytesPerEntry value in query params. Value must be greater than 10 because a single log entry of '{\"msg\":\"\"}' uses 10 bytes", http.StatusBadRequest)
		return
	}

	entryCount, err := strconv.Atoi(query.Get("entryCount"))
	if err != nil || entryCount <= 0 {
		http.Error(resWriter, "Invalid entryCount value in query params. Value must be greater than 0", http.StatusBadRequest)
		return
	}

	sleepMsPerEntry, err := strconv.Atoi(query.Get("sleepMsPerEntry"))
	if err != nil || sleepMsPerEntry <= 0 {
		http.Error(resWriter, "Invalid sleepMsPerEntry value in query params. Value must be greater than 0", http.StatusBadRequest)
		return
	}

	resWriter.WriteHeader(http.StatusAccepted)
	resWriter.Write([]byte("Log generation started"))

	go func() {
		for i := 0; i < entryCount; i++ {
			logEntryMsg := make([]byte, bytesPerEntry)
			for i := range logEntryMsg {
				logEntryMsg[i] = letters[rand.Intn(len(letters))]
			}

			jsonOutput := map[string]string{"msg": string(logEntryMsg)}
			jsonBytes, err := json.Marshal(jsonOutput)
			if err != nil {
				os.Stderr.WriteString("Error marshaling JSON: " + err.Error() + "\n")
				continue
			}
			os.Stdout.Write(jsonBytes)
			os.Stdout.Write([]byte("\n"))

			time.Sleep(time.Duration(sleepMsPerEntry) * time.Millisecond)
		}
	}()
}

func main() {

	httpPort := 9090
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, "{\"haha\": true}")
	})

	http.HandleFunc("/flood-logs", floodLogs)

	http.HandleFunc("/gen-logs", GenerateLogsOfSize)

	fmt.Printf("listening on %v\n", httpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), http.DefaultServeMux)
	if err != nil {
		log.Fatal(err)
	}
}
