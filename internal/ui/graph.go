package ui

import (
	"image"
	"image/color"
	"math"

	"fmt"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
)

// GraphView manages the CBU graph visualization
type GraphView struct {
	Graph     *pb.CbuGraph
	Scale     float32
	OffsetX   float32
	OffsetY   float32
	NodeSize  float32
	ShowLabels bool
}

// NewGraphView creates a new graph view with default settings
func NewGraphView() *GraphView {
	return &GraphView{
		Scale:      1.0,
		NodeSize:   50.0,
		ShowLabels: true,
	}
}

// SetGraph updates the graph data and applies auto-layout if needed
func (gv *GraphView) SetGraph(graph *pb.CbuGraph) {
	gv.Graph = graph
	gv.applyAutoLayout()
}

// Layout draws the CBU graph with nodes and edges
func (gv *GraphView) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if gv.Graph == nil {
		return gv.layoutEmpty(gtx, th)
	}

	// Draw edges first (behind nodes)
	gv.drawEdges(gtx)

	// Draw nodes on top
	gv.drawNodes(gtx, th)

	// Draw title
	gv.drawTitle(gtx, th)

	// Draw stats
	gv.drawStats(gtx, th)

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// layoutEmpty shows a message when no graph is loaded
func (gv *GraphView) layoutEmpty(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.H4(th, "Loading CBU Graph...")
		label.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
		label.Alignment = text.Middle
		return label.Layout(gtx)
	})
}

// drawEdges renders all relationships as lines
func (gv *GraphView) drawEdges(gtx layout.Context) {
	for _, rel := range gv.Graph.Relationships {
		from := gv.findEntity(rel.FromId)
		to := gv.findEntity(rel.ToId)
		if from == nil || to == nil {
			continue
		}

		x1 := from.X*gv.Scale + gv.OffsetX
		y1 := from.Y*gv.Scale + gv.OffsetY
		x2 := to.X*gv.Scale + gv.OffsetX
		y2 := to.Y*gv.Scale + gv.OffsetY

		gv.drawEdge(gtx, x1, y1, x2, y2, rel)
	}
}

// drawEdge renders a single edge with appropriate styling
func (gv *GraphView) drawEdge(gtx layout.Context, x1, y1, x2, y2 float32, rel *pb.CbuRelationship) {
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(x1, y1))
	p.LineTo(f32.Pt(x2, y2))

	// Style based on relationship type
	col := gv.colorForRelationType(rel.RelationType)
	width := unit.Dp(2)
	if rel.ControlPct >= 50 {
		width = unit.Dp(3) // Thicker for majority control
	}

	stroke := clip.Stroke{
		Path:  p.End(),
		Width: float32(gtx.Dp(width)),
	}.Op()

	paint.FillShape(gtx.Ops, col, stroke)

	// Draw arrowhead
	gv.drawArrowhead(gtx, x1, y1, x2, y2, col)
}

// drawArrowhead draws an arrow at the end of an edge
func (gv *GraphView) drawArrowhead(gtx layout.Context, x1, y1, x2, y2 float32, col color.NRGBA) {
	// Calculate angle
	dx := x2 - x1
	dy := y2 - y1
	angle := float32(math.Atan2(float64(dy), float64(dx)))

	// Arrow size
	size := float32(10)
	angleOffset := float32(math.Pi / 6) // 30 degrees

	// Arrow points
	p1x := x2 - size*float32(math.Cos(float64(angle-angleOffset)))
	p1y := y2 - size*float32(math.Sin(float64(angle-angleOffset)))
	p2x := x2 - size*float32(math.Cos(float64(angle+angleOffset)))
	p2y := y2 - size*float32(math.Sin(float64(angle+angleOffset)))

	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(x2, y2))
	p.LineTo(f32.Pt(p1x, p1y))
	p.LineTo(f32.Pt(p2x, p2y))
	p.Close()

	paint.FillShape(gtx.Ops, col, clip.Outline{Path: p.End()}.Op())
}

// drawNodes renders all entities as circles
func (gv *GraphView) drawNodes(gtx layout.Context, th *material.Theme) {
	for _, entity := range gv.Graph.Entities {
		x := entity.X*gv.Scale + gv.OffsetX
		y := entity.Y*gv.Scale + gv.OffsetY
		gv.drawNode(gtx, x, y, entity, th)
	}
}

// drawNode renders a single entity node
func (gv *GraphView) drawNode(gtx layout.Context, x, y float32, entity *pb.CbuEntity, th *material.Theme) {
	defer op.Offset(f32.Pt(x-gv.NodeSize/2, y-gv.NodeSize/2)).Push(gtx.Ops).Pop()

	// Draw circle
	col := gv.colorForEntityType(entity.EntityType)
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(gv.NodeSize/2, gv.NodeSize/2))
	
	// Create circle path
	center := f32.Pt(gv.NodeSize/2, gv.NodeSize/2)
	radius := gv.NodeSize / 2
	
	const segments = 32
	for i := 0; i <= segments; i++ {
		angle := float32(i) * 2 * math.Pi / segments
		px := center.X + radius*float32(math.Cos(float64(angle)))
		py := center.Y + radius*float32(math.Sin(float64(angle)))
		if i == 0 {
			p.MoveTo(f32.Pt(px, py))
		} else {
			p.LineTo(f32.Pt(px, py))
		}
	}
	p.Close()

	paint.FillShape(gtx.Ops, col, clip.Outline{Path: p.End()}.Op())

	// Draw border
	borderCol := color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	var borderPath clip.Path
	borderPath.Begin(gtx.Ops)
	for i := 0; i <= segments; i++ {
		angle := float32(i) * 2 * math.Pi / segments
		px := center.X + radius*float32(math.Cos(float64(angle)))
		py := center.Y + radius*float32(math.Sin(float64(angle)))
		if i == 0 {
			borderPath.MoveTo(f32.Pt(px, py))
		} else {
			borderPath.LineTo(f32.Pt(px, py))
		}
	}
	borderPath.Close()
	
	borderStroke := clip.Stroke{
		Path:  borderPath.End(),
		Width: 2,
	}.Op()
	paint.FillShape(gtx.Ops, borderCol, borderStroke)

	// Draw label if enabled
	if gv.ShowLabels {
		gv.drawNodeLabel(gtx, entity, th)
	}
}

