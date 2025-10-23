package format

import (
	"encoding/json"

	"github.com/extract-scope-go/go-scope/internal/types"
)

// ToJSON converts an Extract to JSON format for the web visualizer
func ToJSON(ext types.Extract, opts types.Options) (string, error) {
	// Create a visualization-friendly structure
	viz := VisualizationData{
		Target:      convertSymbolToNode(ext.Target, 0, true),
		Nodes:       []Node{},
		Edges:       []Edge{},
		External:    ext.External,
		Options:     opts,
		TotalLayers: calculateMaxDepth(ext.References),
	}

	// Add target node
	nodeMap := make(map[string]bool)
	nodeMap[makeNodeID(ext.Target)] = true

	// Convert references to nodes and edges
	for _, ref := range ext.References {
		nodeID := makeNodeID(ref.Symbol)

		// Add node if not already added
		if !nodeMap[nodeID] {
			node := convertSymbolToNode(ref.Symbol, ref.Depth, false)
			node.External = ref.External
			node.Stub = ref.Stub
			viz.Nodes = append(viz.Nodes, node)
			nodeMap[nodeID] = true
		}

		// Add edge from referenced-by to this symbol
		if ref.ReferencedBy != "" {
			edge := Edge{
				From:   ref.ReferencedBy,
				To:     ref.Symbol.Name,
				Type:   ref.Reason,
				Depth:  ref.Depth,
				Label:  ref.Reason,
			}
			viz.Edges = append(viz.Edges, edge)
		}
	}

	// Add metrics if available
	if ext.Metrics != nil {
		viz.Metrics = &MetricsData{
			LinesOfCode:          ext.Metrics.LinesOfCode,
			CyclomaticComplexity: ext.Metrics.CyclomaticComplexity,
			DependencyCount:      ext.Metrics.DependencyCount,
			DirectDeps:           ext.Metrics.DirectDeps,
			TransitiveDeps:       ext.Metrics.TransitiveDeps,
			ExternalPackages:     ext.Metrics.ExternalPackages,
		}
	}

	// Add interface mappings
	if len(ext.InterfaceMappings) > 0 {
		for _, mapping := range ext.InterfaceMappings {
			mappingData := InterfaceMappingData{
				Interface:       convertSymbolToNode(mapping.Interface, 0, false),
				Implementations: []Node{},
				DIFramework:     mapping.DIFramework,
			}

			for _, impl := range mapping.Implementations {
				implNode := convertSymbolToNode(impl, 0, false)
				mappingData.Implementations = append(mappingData.Implementations, implNode)
			}

			if mapping.Constructor != nil {
				constructorNode := convertSymbolToNode(*mapping.Constructor, 0, false)
				mappingData.Constructor = &constructorNode
			}

			viz.InterfaceMappings = append(viz.InterfaceMappings, mappingData)
		}
	}

	// Add DI bindings
	if len(ext.DIBindings) > 0 {
		for _, binding := range ext.DIBindings {
			bindingData := DIBindingData{
				Provider:     convertSymbolToNode(binding.Provider, 0, false),
				Product:      convertSymbolToNode(binding.Product, 0, false),
				Dependencies: []Node{},
				Framework:    binding.Framework,
				Scope:        binding.Scope,
			}

			for _, dep := range binding.Dependencies {
				depNode := convertSymbolToNode(dep, 0, false)
				bindingData.Dependencies = append(bindingData.Dependencies, depNode)
			}

			viz.DIBindings = append(viz.DIBindings, bindingData)
		}
	}

	// Add detected DI framework
	viz.DetectedDIFramework = ext.DetectedDIFramework

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(viz, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// VisualizationData is the JSON structure for the web visualizer
type VisualizationData struct {
	Target              Node                    `json:"target"`
	Nodes               []Node                  `json:"nodes"`
	Edges               []Edge                  `json:"edges"`
	External            []string                `json:"external,omitempty"`
	Metrics             *MetricsData            `json:"metrics,omitempty"`
	Options             types.Options           `json:"options"`
	TotalLayers         int                     `json:"totalLayers"`
	InterfaceMappings   []InterfaceMappingData  `json:"interfaceMappings,omitempty"`
	DIBindings          []DIBindingData         `json:"diBindings,omitempty"`
	DetectedDIFramework string                  `json:"detectedDIFramework,omitempty"`
}

// Node represents a symbol node in the visualization
type Node struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Package  string `json:"package"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	EndLine  int    `json:"endLine"`
	Code     string `json:"code"`
	Doc      string `json:"doc,omitempty"`
	Exported bool   `json:"exported"`
	Depth    int    `json:"depth"`
	IsTarget bool   `json:"isTarget"`
	External bool   `json:"external"`
	Stub     bool   `json:"stub"`
}

// Edge represents a dependency relationship
type Edge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Type  string `json:"type"`
	Depth int    `json:"depth"`
	Label string `json:"label"`
}

// MetricsData holds code metrics
type MetricsData struct {
	LinesOfCode          int      `json:"linesOfCode"`
	CyclomaticComplexity int      `json:"cyclomaticComplexity"`
	DependencyCount      int      `json:"dependencyCount"`
	DirectDeps           int      `json:"directDeps"`
	TransitiveDeps       int      `json:"transitiveDeps"`
	ExternalPackages     []string `json:"externalPackages"`
}

// InterfaceMappingData holds interface-to-implementation mapping
type InterfaceMappingData struct {
	Interface       Node    `json:"interface"`
	Implementations []Node  `json:"implementations"`
	Constructor     *Node   `json:"constructor,omitempty"`
	DIFramework     string  `json:"diFramework,omitempty"`
}

// DIBindingData holds dependency injection binding information
type DIBindingData struct {
	Provider     Node     `json:"provider"`
	Product      Node     `json:"product"`
	Dependencies []Node   `json:"dependencies"`
	Framework    string   `json:"framework"`
	Scope        string   `json:"scope"`
}

// convertSymbolToNode converts a Symbol to a visualization Node
func convertSymbolToNode(sym types.Symbol, depth int, isTarget bool) Node {
	return Node{
		ID:       makeNodeID(sym),
		Name:     sym.Name,
		Kind:     sym.Kind,
		Package:  sym.Package,
		File:     sym.File,
		Line:     sym.Line,
		EndLine:  sym.EndLine,
		Code:     sym.Code,
		Doc:      sym.Doc,
		Exported: sym.Exported,
		Depth:    depth,
		IsTarget: isTarget,
	}
}

// makeNodeID creates a unique ID for a symbol
func makeNodeID(sym types.Symbol) string {
	if sym.Package != "" {
		return sym.Package + "." + sym.Name
	}
	return sym.Name
}

// calculateMaxDepth finds the maximum depth in references
func calculateMaxDepth(refs []types.Reference) int {
	maxDepth := 0
	for _, ref := range refs {
		if ref.Depth > maxDepth {
			maxDepth = ref.Depth
		}
	}
	return maxDepth
}
