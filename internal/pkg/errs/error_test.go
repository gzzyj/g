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

package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageNotFoundError(t *testing.T) {
	t.Run("软件包不存在错误", func(t *testing.T) {
		kind := "Archive"
		goos := "linux"
		goarch := "amd64"

		err := NewPackageNotFoundError(kind, goos, goarch)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("package not found [%s,%s,%s]", goos, goarch, kind), err.Error())

		e, ok := err.(*PackageNotFoundError)
		assert.True(t, IsPackageNotFound(err))
		assert.False(t, IsPackageNotFound(nil))
		assert.True(t, ok)
		assert.NotNil(t, e)
	})
}

func TestVersionNotFoundError(t *testing.T) {
	t.Run("版本号不存在错误", func(t *testing.T) {
		v := "abcdef"
		goos := "linux"
		goarch := "amd64"

		err := NewVersionNotFoundError(v, goos, goarch)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("version not found %q [%s,%s]", v, goos, goarch), err.Error())

		e, ok := err.(*VersionNotFoundError)
		assert.True(t, IsVersionNotFound(err))
		assert.False(t, IsVersionNotFound(nil))
		assert.True(t, ok)
		assert.NotNil(t, e)
		assert.Equal(t, v, e.Version())
	})
}

func TestMalformedVersionError(t *testing.T) {
	t.Run("版本号格式错误", func(t *testing.T) {
		v := "abcdef"
		core := errors.New("malformed version string")
		err := NewMalformedVersionError(v, errors.New("malformed version string"))
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("malformed version string %q", v), err.Error())

		e, ok := err.(*MalformedVersionError)
		assert.True(t, IsMalformedVersion(err))
		assert.False(t, IsMalformedVersion(nil))
		assert.True(t, ok)
		assert.NotNil(t, e)
		assert.Equal(t, v, e.Version())
		assert.Equal(t, core, e.Unwrap())
	})
}

func TestURLUnreachableError(t *testing.T) {
	t.Run("URL不可达错误", func(t *testing.T) {
		url := "https://github.com/voidint"
		core := errors.New("hello error")

		err := NewURLUnreachableError(url, core)
		assert.NotNil(t, err)

		e, ok := err.(*URLUnreachableError)
		assert.True(t, IsURLUnreachable(err))
		assert.False(t, IsURLUnreachable(nil))
		assert.True(t, ok)
		assert.NotNil(t, e)
		assert.Equal(t, url, e.URL())
		assert.Equal(t, core, e.Unwrap())
		assert.Equal(t, fmt.Sprintf("URL %q is unreachable ==> %s", url, core.Error()), e.Error())
	})
}

func TestDownloadError(t *testing.T) {
	t.Run("安装包下载错误", func(t *testing.T) {
		url := "https://dl.google.com/go/go1.12.5.linux-amd64.tar.gz"
		core := errors.New("hello error")

		err := NewDownloadError(url, core)
		assert.NotNil(t, err)
		e, ok := err.(*DownloadError)
		assert.True(t, IsDownload(err))
		assert.False(t, IsDownload(nil))
		assert.True(t, ok)
		assert.NotNil(t, e)
		assert.Equal(t, url, e.URL())
		assert.Equal(t, core, e.Unwrap())
		assert.Equal(t, fmt.Sprintf("resource(%s) download failed ==> %s", url, core.Error()), e.Error())
	})
}
