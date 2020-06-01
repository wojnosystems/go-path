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
		//"empty": {
		//	input: "",
		//	expected: func() Pather {
		//		return NewRoot()
		//	},
		//},
		"structVar": {
			input: "variableName",
			expected: func() Pather {
				r := NewRoot()
				r.Append(NewInstanceVariableNamed("variableName"))
				return r
			},
		},
		//"arrayIndex": {
		//	input: "[4]",
		//	expected: func() Pather {
		//		r := NewRoot()
		//		r.Append(NewArrayIndex(4))
		//		return r
		//	},
		//},
	}

	for caseName, c := range cases {
		in := bytes.NewBufferString(c.input)
		actual, err := Parse(in)
		require.NoError(t, err, caseName)
		assert.True(t, c.expected().IsEqual(actual), caseName)
	}
}
