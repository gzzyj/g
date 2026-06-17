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
	"strings"
)

var (
	// ErrUnsupportedChecksumAlgorithm Unsupported checksum algorithm
	ErrUnsupportedChecksumAlgorithm = errors.New("unsupported checksum algorithm")
	// ErrChecksumNotMatched File checksum does not match the computed checksum
	ErrChecksumNotMatched = errors.New("file checksum does not match the computed checksum")
	// ErrChecksumFileNotFound Checksum file not found
	ErrChecksumFileNotFound = errors.New("checksum file not found")
	// ErrAssetNotFound Asset not found
	ErrAssetNotFound = errors.New("asset not found")
	// ErrCollectorNotFound Collector not found
	ErrCollectorNotFound = errors.New("collector not found")
	// ErrEmptyURL URL is empty
	ErrEmptyURL = errors.New("empty url")
)

// PackageNotFoundError indicates the requested package does not exist.
type PackageNotFoundError struct {
	kind   string
	goos   string
	goarch string
}

// IsPackageNotFound checks if the error indicates missing package.
func IsPackageNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*PackageNotFoundError)
	return ok
}

// NewPackageNotFoundError creates a package missing error instance.
func NewPackageNotFoundError(kind, goos, goarch string) error {
	return &PackageNotFoundError{
		kind:   kind,
		goos:   goos,
		goarch: goarch,
	}
}

// Error returns detailed error message.
func (e PackageNotFoundError) Error() string {
	return fmt.Sprintf("package not found [%s,%s,%s]", e.goos, e.goarch, e.kind)
}

// VersionNotFoundError indicates the specified version is unavailable.
type VersionNotFoundError struct {
	version string
	goos    string
	goarch  string
}

// IsVersionNotFound checks if the error indicates missing version.
func IsVersionNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*VersionNotFoundError)
	return ok
}

// NewVersionNotFoundError creates a version missing error instance.
func NewVersionNotFoundError(version, goos, goarch string) error {
	return &VersionNotFoundError{
		version: version,
		goos:    goos,
		goarch:  goarch,
	}
}

// Error returns detailed error message.
func (e VersionNotFoundError) Error() string {
	return fmt.Sprintf("version not found %q [%s,%s]", e.version, e.goos, e.goarch)
}

// Version returns the semantic version string.
func (e VersionNotFoundError) Version() string {
	return e.version
}

// MalformedVersionError indicates invalid version format.
type MalformedVersionError struct {
	err     error
	version string
}

// IsMalformedVersion checks if the error indicates invalid version syntax.
func IsMalformedVersion(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*MalformedVersionError)
	return ok
}

// NewMalformedVersionError creates malformed version error instance.
func NewMalformedVersionError(version string, err error) error {
	return &MalformedVersionError{
		err:     err,
		version: version,
	}
}

// Error returns detailed error message.
func (e MalformedVersionError) Error() string {
	return fmt.Sprintf("malformed version string %q", e.version)
}

// Unwrap returns the original error object.
func (e MalformedVersionError) Unwrap() error {
	return e.err
}

// Version returns the semantic version string.
func (e MalformedVersionError) Version() string {
	return e.version
}

// URLUnreachableError indicates failure to access the remote resource.
type URLUnreachableError struct {
	err error
	url string
}

// IsURLUnreachable checks if the error indicates network unreachable.
func IsURLUnreachable(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*URLUnreachableError)
	return ok
}

// NewURLUnreachableError creates a URL unreachable error instance.
func NewURLUnreachableError(url string, err error) error {
	return &URLUnreachableError{
		err: err,
		url: url,
	}
}

// Error returns detailed error message.
func (e URLUnreachableError) Error() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("URL %q is unreachable", e.url))
	if e.err != nil {
		buf.WriteString(" ==> " + e.err.Error())
	}
	return buf.String()
}

// Unwrap returns the original error object.
func (e URLUnreachableError) Unwrap() error {
	return e.err
}

// URL returns the resource location URL.
func (e URLUnreachableError) URL() string {
	return e.url
}

// DownloadError indicates failure during file download process.
type DownloadError struct {
	url string
	err error
}

// IsDownload checks if the error occurred during download operation.
func IsDownload(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*DownloadError)
	return ok
}

// NewDownloadError creates a download failure error instance.
func NewDownloadError(url string, err error) error {
	return &DownloadError{
		url: url,
		err: err,
	}
}

// Error returns detailed error message.
func (e DownloadError) Error() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("resource(%s) download failed", e.url))
	if e.err != nil {
		buf.WriteString(" ==> " + e.err.Error())
	}
	return buf.String()
}

// Unwrap returns the wrapped error.
func (e DownloadError) Unwrap() error {
	return e.err
}

// URL returns the resource location.
func (e DownloadError) URL() string {
	return e.url
}
