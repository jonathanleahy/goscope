# Phase 2: Interactive Visualizer

**Status**: Planned for after v1.0.0 core tool is complete
**Priority**: High - Major UX enhancement

---

## Goal

Create an **interactive web-based visualizer** that reads the JSON output from `go-scope` and displays it as an explorable, layered canvas where users can visually navigate through code dependencies.

---

## Concept

```
Start at target symbol (center)
  ↓
Click to expand depth 1 dependencies
  ↓
Click individual nodes to expand their dependencies
  ↓
Interactive exploration of the entire dependency tree
```

**Visual Metaphor**: Like a mind map or dependency graph that you can explore interactively

---

## Features

### Core Visualization
- **Central Node**: Target symbol (highlighted)
- **Depth Layers**: Visual rings/layers for depth 0, 1, 2, 3...
- **Connection Lines**: Show which symbol calls/uses which
- **Expandable Nodes**: Click to expand/collapse dependencies
- **Zoom & Pan**: Navigate large dependency trees
- **Minimap**: Overview of entire graph

### Interactive Features
- **Hover**: Show quick info (type, location, LOC)
- **Click Node**: Expand dependencies or show full code
- **Click Edge**: Show reference type (call, type-ref, field-access)
- **Search**: Find symbols in the visualization
- **Filter**: Show only certain types (functions, types, etc.)
- **Path Highlighting**: Highlight path from root to selected node

### Code Display
- **Split Panel**: Graph on left, code on right
- **Syntax Highlighting**: Beautiful code display
- **Line Annotations**: Show where references occur
- **Quick Nav**: Jump between symbols

### Export & Sharing
- **Save Layout**: Remember zoom/pan/expanded state
- **Export PNG/SVG**: Static images of the graph
- **Share Link**: Generate shareable visualization URL
- **Embed**: Iframe embed for documentation

---

## Technology Stack

### Frontend Framework
- **React** or **Vue** - Component-based UI
- **TypeScript** - Type safety

### Visualization Library Options

#### Option 1: D3.js ⭐ (Recommended)
**Pros**:
- Most flexible and powerful
- Great for custom graph layouts
- Force-directed graphs, tree layouts, radial layouts
- Full control over rendering

**Cons**:
- Steeper learning curve
- More code to write

**Example Libraries**:
- `d3.js` - Core library
- `react-d3-graph` - React wrapper
- `d3-force` - Force-directed layout

#### Option 2: Cytoscape.js
**Pros**:
- Specifically designed for graph visualization
- Built-in layouts (force, circle, concentric, hierarchical)
- Good performance with large graphs
- Easy to use

**Cons**:
- Less flexible than D3
- Heavier library

#### Option 3: Vis.js / vis-network
**Pros**:
- Very easy to use
- Good built-in interactions
- Nice default styling

**Cons**:
- Less control over customization
- Somewhat dated

#### Option 4: Sigma.js
**Pros**:
- Optimized for large graphs
- WebGL rendering for performance
- Good for 1000+ nodes

**Cons**:
- Less feature-rich
- Fewer layout options

**Recommendation**: Start with **Cytoscape.js** for quick MVP, migrate to **D3.js** if we need more customization.

---

## Architecture

```
┌─────────────────────────────────────────┐
│  Go-Scope CLI Tool (Phase 1)            │
│  Output: JSON with full extract data    │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Visualizer Web App (Phase 2)           │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Input: JSON from go-scope        │ │
│  └───────────────┬────────────────────┘ │
│                  │                       │
│                  ▼                       │
│  ┌────────────────────────────────────┐ │
│  │  Parser (TypeScript)               │ │
│  │  - Parse JSON extract              │ │
│  │  - Build graph data structure      │ │
│  └───────────────┬────────────────────┘ │
│                  │                       │
│                  ▼                       │
│  ┌────────────────────────────────────┐ │
│  │  Graph Renderer (Cytoscape/D3)     │ │
│  │  - Layout nodes by depth           │ │
│  │  - Draw connections                │ │
│  │  - Handle interactions             │ │
│  └───────────────┬────────────────────┘ │
│                  │                       │
│                  ▼                       │
│  ┌────────────────────────────────────┐ │
│  │  UI Components (React/Vue)         │ │
│  │  - Control panel                   │ │
│  │  - Code viewer                     │ │
│  │  - Search/filter                   │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

---

## File Structure

```
visualizer/
├── package.json
├── tsconfig.json
├── vite.config.ts         # Build tool
├── index.html
├── src/
│   ├── main.ts            # Entry point
│   ├── App.vue            # or App.tsx
│   ├── types/
│   │   └── extract.ts     # TypeScript types for go-scope JSON
│   ├── components/
│   │   ├── GraphView.vue  # Main graph visualization
│   │   ├── CodePanel.vue  # Code display panel
│   │   ├── ControlPanel.vue  # Filters, search, etc.
│   │   ├── NodeInfo.vue   # Node detail popup
│   │   └── Minimap.vue    # Overview map
│   ├── graph/
│   │   ├── parser.ts      # Parse JSON to graph structure
│   │   ├── layout.ts      # Graph layout algorithms
│   │   └── renderer.ts    # Graph rendering logic
│   ├── utils/
│   │   ├── fileReader.ts  # Load JSON from file/URL
│   │   └── export.ts      # Export PNG/SVG
│   └── styles/
│       └── main.css
└── public/
    └── examples/          # Example JSON files
