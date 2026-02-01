package main

// Node structure (golang)
type Node struct {
	ID      string  `json:"id"`
	Group   int     `json:"group"`
	Layer   int     `json:"layer"`
	Value   float64 `json:"value"`
	ActivFn string  `json:"activFn"`
}

// Link structure (golang), source and target will be IDs
type Link struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Weight float64 `json:"weight"`
}

// GraphData holds both nodes and links
type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}
