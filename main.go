package main

import (
	"encoding/json"
	"errors"
	"flag"
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
	// Parse command-line arguments
	var port int
	flag.IntVar(&port, "port", 8090, "port number for HTTP server")
	flag.Parse()

	http.HandleFunc("/items", itemsHandler)
	http.HandleFunc("/items/", itemsHandlerByID)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on port %d...\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if err := getItems(w, r); err != nil {
			log.Printf("error getting items: %v", err)
		}
	case "POST":
		if err := addItem(w, r); err != nil {
			log.Printf("error adding item: %v", err)
		}

	default:
		http.Error(w, fmt.Sprintf("cannot do an HTTP %q request on the endpoint %s. Please review your request and retry.", r.Method, r.URL.Path), http.StatusMethodNotAllowed)
	}
}

func itemsHandlerByID(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from path
	id := r.URL.Path[len("/items/"):]

	// If id is empty, call itemsHandler to list all items
	if id == "" {
		itemsHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		if err := getItemsByID(w, r, id); err != nil {
			log.Printf("error getting item by ID: %v", err)
		}
	case "DELETE":
		deleteItemsByID(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getItemsByID(w http.ResponseWriter, r *http.Request, id string) error {
	// Find item by ID
	_, foundItem, found := searchID(items, id)

	// Check if an item exists with the specified ID
	if !found {
		http.Error(w, fmt.Sprintf("there is no item with the ID %q", id), http.StatusNotFound)
		return nil
	}

	params, err := parseQuery(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse query: %s", err), http.StatusBadRequest)
		return err
	}

	if params.Format == "yaml" {
		err = yaml.NewEncoder(w).Encode(foundItem)
	} else {
		err = json.NewEncoder(w).Encode(foundItem)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func deleteItemsByID(w http.ResponseWriter, r *http.Request, id string) {
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

func addItem(w http.ResponseWriter, r *http.Request) error {
	var newItem item
	// Decode the JSON data from the request body into the newItem variable
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Check if name is provided
	if newItem.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return nil
	}

	// Generate ID if not provided
	if newItem.ID == "" {
		newItem.ID = uuid.NewString()
	} else {
		// Check if provided ID is duplicate
		_, _, found := searchID(items, newItem.ID)
		if found {
			http.Error(w, "specified ID already exists", http.StatusConflict)
			return nil
		}
	}

	newItem.Timestamp = time.Now()

	// Add item to the in-memory data store
	items = append(items, newItem)

	// Respond with the newly added item
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newItem); err != nil {
		log.Println("error encoding JSON:", err)
		return err
	}
	return nil
}

func getItems(w http.ResponseWriter, r *http.Request) error {
	// Set header of the application response to json
	w.Header().Set("Content-Type", "application/json")
	// Check if there are items
	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	params, err := parseQuery(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse query: %s", err), http.StatusBadRequest)
		return err
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
		return err
	}
	return nil
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