```

---

## JSON Schema (Input)

The visualizer will read the JSON output from go-scope:

```typescript
interface Extract {
  target: Symbol;
  references: Reference[];
  external: string[];
  callers?: Caller[];
  metrics?: Metrics;
  gitHistory?: GitBlame[];
  graph?: string;
}

interface Symbol {
  package: string;
  name: string;
  kind: string;
  receiver?: string;
  file: string;
  line: number;
  endLine: number;
  code: string;
  doc: string;
  exported: boolean;
}

interface Reference {
  symbol: Symbol;
  reason: string;
  depth: number;
  external: boolean;
  stub: boolean;
  signature?: string;
  referencedBy: string;
}

// ... (other types)
```

---

## User Workflow

### Step 1: Extract Code
```bash
# User runs go-scope
go-scope --root . --file service.go --line 128 \
  --depth 3 \
  --show-callers \
  --format json \
  --output extract.json
```

### Step 2: Open Visualizer
```bash
# Option A: Open in browser
open http://localhost:3000?file=extract.json

# Option B: CLI launches visualizer
go-scope visualize extract.json
```

### Step 3: Explore Interactively
- See target symbol in center
- Depth 1 dependencies around it
- Click any node to:
  - Expand its dependencies
  - View its code
  - See where it's used
- Filter by type, package, etc.
- Search for symbols
- Export visualization

---

## Layout Options

### Option 1: Radial Layout (Recommended)
```
          Depth 2
         /   |   \
        /    |    \
   Depth 1  Target  Depth 1
        \    |    /
         \   |   /
          Depth 2
```
- Target in center
- Depth 1 in inner circle
- Depth 2 in outer circle
- Clear visual hierarchy

### Option 2: Hierarchical (Tree)
```
                 Target
               /   |   \
              /    |    \
         Dep1    Dep2   Dep3
         / \      |      / \
      Sub1 Sub2  Sub3  Sub4 Sub5
```
- Top-down or left-right
- Clear parent-child relationships
- Good for deep trees

### Option 3: Force-Directed
```
     Dep2---Dep3
      |  \ /  |
      |   X   |
      |  / \  |
   Target  Dep1
      |     |
    Dep4  Dep5
