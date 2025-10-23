# 🚀 Quick Start: Go Scope Visualizer

## Your Extract is Ready!

**File**: `accounts-extract.json` (67KB)
**Target**: `SearchAccounts` method from customer-management-api
**Content**: 82 nodes, 120 edges, 21 external packages

---

## Open the Visualizer

### 1. Open in Browser
```
http://localhost:8080
```

### 2. Load the JSON File
- Click the **"📁 Load Extract JSON"** button
- Navigate to: `/home/jon/personal/extract-scope-go/`
- Select: `accounts-extract.json`
- Click Open

### 3. Explore!

The interactive graph will appear showing:
- 🔴 **Red node** = SearchAccounts (your target)
- 🔵 **Blue nodes** = Internal dependencies
- ⚪ **Gray nodes** = External packages
- **Lines** = Dependency relationships

---

## Quick Actions

### View Code
**Click any node** → See its code in the right panel

### Rearrange Graph
**Click & drag nodes** → Reposition for better view

### Zoom
**Mouse wheel** → Zoom in/out
**➕/➖ buttons** → Zoom controls
**🎯 button** → Reset view

### Filter
**☑️ Show External** → Toggle external package visibility
**🏷️ Toggle Labels** → Show/hide node labels

---

## What You'll See

### Red Node: SearchAccounts
The method you extracted at line 71 from accounts.go

**Click it to see**:
- Full source code (42 lines)
- Package info
- File location
- Method signature

### Blue Nodes: Internal Dependencies

**Depth 1** (Direct calls):
- `NewSpanWithCtx` - Distributed tracing
- `AddEvent` - Trace event logging
- `GetTenantFromContext` - Multi-tenancy
- `Network.Post` - HTTP client
- `ConvertResponse` - Response handler
- `SetError` - Error tracing
- `NewErrorAndLog` - Error creation

**Depth 2** (Called by depth 1):
- OpenTelemetry integration
- Error message formatting
- HTTP header building
- Context management

### Gray Nodes: External Packages
- `context.Context`
- `fmt.Sprintf`
- `json.Marshal`
- `http.StatusNotFound`
- OpenTelemetry packages
- And 16 more...

---

## Example Workflow

### 1. Start with Target
Click the **red SearchAccounts node**
- Read the code
- Understand what it does
- Note the dependencies it calls

### 2. Explore Direct Dependencies
Click **blue nodes** connected to SearchAccounts:
- `NewSpanWithCtx` - See how tracing is initialized
- `GetTenantFromContext` - See multi-tenant handling
- `Network.Post` - See HTTP client usage
- `ConvertResponse` - See response handling

### 3. Go Deeper
Click nodes at **depth 2** to understand implementation:
- How OpenTelemetry integrates
- How errors are formatted
- How headers are built

### 4. Organize View
- **Drag nodes** to group related ones
- **Toggle External** to focus on internal code
- **Zoom in** to see details
- **Zoom out** for overview

---

## Tips

### For Large Graphs
1. Turn off "Show External" first
2. Zoom out to see structure
3. Drag important nodes to center
4. Zoom in to examine details

### To Find Something
- Look at node names in the graph
- Click nodes to see full info
- Check the statistics at bottom

### To Understand Flow
1. Start at red target node
2. Follow edges to blue nodes
3. Click each to see code
4. Build mental model of execution

---

## Keyboard Shortcuts

- **Mouse Wheel** - Zoom
- **Click & Drag Node** - Move node
- **Click & Drag Background** - Pan view
- **Click Node** - View details

---

## Statistics (Bottom Bar)

- **Total Nodes**: 82 symbols
- **Total Edges**: 120 dependencies
- **Max Depth**: 2 levels
- **External Refs**: 21 packages

---

## Troubleshooting

**Graph too crowded?**
→ Toggle "Show External" to hide gray nodes

**Can't see labels?**
→ Click 🏷️ button to toggle labels

**Lost in the graph?**
→ Click 🎯 button to reset view

**Want to focus on one area?**
→ Zoom in with mouse wheel

---

## Next Steps

### Extract Other Methods
```bash
# Try another method from same file
cd /home/jon/w/customer-management-api/code
go-scope -file=internal/app/domain/service/accounts.go -line=113 -depth=2 -format=json -output=/tmp/other-method.json
```

### Different Depth Levels
```bash
# Depth 1: Just direct dependencies (cleaner)
go-scope -file=... -line=71 -depth=1 -format=json -output=depth1.json

# Depth 3: More complete picture (more complex)
go-scope -file=... -line=71 -depth=3 -format=json -output=depth3.json
```

### Compare Multiple Methods
Extract different methods and load them one at a time to compare their dependency patterns.

---

## Files Available

📁 **JSON** (for visualizer): `accounts-extract.json`
📝 **Markdown** (for reading): `/tmp/accounts-extract.md`
📋 **Analysis** (detailed): `/tmp/EXTRACTION_SUMMARY.md`

---

## More Information

- **Full Docs**: `web/README.md`
- **Phase 2 Summary**: `docs/PHASE_2_COMPLETE.md`
- **Main README**: `README.md`

---

**Enjoy exploring your code visually!** 🎨

*Go Scope Visualizer - Phase 2*
*Repository: https://github.com/jonathanleahy/goscope*
*PR: https://github.com/jonathanleahy/goscope/pull/1*
