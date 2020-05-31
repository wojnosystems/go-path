package go_path

import (
	"fmt"
	paths "github.com/wojnosystems/go-path"
)

type pathArrayInstanceVariable struct {
	index int
}

func NewArrayIndex(index int) Componenter {
	return &pathArrayInstanceVariable{
		index: index,
	}
}

func (p pathArrayInstanceVariable) IsEqual(componenter paths.Componenter) bool {
	if componenter == nil {
		return false
	}
	if component, ok := componenter.(*pathArrayInstanceVariable); !ok {
		return false
	} else {
		return p.index == component.index
	}
}

func (p pathArrayInstanceVariable) Serialize() string {
	return fmt.Sprintf("[%d]", p.index)
}
