# Gio CBU Graph Viewer - Complete Guide

**Version**: 1.5  
**Platform**: Pure Go Desktop Application  
**UI Framework**: Gio (immediate-mode GUI)  
**Status**: Production Ready

---

## 📋 Overview

The Gio CBU Graph Viewer is a native desktop application that visualizes Client Business Unit organizational networks. Built entirely in Go with no JavaScript or HTML dependencies.

### Key Features

✅ **Pure Go** - No web dependencies, single binary  
✅ **Cross-Platform** - Linux, macOS, Windows  
✅ **Immediate-Mode UI** - Gio's efficient rendering  
✅ **gRPC Integration** - Live data from server  
✅ **Auto-Layout** - Circular layout for unpositioned graphs  
✅ **Visual Encoding** - Colors by entity type, line styles by relationship  
✅ **Native Performance** - Direct GPU rendering via OpenGL/Metal/DirectX

---

## 🚀 Quick Start

### Prerequisites

```bash
# macOS
brew install xcode-select
xcode-select --install

# Ubuntu/Debian
sudo apt install libwayland-dev libx11-dev libxkbcommon-x11-dev \
    libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev \
    libvulkan-dev

# Fedora/RHEL
sudo dnf install wayland-devel libX11-devel libxkbcommon-x11-devel \
    mesa-libGLES-devel mesa-libEGL-devel libffi-devel libxcursor-devel \
    vulkan-loader-devel
```

### Build

```bash
# Generate proto code (if not done)
make proto

# Build client
make build-client
```

### Run

```bash
# Terminal 1: Start gRPC server
make run-grpc

# Terminal 2: Start client
make run-client
```

---

## 🎮 Usage

### Environment Variables

```bash
# Customize gRPC server address
export GRPC_SERVER="localhost:50051"

# Customize CBU ID to load
export CBU_ID="BLACKROCK-GLOBAL"

# Run client
make run-client
```

### Command Line

```bash
# Default (localhost:50051, BLACKROCK-GLOBAL)
./bin/kycclient

# Custom server
GRPC_SERVER="prod-server:50051" ./bin/kycclient

# Custom CBU
CBU_ID="VANGUARD-ASIA" ./bin/kycclient

# Both
GRPC_SERVER="prod:50051" CBU_ID="MY-FUND" ./bin/kycclient
```

---

## 🎨 Visual Design

### Entity Colors

| Entity Type | Color | RGB |
|------------|-------|-----|
| Parent | Blue | `rgb(68, 153, 204)` |
| Fund | Green | `rgb(34, 170, 34)` |
| SubFund | Light Green | `rgb(102, 204, 102)` |
| Manager | Orange | `rgb(255, 153, 0)` |
| Custodian | Purple | `rgb(153, 102, 255)` |
| Administrator | Yellow | `rgb(255, 204, 0)` |
| Default | Gray | `rgb(187, 187, 187)` |

### Relationship Colors

| Relationship | Color | Style |
|-------------|-------|-------|
| owns | Forest Green | Solid, thick if ≥50% |
| controls | Dark Orange | Solid |
| delegates | Steel Blue | Solid |
| reports_to | Gray | Solid |
| custodies | Medium Purple | Solid |

### Layout

- **Nodes**: Circles with 50px diameter
- **Labels**: Entity name below each node
- **Edges**: Directed arrows showing relationship direction
- **Title**: Graph name at top
- **Stats**: Entity and relationship counts at bottom

---

## 📐 Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Gio Window                        │
│  ┌───────────────────────────────────────────────┐  │
│  │  Title: "BlackRock Global Equity CBU"        │  │
│  ├───────────────────────────────────────────────┤  │
│  │                                               │  │
│  │    ●──────→●                                  │  │
│  │  Parent   Fund                                │  │
│  │            │                                  │  │
│  │            ↓                                  │  │
│  │            ●                                  │  │
│  │        SubFund                                │  │
│  │                                               │  │
│  ├───────────────────────────────────────────────┤  │
│  │  Stats: 5 entities | 8 relationships         │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
          │
          │ gRPC
          ▼
┌─────────────────────────────────────────────────────┐
│            gRPC Server (port 50051)                  │
│  CbuGraphService.GetGraph()                         │
└─────────────────────────────────────────────────────┘
```

---

## 🧩 Code Structure

### Entry Point: `cmd/client/main.go`

```go
func main() {
    go run()
    app.Main()
}

func run() {
    // 1. Create window
    w := app.NewWindow(...)
    
    // 2. Connect to gRPC
    client := pb.NewCbuGraphServiceClient(conn)
    
    // 3. Fetch graph
    graph := client.GetGraph(ctx, req)
    
    // 4. Create view
    view := ui.NewGraphView()
    view.SetGraph(graph)
    
    // 5. Event loop
    for event := range w.Events() {
        view.Layout(gtx, th)
    }
}
```

### UI Package: `internal/ui/graph.go`

```go
type GraphView struct {
    Graph      *pb.CbuGraph
    Scale      float32
    OffsetX    float32
    OffsetY    float32
    NodeSize   float32
    ShowLabels bool
}

