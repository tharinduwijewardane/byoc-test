package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type OrderRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type InventoryResponse struct {
	InStock bool `json:"in_stock"`
}

func placeOrder(w http.ResponseWriter, r *http.Request) {
	log.Println("Received order request")

	var order OrderRequest
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		log.Printf("Error decoding request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("Order request decoded: ProductID=%s, Quantity=%d", order.ProductID, order.Quantity)

	inventoryResponse, err := checkInventory(order.ProductID)
	if err != nil {
		log.Printf("Error checking inventory: %v", err)
		http.Error(w, "Failed to check inventory", http.StatusInternalServerError)
		return
	}
	log.Printf("Inventory check result: InStock=%v", inventoryResponse.InStock)

	response := OrderResponse{Success: false}

	if inventoryResponse.InStock {
		response.Success = true
		response.Message = "Order placed successfully"
		log.Println("Order placed successfully")
	} else {
		response.Message = "Product is out of stock"
		log.Println("Order failed: Product out of stock")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("Order response sent")
}

func checkInventory(productID string) (InventoryResponse, error) {
	inventoryServiceURL := os.Getenv("INVENTORY_SERVICE_URL")
	apiKey := os.Getenv("INVENTORY_SERVICE_API_KEY")

	if inventoryServiceURL == "" {
		return InventoryResponse{}, fmt.Errorf("inventory service URL not set")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/check_inventory/%s", inventoryServiceURL, productID), nil)
	if err != nil {
		return InventoryResponse{}, err
	}

	req.Header.Set("API-Key", fmt.Sprintf("%s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return InventoryResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return InventoryResponse{}, err
	}

	var inventoryResponse InventoryResponse
	err = json.Unmarshal(body, &inventoryResponse)
	if err != nil {
		return InventoryResponse{}, err
	}

	return inventoryResponse, nil
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/place_order", placeOrder).Methods("POST")
	fmt.Println("Order Service is running on :9090")
	log.Fatal(http.ListenAndServe(":9090", router))
}