// drawNodeLabel renders the entity name below the node
func (gv *GraphView) drawNodeLabel(gtx layout.Context, entity *pb.CbuEntity, th *material.Theme) {
	defer op.Offset(f32.Pt(gv.NodeSize/2, gv.NodeSize+5)).Push(gtx.Ops).Pop()

	label := material.Caption(th, entity.Name)
	label.Color = color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	label.Alignment = text.Middle
	label.Font.Weight = font.Medium
	label.TextSize = unit.Sp(10)
	label.Layout(gtx)
}

// drawTitle renders the graph title at the top
func (gv *GraphView) drawTitle(gtx layout.Context, th *material.Theme) {
	defer op.Offset(f32.Pt(20, 20)).Push(gtx.Ops).Pop()

	title := material.H5(th, gv.Graph.Name)
	title.Color = color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	title.Font.Weight = font.Bold
	title.Layout(gtx)
}

// drawStats renders statistics in the bottom-left corner
func (gv *GraphView) drawStats(gtx layout.Context, th *material.Theme) {
	defer op.Offset(f32.Pt(20, float32(gtx.Constraints.Max.Y-60))).Push(gtx.Ops).Pop()

	stats := material.Caption(th, 
		fmt.Sprintf("Entities: %d | Relationships: %d", 
			gv.Graph.EntityCount, 
			gv.Graph.RelationshipCount))
	stats.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	stats.Layout(gtx)
}

// Helper functions

// findEntity looks up an entity by ID
func (gv *GraphView) findEntity(id string) *pb.CbuEntity {
	for _, e := range gv.Graph.Entities {
		if e.Id == id {
			return e
		}
	}
	return nil
}

// colorForEntityType returns a color based on entity type
func (gv *GraphView) colorForEntityType(entityType string) color.NRGBA {
	switch entityType {
	case "Parent":
		return color.NRGBA{R: 68, G: 153, B: 204, A: 255} // Blue
	case "Fund":
		return color.NRGBA{R: 34, G: 170, B: 34, A: 255} // Green
	case "SubFund":
		return color.NRGBA{R: 102, G: 204, B: 102, A: 255} // Light green
	case "Manager":
		return color.NRGBA{R: 255, G: 153, B: 0, A: 255} // Orange
	case "Custodian":
		return color.NRGBA{R: 153, G: 102, B: 255, A: 255} // Purple
	case "Administrator":
		return color.NRGBA{R: 255, G: 204, B: 0, A: 255} // Yellow
	default:
		return color.NRGBA{R: 187, G: 187, B: 187, A: 255} // Gray
	}
}

// colorForRelationType returns a color based on relationship type
func (gv *GraphView) colorForRelationType(relType string) color.NRGBA {
	switch relType {
	case "owns":
		return color.NRGBA{R: 34, G: 139, B: 34, A: 180} // Forest green
	case "controls":
		return color.NRGBA{R: 255, G: 140, B: 0, A: 180} // Dark orange
	case "delegates":
		return color.NRGBA{R: 70, G: 130, B: 180, A: 180} // Steel blue
	case "reports_to":
		return color.NRGBA{R: 128, G: 128, B: 128, A: 180} // Gray
	case "custodies":
		return color.NRGBA{R: 147, G: 112, B: 219, A: 180} // Medium purple
	default:
		return color.NRGBA{R: 100, G: 100, B: 100, A: 180} // Dark gray
	}
}

// applyAutoLayout applies a simple circular layout if positions are not set
func (gv *GraphView) applyAutoLayout() {
	if gv.Graph == nil || len(gv.Graph.Entities) == 0 {
		return
	}

	// Check if any entity already has positions
	hasPositions := false
	for _, e := range gv.Graph.Entities {
		if e.X != 0 || e.Y != 0 {
			hasPositions = true
			break
		}
	}

	if hasPositions {
		return // Use existing positions
	}

	// Apply circular layout
	centerX := float32(400)
	centerY := float32(300)
	radius := float32(200)

	count := len(gv.Graph.Entities)
	for i, entity := range gv.Graph.Entities {
		angle := float32(i) * 2 * math.Pi / float32(count)
		entity.X = centerX + radius*float32(math.Cos(float64(angle)))
		entity.Y = centerY + radius*float32(math.Sin(float64(angle)))
	}

	// Center the graph
	gv.centerGraph()
}

// centerGraph adjusts offsets to center the graph in the viewport
func (gv *GraphView) centerGraph() {
	if gv.Graph == nil || len(gv.Graph.Entities) == 0 {
		return
	}

	// Find bounding box
	minX, minY := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY := float32(-math.MaxFloat32), float32(-math.MaxFloat32)

	for _, e := range gv.Graph.Entities {
		if e.X < minX {
			minX = e.X
		}
		if e.X > maxX {
			maxX = e.X
		}
		if e.Y < minY {
			minY = e.Y
		}
		if e.Y > maxY {
			maxY = e.Y
		}
	}

	// Center offset
	width := maxX - minX
	height := maxY - minY
	gv.OffsetX = (800 - width*gv.Scale) / 2 - minX*gv.Scale
	gv.OffsetY = (600 - height*gv.Scale) / 2 - minY*gv.Scale
}
