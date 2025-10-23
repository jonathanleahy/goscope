package format

import (
	"encoding/json"
	"testing"

	"github.com/extract-scope-go/go-scope/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToJSON tests basic JSON output
func TestToJSON(t *testing.T) {
	// Given: A simple extract with target and one dependency
	ext := types.Extract{
		Target: types.Symbol{
			Name:    "Add",
			Kind:    "func",
			Package: "example.com/test",
			File:    "/path/to/add.go",
			Line:    10,
			Code:    "func Add(a, b int) int { return a + b }",
			Doc:     "Add returns sum",
		},
		References: []types.Reference{
			{
				Symbol: types.Symbol{
					Name:    "Helper",
					Kind:    "func",
					Package: "example.com/test",
					File:    "/path/to/helper.go",
					Line:    5,
					Code:    "func Helper() {}",
				},
				Depth:        1,
				ReferencedBy: "Add",
				Reason:       "direct-call",
			},
		},
		External: []string{"fmt.Println"},
	}

	// When: We convert to JSON
	result, err := ToJSON(ext, types.Options{Depth: 1})

	// Then: Should produce valid JSON
	require.NoError(t, err)
	assert.Contains(t, result, "\"name\": \"Add\"")
	assert.Contains(t, result, "\"name\": \"Helper\"")
	assert.Contains(t, result, "fmt.Println")

	// Should be parseable as JSON
	var viz VisualizationData
	err = json.Unmarshal([]byte(result), &viz)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "Add", viz.Target.Name)
	assert.Equal(t, 1, len(viz.Nodes))
	assert.Equal(t, 1, len(viz.Edges))
	assert.Equal(t, 1, len(viz.External))
}

// TestJSONStructure tests the JSON structure
func TestJSONStructure(t *testing.T) {
	// Given: Extract with multiple dependencies
	ext := types.Extract{
		Target: types.Symbol{
			Name: "Main",
			Kind: "func",
		},
		References: []types.Reference{
			{
				Symbol:       types.Symbol{Name: "Func1", Kind: "func"},
				Depth:        1,
				ReferencedBy: "Main",
			},
			{
				Symbol:       types.Symbol{Name: "Func2", Kind: "func"},
				Depth:        2,
				ReferencedBy: "Func1",
			},
		},
	}

	// When: We convert to JSON
	result, err := ToJSON(ext, types.Options{Depth: 2})
	require.NoError(t, err)

	var viz VisualizationData
	err = json.Unmarshal([]byte(result), &viz)
	require.NoError(t, err)

	// Then: Should have correct structure
	assert.Equal(t, "Main", viz.Target.Name)
	assert.True(t, viz.Target.IsTarget)
	assert.Equal(t, 0, viz.Target.Depth)
	assert.Equal(t, 2, viz.TotalLayers)
	assert.Equal(t, 2, len(viz.Nodes))
	assert.Equal(t, 2, len(viz.Edges))
}

// TestJSONWithMetrics tests JSON with metrics
func TestJSONWithMetrics(t *testing.T) {
	// Given: Extract with metrics
	ext := types.Extract{
		Target: types.Symbol{Name: "Test"},
		Metrics: &types.Metrics{
			LinesOfCode:          50,
			CyclomaticComplexity: 5,
			DependencyCount:      3,
		},
	}

	// When: We convert to JSON
	result, err := ToJSON(ext, types.Options{})
	require.NoError(t, err)

	var viz VisualizationData
	err = json.Unmarshal([]byte(result), &viz)
	require.NoError(t, err)

	// Then: Should include metrics
	require.NotNil(t, viz.Metrics)
	assert.Equal(t, 50, viz.Metrics.LinesOfCode)
	assert.Equal(t, 5, viz.Metrics.CyclomaticComplexity)
	assert.Equal(t, 3, viz.Metrics.DependencyCount)
}

// TestJSONEdges tests edge generation
func TestJSONEdges(t *testing.T) {
	// Given: Extract with clear dependency chain
	ext := types.Extract{
		Target: types.Symbol{Name: "A"},
		References: []types.Reference{
			{
				Symbol:       types.Symbol{Name: "B"},
				ReferencedBy: "A",
				Reason:       "direct-call",
				Depth:        1,
			},
			{
				Symbol:       types.Symbol{Name: "C"},
				ReferencedBy: "B",
				Reason:       "type-reference",
				Depth:        2,
			},
		},
	}

	// When: We convert to JSON
	result, err := ToJSON(ext, types.Options{})
	require.NoError(t, err)

	var viz VisualizationData
	err = json.Unmarshal([]byte(result), &viz)
	require.NoError(t, err)

	// Then: Should have correct edges
	assert.Equal(t, 2, len(viz.Edges))

	// Edge A -> B
	assert.Equal(t, "A", viz.Edges[0].From)
	assert.Equal(t, "B", viz.Edges[0].To)
	assert.Equal(t, "direct-call", viz.Edges[0].Type)

	// Edge B -> C
	assert.Equal(t, "B", viz.Edges[1].From)
	assert.Equal(t, "C", viz.Edges[1].To)
	assert.Equal(t, "type-reference", viz.Edges[1].Type)
}

// TestJSONExternalSymbols tests external symbol handling
func TestJSONExternalSymbols(t *testing.T) {
	// Given: Extract with external dependencies
	ext := types.Extract{
		Target: types.Symbol{Name: "Test"},
		References: []types.Reference{
			{
				Symbol:   types.Symbol{Name: "Printf", Package: "fmt"},
				External: true,
				Stub:     true,
				Depth:    1,
			},
		},
		External: []string{"fmt.Printf", "os.Exit"},
	}

	// When: We convert to JSON
	result, err := ToJSON(ext, types.Options{})
	require.NoError(t, err)

	var viz VisualizationData
	err = json.Unmarshal([]byte(result), &viz)
	require.NoError(t, err)

	// Then: Should mark external nodes
	assert.True(t, viz.Nodes[0].External)
	assert.True(t, viz.Nodes[0].Stub)
	assert.Equal(t, 2, len(viz.External))
}
