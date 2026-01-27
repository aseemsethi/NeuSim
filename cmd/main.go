package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var graphFile = "cmd/neuData.json"
var data GraphData

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

// ---------- LOAD GRAPH FROM FILE ----------
func loadGraphData(filename string) (*GraphData, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data GraphData
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// ---------- SAVE GRAPH TO FILE ----------
func saveGraphToFile(data GraphData) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(graphFile, bytes, 0644)
}

// ---------- VALIDATE GRAPH ----------
func validateGraph(data GraphData) error {
	nodeIDs := make(map[string]bool)

	for _, n := range data.Nodes {
		if n.ID == "" {
			return fmt.Errorf("node id cannot be empty")
		}
		nodeIDs[n.ID] = true
	}

	for _, l := range data.Links {
		if !nodeIDs[l.Source] || !nodeIDs[l.Target] {
			return fmt.Errorf("invalid link: %s -> %s", l.Source, l.Target)
		}
	}

	return nil
}

// ---------- API: GET GRAPH ----------
func getGraphHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(data)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonBytes)
}

// ---------- API: SAVE GRAPH ----------
func saveGraphHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	var incoming GraphData
	if err := json.Unmarshal(body, &incoming); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate data integrity
	if err := validateGraph(incoming); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save to disk
	if err := saveGraphToFile(incoming); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Update memory cache
	data = incoming

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Saved successfully"))
}

func main() {

	// data := GraphData{
	// 	Nodes: []Node{
	// 		{ID: "Node A", Group: 1, Layer: 1},
	// 		{ID: "Node B", Group: 2, Layer: 1},
	// 		{ID: "Node C", Group: 1, Layer: 2},
	// 		{ID: "Node D", Group: 2, Layer: 2},
	// 		{ID: "Node E", Group: 2, Layer: 3},
	// 		{ID: "Node F", Group: 2, Layer: 3},
	// 	},
	// 	Links: []Link{
	// 		{Source: "Node A", Target: "Node C", Weight: 5},
	// 		{Source: "Node A", Target: "Node D", Weight: 20},
	// 		{Source: "Node B", Target: "Node C", Weight: 40},
	// 		{Source: "Node B", Target: "Node D", Weight: 40},
	// 		{Source: "Node C", Target: "Node E", Weight: 40},
	// 		{Source: "Node C", Target: "Node F", Weight: 40},
	// 		{Source: "Node D", Target: "Node F", Weight: 40},
	// 	},
	// }

	data, err := loadGraphData(graphFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d nodes and %d links\n",
		len(data.Nodes),
		len(data.Links),
	)

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/getGraph", getGraphHandler)
	http.HandleFunc("/api/save", saveGraphHandler)
	http.HandleFunc("/hello", helloHandler)

	port := ":8080"
	log.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
