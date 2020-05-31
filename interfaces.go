package path

// Pather is an abstract path object
// Created to evaluate go Struct paths for a validator
type Pather interface {
	// IsEqual
	// @return true if paths identify the same resource, false if not
	IsEqual(Pather) bool
}

// Componenter is an abstract Path component
type Componenter interface {
	// IsEqual
	// @return true if the components identify the same object, assuming these are evaluated at the same location in the Path
	IsEqual(Componenter) bool
}
