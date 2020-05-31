package go_path

import (
	"github.com/wojnosystems/go-path"
	"io"
	"strings"
)

type goPath struct {
	parts []Componenter
}

func NewRoot() PathMutator {
	return &goPath{
		parts: make([]Componenter, 0, 1),
	}
}

func (p goPath) IsEqual(compareWith path.Pather) bool {
	if compareWithGo, ok := compareWith.(*goPath); !ok {
		return false
	} else {
		if len(p.parts) != len(compareWithGo.parts) {
			return false
		}
		for partIndex, part := range p.parts {
			if !part.(Componenter).IsEqual(compareWithGo.parts[partIndex].(Componenter)) {
				return false
			}
		}
		return true
	}
}

func (p goPath) Serialize() string {
	if len(p.parts) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for _, component := range p.parts {
		sb.WriteString(component.Serialize())
	}
	return sb.String()
}

func Parse(path string) (out Pather, err error) {
	if "" == path {
		return NewRoot(), nil
	}
	outGo := NewRoot()
	var part Componenter
	switch parseComponentType(path) {
	case componentTypeStruct:
		part, err = parseInstanceVariableName
		outGo.Append()
	}
	return out, nil
}

func splitComponents(path string) (components []string, err error) {
	components = make([]string, 0, 10)
	remaining := path
	for {
		// nothing to parse
		if len(remaining) == 0 {
			return
		}
		structStart := strings.Index(path, ".")
		indexedStart := strings.Index(path, "[")
		if -1 == structStart && -1 == indexedStart {
			// didn't find the end, this is the last component
			components = append(components, remaining)
			remaining = ""
			return
		} else {
			startOfNextPart := minInt(indexedStart, structStart)
			if -1 == startOfNextPart {
				// one of them was -1...
				if -1 == indexedStart {
					startOfNextPart = structStart
				} else {
					startOfNextPart = indexedStart
				}
			}
			components = append(components, remaining[0:startOfNextPart])
			remaining = remaining[startOfNextPart:]
		}
	}
}

func (p goPath) Copy() Pather {
	newCopy := NewRoot()
	for _, part := range p.parts {
		newCopy.Append(part)
	}
	return newCopy
}

func (p *goPath) Append(componenter ...Componenter) {
	for _, part := range componenter {
		p.parts = append(p.parts, part)
	}
}

func (p *goPath) Pop(count uint) {
	newLength := maxInt(len(p.parts)-int(count), 0)
	p.parts = p.parts[0:newLength]
}

func (p *goPath) Prepend(itemsToPrepend ...Componenter) {
	newParts := make([]Componenter, len(itemsToPrepend)+len(p.parts))
	for newPartsIndex, part := range itemsToPrepend {
		newParts[newPartsIndex] = part
	}
	for originalIndex, part := range p.parts {
		newParts[originalIndex+len(itemsToPrepend)] = part
	}
	p.parts = newParts
}

func (p *goPath) PopFront(count uint) {
	newLength := maxInt(len(p.parts)-int(count), 0)
	newParts := make([]Componenter, newLength)
	if len(p.parts) > int(count) {
		for i, componenter := range p.parts[count:] {
			newParts[i] = componenter
		}
	}
	p.parts = newParts
}
