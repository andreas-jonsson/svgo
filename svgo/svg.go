/*
Copyright (C) 2016 Andreas T Jonsson

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package svgo

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/andreas-jonsson/nanovgo"
)

type Svg struct {
	Title     string  `xml:"title"`
	Groups    []Group `xml:"g"`
	Transform nanovgo.TransformMatrix
}

type Group struct {
	Id        string
	Attr      Attributes
	Shapes    []interface{}
	Transform nanovgo.TransformMatrix
}

func (g *Group) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	g.Transform = nanovgo.IdentityMatrix()

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			g.Id = attr.Value
		case "style":
			if err := parseAttributes(&g.Attr, attr.Value); err != nil {
				return err
			}
		case "transform":
			//g.TransformString = attr.Value
			//t, err := parseTransform(g.TransformString)
			//g.Transform = &t
		}
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			var shape interface{}

			switch tok.Name.Local {
			case "g":
				shape = &Group{Attr: g.Attr}
			case "path":
				shape = &Path{Attr: g.Attr}
			default:
				return fmt.Errorf("unknown shape: %s", tok.Name.Local)
			}

			if err = decoder.DecodeElement(shape, &tok); err != nil {
				return fmt.Errorf("error decoding element of group: %v", err)
			} else {
				g.Shapes = append(g.Shapes, shape)
			}
		case xml.EndElement:
			return nil
		}
	}
}

func ParseSvg(reader io.Reader, scale float32) (*Svg, error) {
	var s Svg
	s.Transform = nanovgo.IdentityMatrix()
	if scale > 0 {
		nanovgo.ScaleMatrix(scale, scale)
	} else if scale < 0 {
		scale = 1.0 / -scale
		nanovgo.ScaleMatrix(scale, scale)

	}

	err := xml.NewDecoder(reader).Decode(&s)
	if err != nil {
		return &s, err
	}
	return &s, nil
}
