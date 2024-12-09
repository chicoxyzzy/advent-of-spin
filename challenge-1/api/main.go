package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/kv"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Handle GET request
			store, err := kv.OpenStore("default")
			if err != nil {
				http.Error(w, fmt.Sprintf("Error opening store: %v", err), http.StatusInternalServerError)
				return
			}
			defer store.Close()

			keys, err := store.GetKeys()
			if err != nil {
				http.Error(w, fmt.Sprintf("Error getting keys: %v", err), http.StatusInternalServerError)
				return
			}

			results := []map[string]interface{}{}
			for _, key := range keys {
				value, err := store.Get(key)
				if err != nil {
					http.Error(w, fmt.Sprintf("Error getting value for key %s: %v", key, err), http.StatusInternalServerError)
					return
				}

				// Decode the JSON-encoded value back into a map
				var storedData struct {
					Name  string   `json:"name"`
					Items []string `json:"items"`
				}
				if err := json.Unmarshal(value, &storedData); err != nil {
					http.Error(w, fmt.Sprintf("Error decoding items for key %s: %v", key, err), http.StatusInternalServerError)
					return
				}

				results = append(results, map[string]interface{}{
					"name":  storedData.Name,
					"items": storedData.Items,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(results); err != nil {
				http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
				return
			}

		case http.MethodPost:
			// Handle POST request
			store, err := kv.OpenStore("default")
			if err != nil {
				http.Error(w, fmt.Sprintf("Error opening store: %v", err), http.StatusInternalServerError)
				return
			}
			defer store.Close()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			// Define the expected data structure
			var inputData struct {
				Name  string   `json:"name"`
				Items []string `json:"items"`
			}

			// Unmarshal JSON into the struct
			if err := json.Unmarshal(body, &inputData); err != nil {
				http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
				return
			}

			// Validate the input data
			if inputData.Name == "" || len(inputData.Items) == 0 {
				http.Error(w, "Invalid data: 'name' must be non-empty and 'items' must contain at least one item.", http.StatusBadRequest)
				return
			}

			// Convert the entire input structure to JSON for storage
			storedBytes, err := json.Marshal(inputData)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error encoding items: %v", err), http.StatusInternalServerError)
				return
			}

			// Store the data in the key-value store
			if err := store.Set(inputData.Name, storedBytes); err != nil {
				http.Error(w, fmt.Sprintf("Error storing data: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			store, err := kv.OpenStore("default")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer store.Close()

			keys, err := store.GetKeys()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, key := range keys {
				if err := store.Delete(key); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			// Handle unsupported methods
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func main() {}