func (gv *GraphView) Layout(gtx layout.Context, th *material.Theme)
func (gv *GraphView) drawEdges(gtx layout.Context)
func (gv *GraphView) drawNodes(gtx layout.Context, th *material.Theme)
func (gv *GraphView) applyAutoLayout()
```

---

## 🎨 Layout Algorithm

### Auto-Layout (Circular)

If entities don't have X/Y coordinates, the client applies a circular layout:

```go
func applyAutoLayout() {
    centerX := 400
    centerY := 300
    radius := 200
    
    for i, entity := range entities {
        angle := 2π * i / count
        entity.X = centerX + radius * cos(angle)
        entity.Y = centerY + radius * sin(angle)
    }
}
```

### Custom Positions

To provide custom layout, set `x` and `y` in the proto:

```protobuf
message CbuEntity {
    string id = 1;
    string name = 2;
    float x = 8;  // Pixels from left
    float y = 9;  // Pixels from top
}
```

---

## 🔧 Customization

### Change Node Size

```go
view := ui.NewGraphView()
view.NodeSize = 70.0  // Larger nodes
```

### Disable Labels

```go
view.ShowLabels = false
```

### Adjust Scale

```go
view.Scale = 1.5  // Zoom in
view.Scale = 0.7  // Zoom out
```

### Custom Colors

Edit `internal/ui/graph.go`:

```go
func colorForEntityType(entityType string) color.NRGBA {
    switch entityType {
    case "Fund":
        return color.NRGBA{R: 255, G: 0, B: 0, A: 255} // Red
    // ...
    }
}
```

---

## 🧪 Testing

### Test with Example Graph

```bash
# Terminal 1: Start server
make run-grpc

# Terminal 2: Run client
make run-client
```

Expected window:
- **Title**: "BlackRock Global Equity CBU Network"
- **Nodes**: 5 entities (E1-E5)
- **Edges**: 5 relationships
- **Layout**: Circular arrangement

### Test with Custom Data

Update `internal/service/cbu_graph_service.go` to return different graph data.

---

## 📦 Distribution

### Single Binary

```bash
# Build
make build-client

# Distribute
cp bin/kycclient ~/Desktop/
```

### Cross-Compile

```bash
# macOS → Windows
GOOS=windows GOARCH=amd64 go build -o kycclient.exe ./cmd/client

# macOS → Linux
GOOS=linux GOARCH=amd64 go build -o kycclient-linux ./cmd/client

# Linux → macOS
GOOS=darwin GOARCH=amd64 go build -o kycclient-mac ./cmd/client
```

---

## 🚀 Advanced Features

### Add Zoom/Pan

```go
type GraphView struct {
    // ... existing fields
    Zoom      float32
    PanX      float32
    PanY      float32
}

// Handle mouse wheel for zoom
// Handle mouse drag for pan
```

### Add Node Selection

```go
type GraphView struct {
    SelectedNode string
}

// Handle click events
// Highlight selected node
```

### Add Real-Time Updates

```go
// Poll for updates
ticker := time.NewTicker(5 * time.Second)
go func() {
    for range ticker.C {
        graph := client.GetGraph(ctx, req)
        view.SetGraph(graph)
        w.Invalidate() // Trigger redraw
    }
}()
```

### Add Export

```go
// Export to PNG
// Export to SVG
// Export to DOT format
```

---

## 🐛 Troubleshooting

### "cannot connect to gRPC server"

**Cause**: Server not running  
**Solution**:
```bash
# Terminal 1
make run-grpc

# Terminal 2
make run-client
```

### "failed to build: missing dependencies"

**Cause**: Missing system libraries  
**Solution**: Install prerequisites (see Quick Start)

### Blank window

**Cause**: No graph data or layout issue  
**Solution**: Check server logs, verify graph has entities

### Window doesn't appear

**Cause**: Platform-specific issue  
**Solution**:
```bash
# macOS: Grant accessibility permissions
# Linux: Check X11/Wayland
# Windows: Run as administrator
```

---

## 🎯 Use Cases

### 1. **Regulatory Reporting**
Visualize ultimate beneficial ownership chains for FATCA/CRS/AMLD5 compliance.

### 2. **Risk Assessment**
Identify concentration risk and contagion paths in organizational networks.

### 3. **Due Diligence**
Present fund structures to clients and regulators.

### 4. **Monitoring Dashboards**
Real-time view of organizational changes.

### 5. **Audit Trail**
Visual record of ownership structures at specific points in time.

---

## 🔮 Future Enhancements

### Phase 2
- [ ] Interactive node selection
- [ ] Zoom and pan
- [ ] Filter by entity type
- [ ] Search nodes
- [ ] Export to image/PDF

### Phase 3
- [ ] Force-directed layout (physics-based)
- [ ] Hierarchical layout
- [ ] Timeline view (version history)
- [ ] Side panel with details
- [ ] Edit mode (add/remove entities)

### Phase 4
- [ ] Multi-graph comparison
- [ ] Animation for changes
- [ ] 3D visualization
- [ ] VR/AR support
- [ ] Web Assembly export

---

## 📚 Resources

### Gio Documentation
- [Gio Website](https://gioui.org/)
- [Gio Tutorial](https://gioui.org/doc/learn)
- [Gio Examples](https://git.sr.ht/~eliasnaur/gio-example)

### Graph Visualization
- [Force-Directed Graphs](https://en.wikipedia.org/wiki/Force-directed_graph_drawing)
- [Graph Layout Algorithms](https://cs.brown.edu/people/rtamassi/gdhandbook/)

---

## 🎉 Summary

✅ **Pure Go Desktop App**  
✅ **No Web Dependencies**  
✅ **Cross-Platform**  
✅ **gRPC Integration**  
✅ **Auto-Layout**  
✅ **Production Ready**  

The Gio CBU Graph Viewer provides a native, high-performance way to visualize organizational networks directly from your gRPC service.

---

**Version**: 1.5  
**Last Updated**: 2024  
**Status**: Production Ready
