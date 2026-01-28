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
		log.Printf("loadGraphData: error reading file: %v", err)
		return nil, err
	}

	//var data GraphData
	if err := json.Unmarshal(file, &data); err != nil {
		log.Printf("loadGraphData: error unmarshaling file: %v", err)
		return nil, err
	}

	log.Printf("loadGraphData: loaded data: %v", data)
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

// ---------- API: UPDATE SINGLE NODE ----------
func updateNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedNode Node
	if err := json.NewDecoder(r.Body).Decode(&updatedNode); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("updateNodeHandler: Updated Node data %v", updatedNode)

	if updatedNode.ID == "" {
		http.Error(w, "Node ID required", http.StatusBadRequest)
		return
	}

	found := false
	for i, node := range data.Nodes {
		if node.ID == updatedNode.ID {
			data.Nodes[i].Group = updatedNode.Group
			data.Nodes[i].Layer = updatedNode.Layer
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	if err := saveGraphToFile(data); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	log.Printf("updateNodeHandler: Writing to file")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Node updated"))
}

type AddNodeRequest struct {
	Node Node `json:"node"`
	Link Link `json:"link"`
}

/*
	 newNode:
		{
		  "node": { "id": "Node_123", "group": 1, "layer": 3 },
		  "link": { "source": "Node_A", "target": "Node_123", "weight": 1 }
		}
*/
func addNodeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("addNodeHandler: called")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate node
	if req.Node.ID == "" {
		http.Error(w, "Node ID required", http.StatusBadRequest)
		return
	}

	// Ensure node does not already exist
	for _, n := range data.Nodes {
		if n.ID == req.Node.ID {
			http.Error(w, "Node already exists", http.StatusConflict)
			return
		}
	}

	// Validate link endpoints
	sourceExists := false
	for _, n := range data.Nodes {
		if n.ID == req.Link.Source {
			sourceExists = true
			break
		}
	}

	if !sourceExists {
		http.Error(w, "Source node does not exist", http.StatusBadRequest)
		return
	}

	log.Printf("addNodeHandler: new node %v", req)

	// Append new node and link
	data.Nodes = append(data.Nodes, req.Node)
	data.Links = append(data.Links, req.Link)

	if err := saveGraphToFile(data); err != nil {
		http.Error(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}

	log.Printf("addNodeHandler: write new node to file %v", req)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Node added"))
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
/*
Sample Data Returned:
{"nodes":[{"id":"Node A","group":1,"layer":1},
{"id":"Node B","group":2,"layer":1},{"id":"Node C","group":1,"layer":2},
{"id":"Node D","group":3,"layer":2},{"id":"Node E","group":3,"layer":3},
{"id":"Node F","group":3,"layer":3}],
"links":[{"source":"Node A","target":"Node C","weight":5},
{"source":"Node A","target":"Node D","weight":3},
{"source":"Node B","target":"Node C","weight":5},
{"source":"Node B","target":"Node D","weight":3},
{"source":"Node C","target":"Node E","weight":5},
{"source":"Node C","target":"Node F","weight":3},
{"source":"Node D","target":"Node F","weight":2}]}
*/
func getGraphHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(data)
	log.Printf("getGraphHandler: returning data: %v", data)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("getGraphHandler: error marshaling data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonBytes)
}

// ---------- API: SAVE GRAPH ----------
func saveGraphHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("saveGraphHandler: called")
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
		log.Printf("saveGraphHandler: invalid JSON err: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("saveGraphHandler: valid JSON: %v", incoming)

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
		log.Printf("Main: error loading GraphData %v", err)
		panic(err)
	}

	fmt.Printf("Loaded %d nodes and %d links\n",
		len(data.Nodes),
		len(data.Links),
	)

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/getGraph", getGraphHandler)
	//http.HandleFunc("/api/save", saveGraphHandler)
	http.HandleFunc("/api/node", updateNodeHandler)
	http.HandleFunc("/api/node/add", addNodeHandler)
	http.HandleFunc("/hello", helloHandler)

	port := ":8080"
	log.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
