package vgr

import (
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
