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

package version

import (
	"runtime"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/gzzyj/g/internal/pkg/errs"
)

// Finder implements version lookup for Go language distributions.
type Finder struct {
	kind   PackageKind
	goos   string
	goarch string
	items  []*Version
}

// WithFinderPackageKind sets the package kind to search for.
func WithFinderPackageKind(kind PackageKind) func(fdr *Finder) {
	return func(fdr *Finder) {
		fdr.kind = kind
	}
}

// WithFinderGoos sets target operating system (e.g. darwin, freebsd, linux).
func WithFinderGoos(goos string) func(fdr *Finder) {
	return func(fdr *Finder) {
		fdr.goos = goos
	}
}

// WithFinderGoarch sets target machine architecture (e.g. 386, amd64, arm, s390x).
func WithFinderGoarch(goarch string) func(fdr *Finder) {
	return func(fdr *Finder) {
		fdr.goarch = goarch
	}
}

// NewFinder creates a new Finder instance with sorted versions and applied options.
func NewFinder(items []*Version, opts ...func(fdr *Finder)) *Finder {
	sort.Sort(Collection(items)) // Sort in ascending order.

	fdr := Finder{
		kind:   ArchiveKind,
		goos:   runtime.GOOS,
		goarch: runtime.GOARCH,
		items:  items,
	}

	for _, setter := range opts {
		if opts != nil {
			setter(&fdr)
		}
	}

	return &fdr
}

// Find returns semantic version matching criteria.
//
//	Supported patterns:
//	1. Specific version (e.g. '1.21.4')
//	2. Latest version identifier 'latest'
//	3. Wildcards (e.g. '1.21.x', '1.x', '1.18.*')
//	4. Caret ranges for minor version compatibility (e.g. '^1', '^1.18', '^1.18.10')
//	5. Tilde ranges for patch version updates (e.g. '~1.18')
//	6. Greater than comparisons (e.g. '>1.18')
//	7. Less than comparisons (e.g. '<1.16')
//	8. Version ranges (e.g. '1.18-1.20')
func (fdr *Finder) Find(vname string) (*Version, error) {
	if vname == Latest {
		return fdr.findLatest()
	}

	for i := len(fdr.items) - 1; i >= 0; i-- {
		if fdr.items[i].name == vname && fdr.items[i].match(fdr.goos, fdr.goarch) {
			return fdr.items[i], nil
		}
	}

	cs, err := semver.NewConstraint(vname)
	if err != nil {
		return nil, errs.NewVersionNotFoundError(vname, fdr.goos, fdr.goarch)
	}

	versionFound := false
	for i := len(fdr.items) - 1; i >= 0; i-- { // Prefer higher versions first.
		if cs.Check(fdr.items[i].sv) {
			versionFound = true

			if fdr.items[i].match(fdr.goos, fdr.goarch) {
				return fdr.items[i], nil
			}
		}
	}
	if versionFound {
		return nil, errs.NewPackageNotFoundError(string(fdr.kind), fdr.goos, fdr.goarch)
	}
	return nil, errs.NewVersionNotFoundError(vname, fdr.goos, fdr.goarch)
}

// MustFind returns matched version or panics on error.
func (fdr *Finder) MustFind(vname string) *Version {
	v, err := fdr.Find(vname)
	if err != nil {
		panic(err)
	}
	return v
}

// Latest represents the current stable release.
const Latest = "latest"

func (fdr *Finder) findLatest() (*Version, error) {
	if len(fdr.items) == 0 {
		return nil, errs.NewVersionNotFoundError(Latest, fdr.goos, fdr.goarch)
	}

	for i := len(fdr.items) - 1; i >= 0; i-- {
		if fdr.items[i].match(fdr.goos, fdr.goarch) {
			return fdr.items[i], nil
		}
	}
	return nil, errs.NewPackageNotFoundError(string(fdr.kind), fdr.goos, fdr.goarch)
}
