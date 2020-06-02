package go_path

import paths "github.com/wojnosystems/go-path"

type Eacher interface {
	Each(func(index int, componenter Componenter))
}

// Pather is an abstract go Struct-path
// This is a way of identifying go variables or structs or maps
type Pather interface {
	// Inherit everything about generic paths
	paths.Pather
	Eacher
	// Allow copies to be made
	Copy() PathMutator
}

type Appender interface {
	// Append a component to the end of the path
	Append(...Componenter)
}

type Popper interface {
	Pop(count uint)
}

type Prepender interface {
	Prepend(components ...Componenter)
}

type PopFronter interface {
	PopFront(count uint)
}

type PathMutator interface {
	Pather
	Appender
	Popper
	Prepender
	PopFronter
}
