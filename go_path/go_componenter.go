package go_path

import paths "github.com/wojnosystems/go-path"

// Componenter is a go Struct-specific component
type Componenter interface {
	// Inherit everything about generic Components
	paths.Componenter
	// string debugging
	Serialize() string
}
