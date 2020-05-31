package go_path

import (
	"github.com/stretchr/testify/assert"
	"github.com/wojnosystems/go-path"
	"testing"
)

func TestEqual(t *testing.T) {
	compareTo := NewRoot()
	compareTo.Append(NewInstanceVariableNamed("dogs"), NewArrayIndex(4), NewMapKey("color"))
	cases := map[string]struct {
		input    func() path.Pather
		expected bool
	}{
		"root": {
			input: func() path.Pather {
				return NewRoot()
			},
		},
		"equal": {
			input: func() path.Pather {
				root := NewRoot()
				root.Append(NewInstanceVariableNamed("dogs"),
					NewArrayIndex(4),
					NewMapKey("color"))
				return root
			},
			expected: true,
		},
		"more parts": {
			input: func() path.Pather {
				root := NewRoot()
				root.Append(NewInstanceVariableNamed("dogs"),
					NewArrayIndex(4),
					NewMapKey("color"),
					NewInstanceVariableNamed("hue"))
				return root
			},
		},
		"fewer parts": {
			input: func() path.Pather {
				root := NewRoot()
				root.Append(NewInstanceVariableNamed("dogs"),
					NewArrayIndex(4))
				return root
			},
		},
		"different part": {
			input: func() path.Pather {
				root := NewRoot()
				root.Append(NewMapKey("dogs"),
					NewArrayIndex(4),
					NewMapKey("color"))
				return root
			},
		},
	}

	for caseName, c := range cases {
		actual := c.input()
		assert.Equal(t, c.expected, compareTo.IsEqual(actual), caseName)
	}
}

func TestGoPath_Copy(t *testing.T) {
	cases := map[string]struct {
		input func() Pather
	}{
		"empty": {
			input: func() Pather {
				return NewRoot()
			},
		},
		"components": {
			input: func() Pather {
				p := NewRoot()
				p.Append(NewMapKey("test"))
				p.Append(NewMapKey("test2"))
				p.Append(NewMapKey("test3"))
				return p
			},
		},
	}

	for caseName, c := range cases {
		actual := c.input().Copy()
		assert.True(t, c.input().IsEqual(actual), caseName)
	}
}

func TestGoPath_Pop(t *testing.T) {
	cases := map[string]struct {
		input    func() PathMutator
		popCount uint
		expected func() Pather
	}{
		"zero": {
			input: func() PathMutator {
				r := NewRoot()
				r.Append(NewMapKey("test"))
				r.Append(NewMapKey("test2"))
				return r
			},
			popCount: 0,
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewMapKey("test"))
				r.Append(NewMapKey("test2"))
				return r
			},
		},
		"one": {
			input: func() PathMutator {
				r := NewRoot()
				r.Append(NewMapKey("test"))
				r.Append(NewMapKey("test2"))
				return r
			},
			popCount: 1,
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewMapKey("test"))
				return r
			},
		},
		"same as length": {
			input: func() PathMutator {
				r := NewRoot()
				r.Append(NewMapKey("test"))
				r.Append(NewMapKey("test2"))
				return r
			},
			popCount: 2,
			expected: func() Pather {
				return NewRoot()
			},
		},
		"greater than the length": {
			input: func() PathMutator {
				r := NewRoot()
				r.Append(NewMapKey("test"))
				r.Append(NewMapKey("test2"))
				return r
			},
			popCount: 5,
			expected: func() Pather {
				return NewRoot()
			},
		},
	}

	for caseName, c := range cases {
		actual := c.input()
		actual.Pop(c.popCount)
		assert.True(t, c.expected().IsEqual(actual), caseName)
	}
}

func TestGoPath_PrependStart(t *testing.T) {
	cases := map[string]struct {
		input    func() PathMutator
		prepend  func() []Componenter
		expected func() Pather
	}{
		"nothing": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
			prepend: func() []Componenter {
				return make([]Componenter, 0)
			},
			expected: func() Pather {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
		},
		"one": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
			prepend: func() []Componenter {
				p := make([]Componenter, 0)
				p = append(p, NewInstanceVariableNamed("jack"))
				return p
			},
			expected: func() Pather {
				p := NewRoot()
				p.Append(NewInstanceVariableNamed("jack"))
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
		},
		"one to blank": {
			input: func() PathMutator {
				p := NewRoot()
				return p
			},
			prepend: func() []Componenter {
				p := make([]Componenter, 0)
				p = append(p, NewInstanceVariableNamed("jack"))
				return p
			},
			expected: func() Pather {
				p := NewRoot()
				p.Append(NewInstanceVariableNamed("jack"))
				return p
			},
		},
		"two": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
			prepend: func() []Componenter {
				p := make([]Componenter, 0)
				p = append(p, NewInstanceVariableNamed("jack"))
				p = append(p, NewMapKey("jill"))
				return p
			},
			expected: func() Pather {
				p := NewRoot()
				p.Append(NewInstanceVariableNamed("jack"))
				p.Append(NewMapKey("jill"))
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
		},
	}

	for caseName, c := range cases {
		actual := c.input()
		actual.Prepend(c.prepend()...)
		assert.True(t, c.expected().IsEqual(actual), caseName)
	}
}

func TestGoPath_PopFront(t *testing.T) {
	cases := map[string]struct {
		input    func() PathMutator
		popCount uint
		expected func() Pather
	}{
		"zero": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				return p
			},
			popCount: 0,
			expected: func() Pather {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				return p
			},
		},
		"one": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
			popCount: 1,
			expected: func() Pather {
				p := NewRoot()
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
		},
		"two": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
			popCount: 2,
			expected: func() Pather {
				p := NewRoot()
				p.Append(NewArrayIndex(2))
				return p
			},
		},
		"same as length": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
			popCount: 3,
			expected: func() Pather {
				p := NewRoot()
				return p
			},
		},
		"more than length": {
			input: func() PathMutator {
				p := NewRoot()
				p.Append(NewArrayIndex(0))
				p.Append(NewArrayIndex(1))
				p.Append(NewArrayIndex(2))
				return p
			},
			popCount: 5,
			expected: func() Pather {
				p := NewRoot()
				return p
			},
		},
	}

	for caseName, c := range cases {
		actual := c.input()
		actual.PopFront(c.popCount)
		assert.True(t, c.expected().IsEqual(actual), caseName)
	}
}

func TestDeserialize(t *testing.T) {

}
