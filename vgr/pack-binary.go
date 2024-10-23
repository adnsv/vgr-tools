package vgr

import "encoding/binary"

func Pack(src *VG) []byte {
	vertex_scale := float64(10)

	width_16 := uint16(src.ViewBox.Width * vertex_scale)
	height_16 := uint16(src.ViewBox.Height * vertex_scale)

	buf := []byte{}
	magic_ver := uint32(0xfff00001)
	block_tag := uint32(0xffee0000)

	start := func(block_id uint32, counter int) {
		buf = binary.LittleEndian.AppendUint32(buf, block_tag|block_id)
		buf = binary.LittleEndian.AppendUint32(buf, uint32(counter))
	}

	buf = binary.LittleEndian.AppendUint32(buf, magic_ver)
	buf = binary.LittleEndian.AppendUint16(buf, width_16)
	buf = binary.LittleEndian.AppendUint16(buf, height_16)

	if len(src.Commands) > 0 {
		start(1, len(src.Commands))
		buf = append(buf, []byte(src.Commands)...)
	}

	if len(src.Vertices) > 0 {
		start(2, len(src.Vertices))
		for _, v := range src.Vertices {
			x := int16((v.X - src.ViewBox.MinX) * vertex_scale)
			y := int16((v.Y - src.ViewBox.MinY) * vertex_scale)
			buf = binary.LittleEndian.AppendUint16(buf, uint16(x))
			buf = binary.LittleEndian.AppendUint16(buf, uint16(y))
		}
	}

	if len(src.ColorIndices) > 0 {
		start(3, len(src.ColorIndices))
		for _, v := range src.ColorIndices {
			buf = binary.LittleEndian.AppendUint16(buf, uint16(v))
		}
	}

	if len(src.ColorValues) > 0 {
		start(4, len(src.ColorValues))
		for _, v := range src.ColorValues {
			buf = append(buf, v.R, v.G, v.B, v.A)
		}
	}

	if len(src.Opacities) > 0 {
		start(5, len(src.Opacities))
		for _, v := range src.Opacities {
			if v < 0.0 {
				v = 0
			} else if v > 1.0 {
				v = 1.0
			}
			buf = append(buf, uint8(v*255))
		}
	}

	if len(src.Ids) > 0 {
		start(6, len(src.Ids))
		s := ""
		cur := 0
		for _, v := range src.Ids {
			cur += len(v)
			s += v
			buf = binary.LittleEndian.AppendUint16(buf, uint16(cur))
		}
		start(7, len(s))
		buf = append(buf, []byte(s)...)
	}

	// eof
	buf = binary.LittleEndian.AppendUint32(buf, block_tag)

	return buf
}
