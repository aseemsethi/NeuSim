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
	http.HandleFunc("/graph-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonBytes)
		//json.NewEncoder(w).Encode(data). - this is another way to do it
	})

	http.HandleFunc("/", helloHandler)
	port := ":8080"
	log.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
