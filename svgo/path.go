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
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/andreas-jonsson/nanovgo"
)

type (
	QuadTo struct {
		IsAbsolute   bool
		Cx, Cy, X, Y float32
	}

	BezierTo struct {
		IsAbsolute               bool
		C1x, C1y, C2x, C2y, X, Y float32
	}

	MoveTo struct {
		IsAbsolute bool
		X, Y       float32
	}

	ArcTo struct {
		IsAbsolute                    bool
		Rx, Ry, XAxisRotate           float32
		LargeArcFlag, SweepFlag, X, Y float32
	}

	LineTo    MoveTo
	ClosePath struct{}
)

type Path struct {
	Id        string
	Attr      Attributes
	Segments  []interface{}
	Transform nanovgo.TransformMatrix
}

func (p *Path) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	p.Transform = nanovgo.IdentityMatrix()

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			p.Id = attr.Value
		case "style":
			if err := parseAttributes(&p.Attr, attr.Value); err != nil {
				return err
			}
		case "d":
			if err := p.parseSegments(attr.Value); err != nil {
				return err
			}
		}
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		if _, ok := token.(xml.EndElement); ok {
			return nil
		}
	}
}

func cleanSegmentString(s string) string {
	var buf bytes.Buffer
	for i, c := range s {
		if i > 0 && c == '-' && unicode.IsDigit(rune(s[i-1])) {
			fmt.Fprint(&buf, " ")
		}
		fmt.Fprintf(&buf, "%c", c)
	}
	return strings.TrimSpace(buf.String())
}

func (p *Path) parseSegments(value string) error {
	const controlCharacters = "aAcChHmMlLsStTvVzZ"
	value = cleanSegmentString(value)

	splitCommands := func(c rune) bool {
		for _, r := range controlCharacters {
			if r == c {
				return true
			}
		}
		return false
	}

	commandData := strings.FieldsFunc(value, splitCommands)
	commands := []rune{}

	for _, c := range value {
		for _, r := range controlCharacters {
			if r == c {
				commands = append(commands, c)
				break
			}
		}
	}

	toFloat := func(s string) float32 {
		f, err := strconv.ParseFloat(s, 32)
		if err != nil {
			log.Panicln(err)
		}
		return float32(f)
	}

	getArgs := func() []string {
		return strings.FieldsFunc(strings.TrimSpace(commandData[0]), func(c rune) bool {
			return unicode.IsSpace(c) || c == ','
		})
	}

	for _, cmd := range commands {
		isAbsolute := unicode.IsUpper(cmd)
		switch cmd {
		case 'a', 'A':
			args := getArgs()
			commandData = commandData[1:]

			for i := 0; i < len(args); i += 7 {
				p.Segments = append(p.Segments, ArcTo{
					isAbsolute,
					toFloat(args[i]),
					toFloat(args[i+1]),
					toFloat(args[i+2]),
					toFloat(args[i+3]),
					toFloat(args[i+4]),
					toFloat(args[i+5]),
					toFloat(args[i+6]),
				})
			}
		case 'm', 'M':
			args := getArgs()
			commandData = commandData[1:]

			for i := 0; i < len(args); i += 2 {
				p.Segments = append(p.Segments, MoveTo{isAbsolute, toFloat(args[i]), toFloat(args[i+1])})
			}
		case 'l', 'L', 'h', 'H', 'v', 'V':
			//TODO Add real support for H and V.

			args := getArgs()
			commandData = commandData[1:]

			for i := 0; i < len(args); i += 2 {
				p.Segments = append(p.Segments, LineTo{isAbsolute, toFloat(args[i]), toFloat(args[i+1])})
			}
		case 'c', 'C':
			args := getArgs()
			commandData = commandData[1:]

			if len(args)%6 == 0 {
				//Cubic Bezier
				for i := 0; i < len(args); i += 6 {
					p.Segments = append(p.Segments, BezierTo{
						isAbsolute,
						toFloat(args[i]),
						toFloat(args[i+1]),
						toFloat(args[i+2]),
						toFloat(args[i+3]),
						toFloat(args[i+4]),
						toFloat(args[i+5]),
					})
				}
			} else {
				//Quadratic Bezier
				for i := 0; i < len(args); i += 4 {
					p.Segments = append(p.Segments, QuadTo{
						isAbsolute,
						toFloat(args[i]),
						toFloat(args[i+1]),
						toFloat(args[i+2]),
						toFloat(args[i+3]),
					})
				}
			}
		case 'z', 'Z':
			p.Segments = append(p.Segments, ClosePath{})
		default:
			return fmt.Errorf("unknown control character: %c", cmd)
		}
	}

	if len(commandData) > 0 {
		return fmt.Errorf("did not consume all data. (%d items left)\n%q", len(commandData), commandData)
	}

	return nil
}
