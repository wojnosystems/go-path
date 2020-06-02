package go_path

import paths "github.com/wojnosystems/go-path"

type pathMapInstanceVariable struct {
	variableName string
}

func NewMapKey(variableName string) Componenter {
	return &pathMapInstanceVariable{
		variableName: variableName,
	}
}

func (p pathMapInstanceVariable) IsEqual(componenter paths.Componenter) bool {
	if componenter == nil {
		return false
	}
	if component, ok := componenter.(*pathMapInstanceVariable); !ok {
		return false
	} else {
		return p.variableName == component.variableName
	}
}

func (p pathMapInstanceVariable) String() string {
	return "[\"" + p.variableName + "\"]"
}
