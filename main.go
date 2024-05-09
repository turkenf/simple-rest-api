package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	stdsort "sort"
	"strings"
	"time"

	"github.com/google/uuid"
	yaml "gopkg.in/yaml.v3"
)

type item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

var items []item

func main() {

	http.HandleFunc("/items", itemsHandler)
	http.HandleFunc("/items/", itemsHandlerByID)

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getItems(w, r)
	} else if r.Method == "POST" {
		addItem(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func itemsHandlerByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getItemsByID(w, r)
	} else if r.Method == "DELETE" {
		deleteItemsByID(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getItemsByID(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from path
	id := r.URL.Path[len("/items/"):]

	// Find item by ID
	var foundItem item
	itemFound := false
	for _, item := range items {
		if item.ID == id {
			foundItem = item
			itemFound = true
			break
		}
	}

	// Check if an item exists with the specified ID
	if !itemFound {
		http.Error(w, "There is no item with the specified ID", http.StatusNotFound)
		return
	}

	format, _, err := parseQuery(r.URL.RawQuery)

	if format == "yaml" {
		err = yaml.NewEncoder(w).Encode(foundItem)
	} else {
		err = json.NewEncoder(w).Encode(foundItem)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func deleteItemsByID(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from path
	id := r.URL.Path[len("/items/"):]

	// Find the index of the item with the specified ID
	index := -1
	for i, item := range items {
		if item.ID == id {
			index = i
			break
		}
	}

	// Check if an item exists with the specified ID
	if index == -1 {
		http.Error(w, "There is no item with the specified ID", http.StatusNotFound)
		return
	}

	// Remove the item from the slice
	items = append(items[:index], items[index+1:]...)

	// Respond with success
	w.WriteHeader(http.StatusOK)
}

func addItem(w http.ResponseWriter, r *http.Request) {

	var newItem item
	// Decode the JSON data from the request body into the newItem variable
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if name is provided
	if newItem.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if newItem.ID == "" {
		newItem.ID = uuid.NewString()
	} else {
		// Check if provided ID is duplicate
		for _, existingItem := range items {
			if newItem.ID == existingItem.ID {
				http.Error(w, "ID already exists", http.StatusBadRequest)
				return
			}
		}
	}

	// Set timestamp
	newItem.Timestamp = time.Now()

	// Add item to the in-memory data store
	items = append(items, newItem)

	// Respond with the newly added item
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newItem)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getItems(w http.ResponseWriter, r *http.Request) {

	// Set header of the application response to json
	w.Header().Set("Content-Type", "application/json")
	// Check if there are items
	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	format, sort, err := parseQuery(r.URL.RawQuery)

	// Sort by Timestamp
	if sort == "timestamp" {
		stdsort.Slice(items, func(i, j int) bool {
			return items[i].Timestamp.Before(items[j].Timestamp)
		})
		// Sort by ID default
	} else {
		stdsort.Slice(items, func(i, j int) bool {
			return items[i].ID > items[j].ID
		})
	}

	if format == "yaml" {
		// Convert the output to yaml format
		err = yaml.NewEncoder(w).Encode(items)
	} else {
		// Convert the output to json format
		err = json.NewEncoder(w).Encode(items)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func parseQuery(rawQuery string) (string, string, error) {
	// Split the query string into individual parameters
	params := strings.Split(rawQuery, "&")

	var format, sort string

	for _, param := range params {
		// Split the parameter into key-value pair
		pair := strings.Split(param, "=")
		if len(pair) != 2 {
			return "", "", errors.New("invalid query parameter format")
		}

		// Extract the key and value from the parameter
		key := pair[0]
		value := pair[1]

		// Check the key to determine the type of parameter
		if key == "format" {
			// If the key is "format", set the format variable to the corresponding value
			format = value
		} else if key == "sort" {
			// If the key is "sort", set the sort variable to the corresponding value
			sort = value
		} else {
			return "", "", errors.New("invalid query parameter key") // bunu nerede görüuoruz?
		}
	}

	return format, sort, nil
}
