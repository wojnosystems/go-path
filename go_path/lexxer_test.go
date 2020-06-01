package go_path

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	cases := map[string]struct {
		input    string
		expected func() Pather
	}{
		"empty": {
			input: "",
			expected: func() Pather {
				return NewRoot()
			},
		},
		"structVar": {
			input: "variableName",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewInstanceVariableNamed("variableName"))
				return r
			},
		},
		"arrayIndex": {
			input: "[4]",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewArrayIndex(4))
				return r
			},
		},
		"mapKey": {
			input: "[\"test\"]",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewMapKey("test"))
				return r
			},
		},
		"multiple variables": {
			input: "var1.var2.var3",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewInstanceVariableNamed("var1"))
				r.Append(NewInstanceVariableNamed("var2"))
				r.Append(NewInstanceVariableNamed("var3"))
				return r
			},
		},
		"multiple arrays": {
			input: "[1][2][3]",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewArrayIndex(1))
				r.Append(NewArrayIndex(2))
				r.Append(NewArrayIndex(3))
				return r
			},
		},
		"multiple maps": {
			input: "[\"1\"][\"2\"][\"3\"]",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewMapKey("1"))
				r.Append(NewMapKey("2"))
				r.Append(NewMapKey("3"))
				return r
			},
		},
		"var -> map": {
			input: "var[\"x\"]",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewInstanceVariableNamed("var"))
				r.Append(NewMapKey("x"))
				return r
			},
		},
		"var -> array": {
			input: "var[4]",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewInstanceVariableNamed("var"))
				r.Append(NewArrayIndex(4))
				return r
			},
		},
		"array -> var": {
			input: "[4].var",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewArrayIndex(4))
				r.Append(NewInstanceVariableNamed("var"))
				return r
			},
		},
		"map -> var": {
			input: "[\"x\"].var",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewMapKey("x"))
				r.Append(NewInstanceVariableNamed("var"))
				return r
			},
		},
		"complex": {
			input: "dogs[5].attributes[\"fur\"].color",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewInstanceVariableNamed("dogs"))
				r.Append(NewArrayIndex(5))
				r.Append(NewInstanceVariableNamed("attributes"))
				r.Append(NewMapKey("fur"))
				r.Append(NewInstanceVariableNamed("color"))
				return r
			},
		},
	}

	for caseName, c := range cases {
		in := bytes.NewBufferString(c.input)
		actual, err := Parse(in)
		require.NoError(t, err, caseName)
		assert.True(t, c.expected().IsEqual(actual), caseName)
	}
}
