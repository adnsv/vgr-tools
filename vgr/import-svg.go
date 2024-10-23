package vgr

import (
	"errors"
	"fmt"
	"os"

	"github.com/adnsv/svg"
)

// ImportSVGFile reads an SVG file and converts it into a VG structure.
func ImportSVGFile(fn string) (*VG, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	sg, err := svg.Parse(string(data))
	if err != nil {
		return nil, err
	}
	vg := &VG{Filename: fn}

	vb, err := sg.ViewBox.Parse()
	if err != nil {
		w, u1, e1 := sg.Width.AsNumeric()
		h, u2, e2 := sg.Height.AsNumeric()
		if e1 == nil && e2 == nil &&
			(u1 == svg.UnitNone || u1 == svg.UnitPX) &&
			(u2 == svg.UnitNone || u2 == svg.UnitPX) {
			vb = &svg.ViewBoxValue{
				Width:  w,
				Height: h,
			}
		} else {
			return nil, fmt.Errorf("bad svg.viewBox attribute: %s", err)
		}
	}

	vg.ViewBox = *vb

	xform := svg.UnitTransform()
	readGroup(vg, &sg.Group, xform)

	return vg, nil
}

func lengthPixels(l svg.Length, reflength *float64) (float64, error) {
	v, u, err := l.AsNumeric()
	if err != nil {
		return 0, err
	}
	switch u {
	case svg.UnitNone, svg.UnitPX:
		return v, nil
	case svg.UnitPercent:
		if reflength == nil {
			return 0, fmt.Errorf("unsupported length percentage")
		} else {
			return *reflength * v / 100.0, nil
		}
	default:
		return 0, fmt.Errorf("unsupported length units")
	}
}

func readGroup(vg *VG, g *svg.Group, xform *svg.Transform) error {
	if id := g.ID(); id != "" {
		vg.PushID(id)
		defer vg.PopID()
	}

	if g.Transform != nil {
		xform = svg.Concatenate(xform, g.Transform)
	}

	needsLayer := g.Opacity != nil && *g.Opacity < 1.0
	if needsLayer {
		vg.StartLayer(*g.Opacity)
	}

	for _, it := range g.Items {
		switch v := it.(type) {
		case *svg.Group:
			err := readGroup(vg, v, xform)
			if err != nil {
				return err
			}

		case *svg.Line:

		case *svg.Rect:
			err := readRect(vg, v, xform)
			if err != nil {
				return err
			}

		case *svg.Circle:
			err := readCircle(vg, v, xform)
			if err != nil {
				return err
			}

		case *svg.Ellipse:
			err := readEllipse(vg, v, xform)
			if err != nil {
				return err
			}

		case *svg.Polygon:
			err := readPolygon(vg, v, xform)
			if err != nil {
				return err
			}

		case *svg.Path:
			err := readPath(vg, v, xform)
			if err != nil {
				return err
			}

		default:
			return errors.New("unsupported element tag")
		}
	}

	if needsLayer {
		vg.StopLayer()
	}

	return nil
}

