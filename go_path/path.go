package go_path

import (
	"fmt"
	"github.com/wojnosystems/go-path"
	"io"
	"strconv"
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

func (p goPath) serialize() string {
	if len(p.parts) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i, component := range p.parts {
		if i != 0 {
			if _, ok := component.(*pathStructInstanceVariable); ok {
				sb.WriteString(".")
			}
		}
		sb.WriteString(component.String())
	}
	return sb.String()
}

func (p goPath) String() string {
	return p.serialize()
}

func Parse(reader io.Reader) (out Pather, err error) {
	lex := newLexer(reader)
	go lex.lex()
	outGo := NewRoot()
	continueParsing := true
	for continueParsing {
		item := lex.getNextItem()
		switch item.typ {
		case itemError:
			continueParsing = false
			err = fmt.Errorf("error parsing: \"%s\" #line %d:%d", item.val, item.line, item.col)
		case itemVariableName:
			outGo.Append(NewInstanceVariableNamed(item.val))
		case itemMapKey:
			outGo.Append(NewMapKey(item.val))
		case itemArrayIndex:
			var val int64
			val, err = strconv.ParseInt(item.val, 10, 64)
			if err != nil {
				return
			}
			outGo.Append(NewArrayIndex(int(val)))
		case itemEOF:
			continueParsing = false
		}
	}
	lex.drain()
	return outGo, err
}

func (p goPath) Copy() PathMutator {
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

func (p *goPath) Each(yield func(index int, componenter Componenter)) {
	for index, part := range p.parts {
		yield(index, part)
	}
}
