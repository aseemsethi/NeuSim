package main

import (
	"encoding/json"
	"fmt"
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
		fmt.Printf("loadGraphData: error reading file: %v", err)
		return nil, err
	}

	//var data GraphData
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Printf("loadGraphData: error unmarshaling file: %v", err)
		return nil, err
	}

	fmt.Printf("loadGraphData: loaded data: %v\n", data)
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedNode Node
	if err := json.NewDecoder(r.Body).Decode(&updatedNode); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("updateNodeHandler: Updated Node data %v\n", updatedNode)

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
	fmt.Printf("updateNodeHandler: Writing to file")

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
	fmt.Println("addNodeHandler: called")
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

	fmt.Printf("addNodeHandler: new node %v\n", req)

	// Append new node and link
	data.Nodes = append(data.Nodes, req.Node)
	data.Links = append(data.Links, req.Link)

	if err := saveGraphToFile(data); err != nil {
		http.Error(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}

	fmt.Printf("addNodeHandler: write new node to file %v", req)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Node added"))
}

// ---------- API: ADD LINK ----------
func addLinkHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("addLinkHandler: called\n")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newLink Link
	if err := json.NewDecoder(r.Body).Decode(&newLink); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if newLink.Source == "" || newLink.Target == "" {
		http.Error(w, "Source and target are required", http.StatusBadRequest)
		return
	}

	if newLink.Source == newLink.Target {
		http.Error(w, "Source and target cannot be the same", http.StatusBadRequest)
		return
	}

	// Check that source and target nodes exist
	sourceExists := false
	targetExists := false

	for _, n := range data.Nodes {
		if n.ID == newLink.Source {
			sourceExists = true
		}
		if n.ID == newLink.Target {
			targetExists = true
		}
	}

	if !sourceExists || !targetExists {
		http.Error(w, "Source or target node does not exist", http.StatusBadRequest)
		return
	}

	// Prevent duplicate links (optional but recommended)
	for _, l := range data.Links {
		if l.Source == newLink.Source && l.Target == newLink.Target {
			http.Error(w, "Link already exists", http.StatusConflict)
			return
		}
	}

	// Default weight if not provided
	if newLink.Weight == 0 {
		newLink.Weight = 1
	}

	// Append link
	data.Links = append(data.Links, newLink)
	fmt.Printf("addLinkHandler: newLink %v", newLink)

	// Persist to file
	if err := saveGraphToFile(data); err != nil {
		http.Error(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Link added successfully"))
}

// ---------- API: UPDATE LINK ----------
func updateLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedLink Link
	if err := json.NewDecoder(r.Body).Decode(&updatedLink); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation
	if updatedLink.Source == "" || updatedLink.Target == "" {
		http.Error(w, "Source and target are required", http.StatusBadRequest)
		return
	}

	if updatedLink.Weight <= 0 {
		http.Error(w, "Weight must be greater than zero", http.StatusBadRequest)
		return
	}

	// Find and update the link
	found := false
	for i, link := range data.Links {
		if link.Source == updatedLink.Source && link.Target == updatedLink.Target {
			data.Links[i].Weight = updatedLink.Weight
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	fmt.Printf("updateLinkHandler: updated link %v", updatedLink)

	// Persist to file
	if err := saveGraphToFile(data); err != nil {
		http.Error(w, "Failed to save graph", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Link updated successfully"))
}

// ---------- VALIDATE GRAPH ----------
// func validateGraph(data GraphData) error {
// 	nodeIDs := make(map[string]bool)

// 	for _, n := range data.Nodes {
// 		if n.ID == "" {
// 			return fmt.Errorf("node id cannot be empty")
// 		}
// 		nodeIDs[n.ID] = true
// 	}

// 	for _, l := range data.Links {
// 		if !nodeIDs[l.Source] || !nodeIDs[l.Target] {
// 			return fmt.Errorf("invalid link: %s -> %s", l.Source, l.Target)
// 		}
// 	}

// 	return nil
// }

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
	fmt.Printf("getGraphHandler: returning data: %v", data)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("getGraphHandler: error marshaling data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonBytes)
}

func main() {
	data, err := loadGraphData(graphFile)
	if err != nil {
		fmt.Printf("Main: error loading GraphData %v", err)
		panic(err)
	}

	fmt.Printf("Loaded %d nodes and %d links\n",
		len(data.Nodes),
		len(data.Links),
	)

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/getGraph", getGraphHandler)
	http.HandleFunc("/api/link", updateLinkHandler)
	http.HandleFunc("/api/node", updateNodeHandler)
	http.HandleFunc("/api/node/add", addNodeHandler)
	http.HandleFunc("/api/link/add", addLinkHandler)
	http.HandleFunc("/hello", helloHandler)

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
