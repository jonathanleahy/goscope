package extract

import (
	"context"
	"fmt"

	"github.com/extract-scope-go/go-scope/internal/extract/format"
	"github.com/extract-scope-go/go-scope/internal/types"
)

// ExtractAndFormat extracts a symbol and formats the output
// This is the main public API that avoids circular imports
func ExtractAndFormat(ctx context.Context, target types.Target, opts types.Options) (*types.Result, error) {
	// Step 1: Extract symbol (returns unformatted result)
	result, err := ExtractSymbol(ctx, target, opts)
	if err != nil {
		return nil, err
	}

	// Step 2: Format based on requested format
	switch opts.Format {
	case "markdown", "":
		result.Rendered, err = format.ToMarkdown(result.Extract, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to format markdown: %w", err)
		}
	case "json":
		result.Rendered, err = format.ToJSON(result.Extract, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to format json: %w", err)
		}
	case "html":
		// TODO: Implement HTML formatter
		result.Rendered = "HTML formatting not yet implemented"
	default:
		result.Rendered = fmt.Sprintf("Unknown format: %s", opts.Format)
	}

	return result, nil
}