func readRect(vg *VG, p *svg.Rect, xform *svg.Transform) error {
	if id := p.ID(); id != "" {
		vg.PushID(id)
		defer vg.PopID()
	}

	x, err := lengthPixels(p.X, &vg.ViewBox.Width)
	if err != nil {
		return err
	}
	y, err := lengthPixels(p.Y, &vg.ViewBox.Height)
	if err != nil {
		return err
	}
	width, err := lengthPixels(p.Width, &vg.ViewBox.Width)
	if err != nil {
		return err
	}
	height, err := lengthPixels(p.Height, &vg.ViewBox.Height)
	if err != nil {
		return err
	}

	rx, ry := 0.0, 0.0
	if p.Rx != "" {
		rx, err = lengthPixels(p.Rx, &width)
		if err != nil {
			return err
		}
		if p.Ry == "" {
			ry = rx
		}
	}
	if p.Ry != "" {
		ry, err = lengthPixels(p.Ry, &height)
		if err != nil {
			return err
		}
		if p.Rx == "" {
			rx = ry
		}
	}

	if rx <= 0 || ry <= 0 {
		vg.MoveTo(xform, svg.Vertex{X: x, Y: y})
		vg.LineTo(xform, svg.Vertex{X: x + width, Y: y})
		vg.LineTo(xform, svg.Vertex{X: x + width, Y: y + height})
		vg.LineTo(xform, svg.Vertex{X: x, Y: y + height})
		vg.Close()
	} else {
		if rx > width*0.5 {
			rx = width * 0.5
		}
		if ry > height*0.5 {
			ry = height * 0.5
		}

		kx := (1.0 - 0.551784) * rx
		ky := (1.0 - 0.551784) * ry

		vg.MoveTo(xform, svg.Vertex{X: x + rx, Y: y})
		vg.LineTo(xform, svg.Vertex{X: x + width - rx, Y: y})
		vg.LineTo(xform, svg.Vertex{X: x + width - rx, Y: y})
		vg.CurveTo(xform, svg.Vertex{X: x + width - kx, Y: y}, svg.Vertex{X: x + width, Y: y + ky}, svg.Vertex{X: x + width, Y: y + ry})
		vg.LineTo(xform, svg.Vertex{X: x + width, Y: y + height - ry})
		vg.CurveTo(xform, svg.Vertex{X: x + width, Y: y - ky}, svg.Vertex{X: x + width - kx, Y: y + height}, svg.Vertex{X: x + width - rx, Y: y + height})
		vg.LineTo(xform, svg.Vertex{X: x + rx, Y: y + height})
		vg.CurveTo(xform, svg.Vertex{X: x + kx, Y: y + height}, svg.Vertex{X: x, Y: y + height - ky}, svg.Vertex{X: x, Y: y + height - ry})
		vg.LineTo(xform, svg.Vertex{X: x, Y: y + ry})
		vg.CurveTo(xform, svg.Vertex{X: x, Y: y + ky}, svg.Vertex{X: x + kx, Y: y}, svg.Vertex{X: x + rx, Y: y})
		vg.Close()
	}

	vg.Fill(calcShapePaint(&p.Shape))

	return nil
}

func readCircle(vg *VG, p *svg.Circle, xform *svg.Transform) error {
	if id := p.ID(); id != "" {
		vg.PushID(id)
		defer vg.PopID()
	}

	cx, err := lengthPixels(p.Cx, &vg.ViewBox.Width)
	if err != nil {
		return err
	}
	cy, err := lengthPixels(p.Cy, &vg.ViewBox.Height)
	if err != nil {
		return err
	}
	r := 1.0
	if p.Radius != "" {
		r, err = lengthPixels(p.Radius, &vg.ViewBox.Width)
		if err != nil {
			return err
		}
	}

	k := 0.551784 * r

	vg.MoveTo(xform, svg.Vertex{X: cx - r, Y: cy})
	vg.CurveTo(xform,
		svg.Vertex{X: cx - r, Y: cy - k},
		svg.Vertex{X: cx - k, Y: cy - r},
		svg.Vertex{X: cx, Y: cy - r})
	vg.CurveTo(xform,
		svg.Vertex{X: cx + k, Y: cy - r},
		svg.Vertex{X: cx + r, Y: cy - k},
		svg.Vertex{X: cx + r, Y: cy})
	vg.CurveTo(xform,
		svg.Vertex{X: cx + r, Y: cy + k},
		svg.Vertex{X: cx + k, Y: cy + r},
		svg.Vertex{X: cx, Y: cy + r})
	vg.CurveTo(xform,
		svg.Vertex{X: cx - k, Y: cy + r},
		svg.Vertex{X: cx - r, Y: cy + k},
		svg.Vertex{X: cx - r, Y: cy})
	vg.Close()

	vg.Fill(calcShapePaint(&p.Shape))
	return nil
}

