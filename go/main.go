package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Simplified inventory (in a real scenario, this would be a database)
var inventory = map[string]int{
	"123": 5,
	"456": 0,
	"789": 10,
	"500": 5,
}

type InventoryResponse struct {
	InStock bool `json:"in_stock"`
}

type RequestTracker struct {
	requests []time.Time
	mutex    sync.Mutex
}

var tracker = RequestTracker{
	requests: make([]time.Time, 0),
}

func (rt *RequestTracker) AddRequest() {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	now := time.Now()
	rt.requests = append(rt.requests, now)

	// Remove requests older than 5 seconds
	for len(rt.requests) > 0 && now.Sub(rt.requests[0]) > 5*time.Second {
		rt.requests = rt.requests[1:]
	}
}

func (rt *RequestTracker) CountRequests() int {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	return len(rt.requests)
}

func checkInventory(w http.ResponseWriter, r *http.Request) {
	requestCount := tracker.CountRequests()
	tracker.AddRequest()
	log.Printf("Received request to check inventory. Requests in last 5 seconds: %d", requestCount)

	vars := mux.Vars(r)
	productID := vars["productId"]
	log.Printf("Checking inventory for product ID: %s", productID)

	// Check if the product ID is "500" and return a 500 status code if true
	if productID == "500" && requestCount == 0 {
		log.Printf("Product ID 500 requested for the first time. Simiulating a 500 error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := InventoryResponse{InStock: false}

	if quantity, exists := inventory[productID]; exists {
		log.Printf("Product %s found in inventory. Quantity: %d", productID, quantity)
		if quantity > 0 {
			response.InStock = true
			log.Printf("Product %s is in stock", productID)
		} else {
			log.Printf("Product %s is out of stock", productID)
		}
	} else {
		log.Printf("Product %s not found in inventory", productID)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("Response sent for product %s: %+v", productID, response)
}

func main() {
	log.Println("Starting Inventory Service")

	router := mux.NewRouter()

	router.HandleFunc("/check_inventory/{productId}", checkInventory).Methods("GET")

	fmt.Println("Inventory Service is running on :9090")
	log.Fatal(http.ListenAndServe(":9090", router))
}
