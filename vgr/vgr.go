package vgr

import (
	"math"

	"github.com/adnsv/svg"
)

type VG struct {
	Filename     string
	ViewBox      svg.ViewBoxValue
	Commands     string
	Vertices     []svg.Vector
	ColorIndices []int
	ColorValues  []RGBA
	Opacities    []float64
	Ids          []string
}

type RGBA struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func addVertex(vg *VG, xform *svg.Transform, v svg.Vertex) {
	x, y := xform.CalcAbs(v.X, v.Y)
	vg.Vertices = append(vg.Vertices, svg.Vertex{X: x, Y: y})
}

func (vg *VG) addColor(c RGBA) int {
	for i, v := range vg.ColorValues {
		if v == c {
			return i
		}
	}
	vg.ColorValues = append(vg.ColorValues, c)
	return len(vg.ColorValues) - 1
}

func (vg *VG) Close() {
	vg.Commands += "z"
}

func (vg *VG) MoveTo(xform *svg.Transform, v svg.Vertex) {
	vg.Commands += "m"
	addVertex(vg, xform, v)
}
func (vg *VG) LineTo(xform *svg.Transform, v svg.Vertex) {
	vg.Commands += "l"
	addVertex(vg, xform, v)
}
func (vg *VG) CurveTo(xform *svg.Transform, c1, c2, v svg.Vertex) {
	vg.Commands += "c"
	addVertex(vg, xform, c1)
	addVertex(vg, xform, c2)
	addVertex(vg, xform, v)
}
func (vg *VG) Fill(rgba RGBA) {
	vg.Commands += "f"
	vg.ColorIndices = append(vg.ColorIndices, vg.addColor(rgba))
}
func (vg *VG) StartLayer(opacity float64) {
	vg.Commands += "{"
	vg.Opacities = append(vg.Opacities, opacity)
}
func (vg *VG) StopLayer() {
	vg.Commands += "}"
}
func (vg *VG) PushID(id string) {
	vg.Commands += "["
	vg.Ids = append(vg.Ids, id)
}
func (vg *VG) PopID() {
	vg.Commands += "]"
}

// Arc draws a circular arc using cubic bezier curves
func (vg *VG) Arc(xform *svg.Transform, center svg.Vector, radius float64, startAngle, endAngle float64) {

	const numSegments = 4 // Number of bezier curves to use for a full circle

	sweepAngle := endAngle - startAngle
	numCurves := int(math.Ceil(math.Abs(sweepAngle) / (2 * math.Pi) * numSegments))
	if numCurves < 1 {
		numCurves = 1
	}

	anglePerCurve := sweepAngle / float64(numCurves)

	// Magic number for optimal approximation
	handle := radius * 0.551915024494

	for i := 0; i < numCurves; i++ {
		theta := startAngle + anglePerCurve*float64(i)
		nextTheta := theta + anglePerCurve

		// Start point
		x0 := center.X + radius*math.Cos(theta)
		y0 := center.Y + radius*math.Sin(theta)

		// End point
		x3 := center.X + radius*math.Cos(nextTheta)
		y3 := center.Y + radius*math.Sin(nextTheta)

		// Control points
		x1 := x0 - handle*math.Sin(theta)
		y1 := y0 + handle*math.Cos(theta)

		x2 := x3 + handle*math.Sin(nextTheta)
		y2 := y3 - handle*math.Cos(nextTheta)

		if i == 0 {
			vg.MoveTo(xform, svg.Vertex{X: x0, Y: y0})
		}
		vg.CurveTo(xform,
			svg.Vertex{X: x1, Y: y1},
			svg.Vertex{X: x2, Y: y2},
			svg.Vertex{X: x3, Y: y3})
	}
}
