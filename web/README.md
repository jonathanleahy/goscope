# Go Scope Visualizer

Interactive web-based dependency graph visualizer for Go Scope extracts.

## Features

✨ **Interactive Graph Visualization**
- Force-directed graph layout using D3.js
- Drag nodes to rearrange
- Zoom and pan controls
- Expandable/collapsible views

🎨 **Visual Clarity**
- Color-coded nodes (target, internal, external)
- Clear dependency edges
- Hoverable tooltips
- Toggle labels and external symbols

📝 **Code Display**
- Click any node to view its code
- Documentation display
- Syntax highlighting
- File location and metadata

📊 **Statistics**
- Total nodes and edges
- Maximum depth
- External reference count

## Quick Start

### 1. Generate JSON Extract

First, extract your Go code as JSON:

```bash
cd /your/go/project
go-scope -file=path/to/file.go -line=42 -depth=2 -format=json -output=extract.json
```

### 2. Open Visualizer

#### Option A: Simple HTTP Server (Python)

```bash
cd web/public
python3 -m http.server 8080
```

Then open: http://localhost:8080

#### Option B: Simple HTTP Server (Node.js)

```bash
npx http-server web/public -p 8080
```

#### Option C: Simple HTTP Server (Go)

```bash
cd web/public
go run ../../cmd/serve/main.go
```

### 3. Load Your Extract

1. Click "📁 Load Extract JSON"
2. Select your `extract.json` file
3. Explore the interactive graph!

## Usage

### Controls

- **📁 Load Extract JSON** - Load a JSON extract file
- **➕ Zoom In** - Zoom into the graph
- **➖ Zoom Out** - Zoom out of the graph
- **🎯 Reset View** - Reset zoom and pan
- **🏷️ Toggle Labels** - Show/hide node labels
- **Show External** - Toggle external symbols
- **Show Docs** - Toggle documentation display

### Interaction

- **Click node** - View code and details in right panel
- **Drag node** - Reposition node (and connected nodes will follow)
- **Hover node** - See tooltip with basic info
- **Mouse wheel** - Zoom in/out
- **Click & drag background** - Pan the view

### Legend

- 🔴 **Red nodes** - Target symbol (what you extracted)
- 🔵 **Blue nodes** - Internal symbols (from your codebase)
- ⚪ **Gray nodes** - External symbols (from other packages)
- **Lines** - Dependency relationships

## Architecture

```
web/
├── public/
│   ├── index.html      # Main HTML page
│   ├── styles.css      # Styling
│   └── app.js          # Visualization logic (D3.js)
└── README.md           # This file
```

### Technologies

- **D3.js v7** - Force-directed graph visualization
- **Vanilla JavaScript** - No build step required
- **CSS3** - Modern responsive design
- **HTML5** - Semantic markup

## Examples

### Visualize a Function

```bash
# Extract Add function with dependencies
cd examples/ex1
../../bin/go-scope -file=pkg/math/add.go -line=7 -depth=2 -format=json -output=add-extract.json

# Open visualizer and load add-extract.json
python3 -m http.server 8080 -d ../../web/public
```

### Visualize Deep Dependencies

```bash
# Extract with depth 3 (deep dependency tree)
go-scope -file=main.go -line=10 -depth=3 -format=json -output=deep.json
```

## Customization

### Modify Appearance

Edit `styles.css` to change:
- Colors (`:root` CSS variables)
- Node sizes
- Font styles
- Layout spacing

### Modify Behavior

Edit `app.js` to change:
- Force simulation parameters (in `renderGraph()`)
- Node radius and spacing
- Zoom limits
- Drag behavior

## Troubleshooting

**Q: Graph is too crowded**
- Reduce depth when generating JSON (`-depth=1` instead of `-depth=2`)
- Toggle off external symbols
- Toggle off labels
- Zoom in to focus on a section

**Q: Nodes overlap**
- Drag nodes to reposition
- Increase collision force in `app.js`
- Try a different extract depth

**Q: Can't load JSON**
- Ensure JSON is valid (test with `python3 -m json.tool extract.json`)
- Check browser console for errors
- Verify JSON was generated with `-format=json`

**Q: External nodes cluttering view**
- Uncheck "Show External" checkbox
- Reduce depth when extracting
- External packages often have many dependencies

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

Requires modern JavaScript (ES6+) and SVG support.

## Performance

- **Small graphs** (< 50 nodes): Instant, smooth
- **Medium graphs** (50-200 nodes): Very good
- **Large graphs** (200-500 nodes): Good, may need to reduce depth
- **Very large graphs** (500+ nodes): Consider filtering external symbols

## Future Enhancements

- [ ] Export graph as PNG/SVG
- [ ] Minimap for large graphs
- [ ] Search/filter nodes
- [ ] Path highlighting between nodes
- [ ] Different layout algorithms
- [ ] Animation controls
- [ ] Dark mode
- [ ] Comparison mode (diff two extracts)

## License

Same as go-scope project (MIT)
