package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Simplified inventory (in a real scenario, this would be a database)
var inventory = map[string]int{
	"123": 5,
	"456": 0,
	"789": 10,
}

type InventoryResponse struct {
	InStock bool `json:"in_stock"`
}

func checkInventory(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to check inventory")

	vars := mux.Vars(r)
	productID := vars["productId"]
	log.Printf("Checking inventory for product ID: %s", productID)

	// Check if the product ID is "500" and return a 500 status code if true
	if productID == "500" {
		log.Printf("Product ID 500 requested, returning HTTP 500 status code")
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
