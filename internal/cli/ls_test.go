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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListLocalVersions(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.Mkdir(filepath.Join(tmpDir, "1.25rc1"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "1.24rc3"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "1.18beta2"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "1.19beta1"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "1.24.4"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "1.20.14"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "1.20"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "invalid_version"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "not_a_directory"), []byte{}, 0644)

	tests := []struct {
		name     string
		dirPath  string
		vnames   []string
		checkErr func(error) bool
	}{
		{"Directory does not exist", "/not/exist/path", nil, func(err error) bool {
			var pathError *fs.PathError
			return errors.As(err, &pathError)
		}},

		{"The file path is not a directory", filepath.Join(tmpDir, "not_a_directory"), nil, func(err error) bool {
			var pathError *fs.PathError
			return errors.As(err, &pathError)
		}},

		{"正常版本目录", tmpDir, []string{"1.18beta2", "1.19beta1", "1.20", "1.20.14", "1.24rc3", "1.24.4", "1.25rc1"}, func(err error) bool {
			return err == nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listLocalVersions(tt.dirPath)
			assert.True(t, tt.checkErr(err))

			var actualNames []string
			for _, v := range got {
				actualNames = append(actualNames, v.Name())
			}
			assert.Equal(t, tt.vnames, actualNames)
		})
	}
}
