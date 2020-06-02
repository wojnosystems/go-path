package go_path

import paths "github.com/wojnosystems/go-path"

type pathStructInstanceVariable struct {
	variableName string
}

func NewInstanceVariableNamed(variableName string) Componenter {
	return &pathStructInstanceVariable{
		variableName: variableName,
	}
}

func (p pathStructInstanceVariable) IsEqual(componenter paths.Componenter) bool {
	if componenter == nil {
		return false
	}
	if component, ok := componenter.(*pathStructInstanceVariable); !ok {
		return false
	} else {
		return p.variableName == component.variableName
	}
}

func (p pathStructInstanceVariable) String() string {
	return p.variableName
}
