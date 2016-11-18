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
	"strconv"
	"strings"

	"github.com/andreas-jonsson/nanovgo"
)

type FillAttribute struct {
	has   bool
	color nanovgo.Color
}

func (a FillAttribute) Has() bool {
	return a.has
}

func (a FillAttribute) Color() nanovgo.Color {
	return a.color
}

type StrokeAttribute struct {
	FillAttribute
	width float32
}

func (a StrokeAttribute) Width() float32 {
	return a.width
}

type Attributes struct {
	Stroke  StrokeAttribute
	Fill    StrokeAttribute
	Unknown map[string]string
}

func parseColor(s string) nanovgo.Color {
	ls := len(s)
	if s[0] != '#' || (ls != 7 && ls != 9) {
		return nanovgo.Color{}
	}
	s = s[1:]

	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)

	col := nanovgo.RGB(byte(r), byte(g), byte(b))
	if ls > 7 {
		a, _ := strconv.ParseUint(s[6:8], 16, 8)
		col = col.TransRGBA(byte(a))
	}
	return col
}

func parseAttributes(attrib *Attributes, value string) error {
	attrib.Unknown = make(map[string]string)
	params := strings.Split(value, ";")

	for _, param := range params {
		kv := strings.Split(param, ":")
		if len(kv) != 2 {
			continue
		}

		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])

		switch k {
		case "stroke":
			a := &attrib.Stroke
			a.color = parseColor(v)
			a.has = true
		case "stroke-width":
			f, _ := strconv.ParseFloat(v, 32)
			a := &attrib.Stroke
			a.width = float32(f)
			a.has = true
		case "fill":
			a := &attrib.Fill
			a.color = parseColor(v)
			a.has = true
		default:
			attrib.Unknown[k] = v
		}
	}

	return nil
}