```
- Organic, physics-based layout
- Shows clustering naturally
- Can be chaotic with many nodes

**Recommendation**: Start with **Radial**, add Hierarchical as alternative

---

## MVP Features (First Release)

### Must-Have
- [ ] Load JSON file (drag-drop or file picker)
- [ ] Display graph with nodes and edges
- [ ] Radial layout centered on target
- [ ] Depth-based coloring
- [ ] Hover to see node info
- [ ] Click node to view code
- [ ] Zoom and pan
- [ ] Basic styling

### Nice-to-Have
- [ ] Expand/collapse nodes
- [ ] Search functionality
- [ ] Filter by type/package
- [ ] Export to PNG
- [ ] Dark/light theme
- [ ] Save/load layout state

### Future
- [ ] Multiple layouts (radial, tree, force)
- [ ] Caller view (reverse dependencies)
- [ ] Diff mode (compare two extracts)
- [ ] Timeline (git history integration)
- [ ] Collaborative annotations
- [ ] Embed in documentation

---

## Development Plan

### Phase 2.1: Setup (Week 7)
- [ ] Initialize Vite + TypeScript + React project
- [ ] Install Cytoscape.js
- [ ] Create basic project structure
- [ ] Write TypeScript types for JSON schema

### Phase 2.2: Core Visualization (Week 8)
- [ ] JSON parser (JSON → Graph data structure)
- [ ] Basic graph renderer with Cytoscape
- [ ] Radial layout implementation
- [ ] Node styling by depth and type
- [ ] Edge styling by reference type

### Phase 2.3: Interactions (Week 9)
- [ ] Hover popups with node info
- [ ] Click to view code in side panel
- [ ] Zoom and pan controls
- [ ] Node expand/collapse
- [ ] Search functionality

### Phase 2.4: Polish (Week 10)
- [ ] Filters (by type, package, depth)
- [ ] Export to PNG/SVG
- [ ] Dark/light theme
- [ ] Responsive design
- [ ] Examples and documentation

---

## Example Visualization Mockup

```
┌────────────────────────────────────────────────────────────────┐
│ Go-Scope Visualizer                            [🔍] [⚙️] [📥] │
├────────────────────┬───────────────────────────────────────────┤
│                    │                                           │
│  Filters           │              Graph View                   │
│  ◉ Functions       │                                           │
│  ◉ Methods         │         ╭─────────╮                       │
│  ◉ Types           │    ╭────│  Dep1   │────╮                 │
│  ◉ Interfaces      │    │    ╰─────────╯    │                 │
│  ○ Vars/Consts     │    │                    │                 │
│                    │  ╭─────────╮      ╭─────────╮            │
│  Depth             │  │  Dep2   │      │  Dep3   │            │
│  ◉ 0 (target)      │  ╰────┬────╯      ╰────┬────╯            │
│  ◉ 1               │       │                 │                 │
│  ◉ 2               │       │    ╭─────────╮  │                │
│  ○ 3               │       └────│ TARGET  │──┘                │
│                    │            ╰─────────╯                    │
│  Search            │              ↙   ↘                        │
│  [_____________]   │       ╭─────────╮ ╭─────────╮            │
│                    │       │  Dep4   │ │  Dep5   │            │
│  Legend            │       ╰─────────╯ ╰─────────╯            │
│  🔵 Function       │                                           │
│  🟢 Type           │  [Zoom: 100%] [Pan: ↑↓←→] [Reset]       │
│  🟡 Method         │                                           │
│  🔴 Interface      │                                           │
│                    │                                           │
├────────────────────┼───────────────────────────────────────────┤
│                    │                                           │
│  Code View         │  // validateEmail checks if email is... │
│                    │  func validateEmail(email string) error {│
│  validateEmail     │      if !strings.Contains(email, "@") { │
│  pkg/service       │          return errors.New("invalid")   │
│  validation.go:45  │      }                                   │
│                    │      return nil                          │
│  [Copy] [Jump]     │  }                                       │
│                    │                                           │
└────────────────────┴───────────────────────────────────────────┘
```

---

## Inspiration / Reference Projects

1. **Go Callvis** - Similar concept for call graphs
   - https://github.com/ofabry/go-callvis
   - Uses graphviz, static output
   - We'll make it interactive

2. **Sourcegraph** - Code navigation
   - Has interactive code exploration
   - Expensive/complex
   - Our tool is lightweight

3. **Madge** - JS/TS dependency graphs
   - https://github.com/pahen/madge
   - Good visual style
   - We adapt for Go

4. **CodeSee** - Visual codebase maps
   - https://www.codesee.io/
   - Interactive, beautiful
   - Inspiration for UX

---

## Success Metrics

- [ ] Load 100-node graph in < 1 second
- [ ] Smooth zoom/pan (60 FPS)
- [ ] Intuitive without tutorial
- [ ] Users prefer visualizer over text output
- [ ] Can handle depth 5+ extracts
- [ ] Works on mobile (basic viewing)

---

## Integration with Phase 1

### CLI Extension
```bash
# Phase 1: Extract to JSON
go-scope --root . --file foo.go --line 10 --format json --output out.json

# Phase 2: Visualize
go-scope visualize out.json  # Opens browser

# Or combined:
go-scope --root . --file foo.go --line 10 --visualize  # Extract + visualize
```

### Web Service Option
```bash
# Start web server
go-scope serve --port 8080

# Navigate to http://localhost:8080
# Drag-drop JSON files or enter file path
```

---

## Notes

- **Standalone Tool**: Visualizer can be separate project or integrated
- **No Backend Needed**: Pure frontend, reads JSON files
- **Offline-First**: Works without internet
- **Shareable**: Can host on GitHub Pages for sharing visualizations
- **Embeddable**: Can embed in documentation sites

---

## Deliverables

1. **Web App**: Standalone HTML/JS/CSS that runs in browser
2. **Documentation**: User guide for visualizer
3. **Examples**: Pre-generated visualizations to showcase
4. **Integration**: CLI command to launch visualizer
5. **Deployment**: Hosted version on GitHub Pages

---

**Status**: 📋 Documented, ready for Phase 2 after v1.0.0 core tool

**Priority**: High - This will make the tool **significantly more useful**

**Estimated Time**: 3-4 weeks after Phase 1 complete

---

Let me know if you want me to adjust anything about this plan!
