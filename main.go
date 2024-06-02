package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	stdsort "sort"
	"time"

	"github.com/google/uuid"
	yaml "gopkg.in/yaml.v3"
)

const (
	parameterKeySort   = "sort"
	parameterKeyFormat = "format"
)

type QueryParams struct {
	Format string
	Sort   string
}

type item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

var items []item

func main() {
	http.HandleFunc("/items", itemsHandler)
	http.HandleFunc("/items/", itemsHandlerByID)

	if err := http.ListenAndServe(":8090", nil); err != nil {
		log.Fatal(err)
	}
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getItems(w, r)
	case "POST":
		addItem(w, r)
	default:
		http.Error(w, fmt.Sprintf("cannot do an HTTP %q request on the endpoint %s. Please review your request and retry.", r.Method, r.URL.Path), http.StatusMethodNotAllowed)
	}
}

func itemsHandlerByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getItemsByID(w, r)
	} else if r.Method == "DELETE" {
		deleteItemsByID(w, r)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getItemsByID(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from path
	id := r.URL.Path[len("/items/"):]

	// Find item by ID
	_, foundItem, found := searchID(items, id)

	// Check if an item exists with the specified ID
	if !found {
		http.Error(w, fmt.Sprintf("there is no item with the ID %q", id), http.StatusNotFound)
		return
	}

	params, err := parseQuery(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse query: %s", err), http.StatusBadRequest)
		return
	}

	if params.Format == "yaml" {
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
	index, _, found := searchID(items, id)

	// Check if an item exists with the specified ID
	if !found {
		http.Error(w, "there is no item with the specified ID", http.StatusNotFound)
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
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if newItem.ID == "" {
		newItem.ID = uuid.NewString()
	} else {
		// Check if provided ID is duplicate
		_, _, found := searchID(items, newItem.ID)
		if found {
			http.Error(w, "specified ID already exists", http.StatusConflict)
			return
		}
	}

	newItem.Timestamp = time.Now()

	// Add item to the in-memory data store
	items = append(items, newItem)

	// Respond with the newly added item
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newItem); err != nil {
		log.Println("error encoding JSON:", err)
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

	params, err := parseQuery(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse query: %s", err), http.StatusBadRequest)
		return
	}

	if params.Sort == "timestamp" { // Sort by Timestamp
		stdsort.Slice(items, func(i, j int) bool {
			return items[i].Timestamp.Before(items[j].Timestamp)
		})
	} else { // Sort by ID default
		stdsort.Slice(items, func(i, j int) bool {
			return items[i].ID > items[j].ID
		})
	}

	if params.Format == "yaml" {
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

func parseQuery(r *http.Request) (QueryParams, error) {
	// Get the query parameters from the URL
	queryParams := r.URL.Query()

	// Extract the 'format' and 'sort' parameters
	format := queryParams.Get(parameterKeyFormat)
	sort := queryParams.Get(parameterKeySort)

	// Check the validity of the 'format' parameter
	if format != "" && format != "json" && format != "yaml" {
		return QueryParams{}, errors.New("invalid value for 'format' parameter")
	}
	// Check the validity of the 'sort' parameter
	if sort != "" && sort != "id" && sort != "timestamp" {
		return QueryParams{}, errors.New("invalid value for 'sort' parameter")
	}

	return QueryParams{
		Format: format,
		Sort:   sort,
	}, nil
}

// searchID finds an item by its ID in the items slice and returns its index, the item, and a boolean indicating if it was found
func searchID(items []item, id string) (int, item, bool) {
	for index, item := range items {
		if item.ID == id {
			return index, item, true
		}
	}
	return -1, item{}, false
}
