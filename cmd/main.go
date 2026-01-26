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
			{ID: "Node A", Group: 1, Layer: 1},
			{ID: "Node B", Group: 2, Layer: 1},
			{ID: "Node C", Group: 1, Layer: 2},
			{ID: "Node D", Group: 2, Layer: 2},
			{ID: "Node E", Group: 2, Layer: 3},
			{ID: "Node F", Group: 2, Layer: 3},
		},
		Links: []Link{
			{Source: "Node A", Target: "Node C", Value: 5},
			{Source: "Node A", Target: "Node D", Value: 20},
			{Source: "Node B", Target: "Node C", Value: 40},
			{Source: "Node B", Target: "Node D", Value: 40},
			{Source: "Node C", Target: "Node E", Value: 40},
			{Source: "Node C", Target: "Node F", Value: 40},
			{Source: "Node D", Target: "Node F", Value: 40},
		},
	}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/hello", helloHandler)

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
