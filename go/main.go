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
	vars := mux.Vars(r)
	productID := vars["productId"]

	response := InventoryResponse{InStock: false}

	if quantity, exists := inventory[productID]; exists && quantity > 0 {
		response.InStock = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/check_inventory/{productId}", checkInventory).Methods("GET")

	fmt.Println("Inventory Service is running on :9090")
	log.Fatal(http.ListenAndServe(":9090", router))
}
