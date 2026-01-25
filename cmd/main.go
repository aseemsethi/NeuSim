package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

// DataPoint represents a single data point for the D3 graph
type DataPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// Data is the structure that will be returned as JSON
type Data struct {
	Points []DataPoint `json:"points"`
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	// Sample data to send to the frontend
	data := Data{
		Points: []DataPoint{
			{Label: "A", Value: 10},
			{Label: "B", Value: 20},
			{Label: "C", Value: 30},
			{Label: "D", Value: 40},
			{Label: "E", Value: 50},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {

	data := GraphData{
		Nodes: []Node{
			{ID: "Node A", Group: 1},
			{ID: "Node B", Group: 2},
			{ID: "Node C", Group: 1},
		},
		Links: []Link{
			{Source: "Node A", Target: "Node B", Value: 1},
			{Source: "Node B", Target: "Node C", Value: 2},
			{Source: "Node C", Target: "Node A", Value: 3},
		},
	}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/api/data", dataHandler)

	http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonBytes)
		//json.NewEncoder(w).Encode(data). - this is another way to do it
	})

	port := ":8080"
	log.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
