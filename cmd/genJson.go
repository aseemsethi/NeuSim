package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
)

// generateJson creates a layered graph based on node counts per layer
func GenerateJson(layerNodeCounts []int) (string, error) {
	var nodes []Node
	var links []Link

	fmt.Printf("GenerateJson: called")
	// Track node IDs per layer
	layerNodes := make([][]string, len(layerNodeCounts))

	// ---------- CREATE NODES ----------
	for layerIndex, count := range layerNodeCounts {
		layer := layerIndex + 1

		for i := 1; i <= count; i++ {
			nodeID := fmt.Sprintf("L%d-N%d", layer, i)

			nodes = append(nodes, Node{
				ID:      nodeID,
				Group:   layer,
				Layer:   layer,
				Value:   0.0,
				ActivFn: "relu",
			})

			layerNodes[layerIndex] = append(layerNodes[layerIndex], nodeID)
		}
	}

	// ---------- CREATE LINKS WITH RANDOM WEIGHTS ----------
	for i := 0; i < len(layerNodes)-1; i++ {
		for _, src := range layerNodes[i] {
			for _, tgt := range layerNodes[i+1] {

				// Random weight between 0.1 and 0.5
				weight := 0.1 + rand.Float64()*0.4

				links = append(links, Link{
					Source: src,
					Target: tgt,
					Weight: round(weight, 2), // optional rounding
				})
			}
		}
	}

	graph := GraphData{
		Nodes: nodes,
		Links: links,
	}

	// ---------- SERIALIZE ----------
	bytes, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func round(val float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(val*p) / p
}
