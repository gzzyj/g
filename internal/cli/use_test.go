// Copyright (c) 2019 voidint <voidint@126.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGoDirective(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "normal go directive",
			input:    []byte("module github.com/gzzyj/g\n\ngo 1.20\n"),
			expected: "1.20",
		},
		{
			name:     "normal go directive",
			input:    []byte("module github.com/gzzyj/g\n\ngo 1.24.0\n"),
			expected: "1.24.0",
		},
		{
			name:     "normal go directive",
			input:    []byte("module github.com/gzzyj/g\n\ngo 1.24.4\n"),
			expected: "1.24.4",
		},
		{
			name:     "normal go directive",
			input:    []byte("module github.com/gzzyj/g\n\ngo 1.25rc1\n"),
			expected: "1.25rc1",
		},
		{
			name:     "no go directive",
			input:    []byte("module github.com/gzzyj/g\n"),
			expected: "",
		},
		{
			name:     "malformed directive",
			input:    []byte("go1.20"),
			expected: "",
		},
		{
			name:     "empty input",
			input:    []byte(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getGoDirective(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