func readEllipse(vg *VG, p *svg.Ellipse, xform *svg.Transform) error {
	if id := p.ID(); id != "" {
		vg.PushID(id)
		defer vg.PopID()
	}

	cx, err := lengthPixels(p.Cx, &vg.ViewBox.Width)
	if err != nil {
		return err
	}
	cy, err := lengthPixels(p.Cy, &vg.ViewBox.Height)
	if err != nil {
		return err
	}
	rx, ry := 0.0, 0.0
	if p.Rx != "" {
		rx, err = lengthPixels(p.Rx, &vg.ViewBox.Width)
		if err != nil {
			return err
		}
		if p.Ry == "" {
			ry = rx
		}
	}
	if p.Ry != "" {
		ry, err = lengthPixels(p.Ry, &vg.ViewBox.Height)
		if err != nil {
			return err
		}
		if p.Rx == "" {
			rx = ry
		}
	}

	kx := 0.551784 * rx
	ky := 0.551784 * ry

	vg.MoveTo(xform, svg.Vertex{X: cx - rx, Y: cy})
	vg.CurveTo(xform,
		svg.Vertex{X: cx - rx, Y: cy - ky},
		svg.Vertex{X: cx - kx, Y: cy - ry},
		svg.Vertex{X: cx, Y: cy - ry})
	vg.CurveTo(xform,
		svg.Vertex{X: cx + kx, Y: cy - ry},
		svg.Vertex{X: cx + rx, Y: cy - ky},
		svg.Vertex{X: cx + rx, Y: cy})
	vg.CurveTo(xform,
		svg.Vertex{X: cx + rx, Y: cy + ky},
		svg.Vertex{X: cx + kx, Y: cy + ry},
		svg.Vertex{X: cx, Y: cy + ry})
	vg.CurveTo(xform,
		svg.Vertex{X: cx - kx, Y: cy + ry},
		svg.Vertex{X: cx - rx, Y: cy + ky},
		svg.Vertex{X: cx - rx, Y: cy})
	vg.Close()

	vg.Fill(calcShapePaint(&p.Shape))
	return nil
}

func readPolygon(vg *VG, p *svg.Polygon, xform *svg.Transform) error {
	pp, err := svg.ParsePoints(p.Points)
	if err != nil {
		return err
	}

	if len(pp) < 2 {
		return nil
	}

	if id := p.ID(); id != "" {
		vg.PushID(id)
		defer vg.PopID()
	}

	vg.MoveTo(xform, pp[0])
	for _, p := range pp[1:] {
		vg.LineTo(xform, p)
	}
	vg.Close()
	vg.Fill(calcShapePaint(&p.Shape))

	return nil
}

func readPath(vg *VG, p *svg.Path, xform *svg.Transform) error {

	pp, err := svg.ParsePath(p.D)
	if err != nil {
		return err
	}

	if id := p.ID(); id != "" {
		vg.PushID(id)
		defer vg.PopID()
	}

	vv := pp.Vertices
	for _, cmd := range pp.Commands {
		switch cmd {
		case svg.PathClose:
			vg.Close()

		case svg.PathMoveTo:
			if len(vv) < 1 {
				return errors.New("invalid # of vertices in path")
			}
			vg.MoveTo(xform, vv[0])
			vv = vv[1:]

		case svg.PathLineTo:

			if len(vv) < 1 {
				return errors.New("invalid # of vertices in path")
			}
			vg.LineTo(xform, vv[0])
			vv = vv[1:]

		case svg.PathCurveTo:
			if len(vv) < 3 {
				return errors.New("invalid # of vertices in path")
			}
			vg.CurveTo(xform, vv[0], vv[1], vv[2])
			vv = vv[3:]

		default:
			return errors.New("unsupported path command")
		}
	}

	vg.Fill(calcShapePaint(&p.Shape))

	return nil
}

func calcShapePaint(s *svg.Shape) RGBA {
	rgba := RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	}
	if s.Fill != nil {
		if s.Fill.Kind == svg.PaintKindRGB {
			rgba.R = s.Fill.Color.R
			rgba.G = s.Fill.Color.G
			rgba.B = s.Fill.Color.B
		}
	}

	if s.FillOpacity != nil {
		v := *s.FillOpacity
		if s.Opacity != nil {
			v = v * *s.Opacity
		}
		if v < 0.0 {
			rgba.A = 0
		} else if v < 1.0 {
			rgba.A = uint8(v * 255)
		}
	} else if s.Opacity != nil {
		v := *s.Opacity
		if v < 0.0 {
			rgba.A = 0
		} else if v < 1.0 {
			rgba.A = uint8(v * 255)
		}
	}

	return rgba
}