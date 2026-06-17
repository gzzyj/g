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

package internal

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/gzzyj/g/internal/pkg/checksum"
	"github.com/gzzyj/g/internal/version"
)

type GoFileItem struct {
	FileName string
	URL      string
	Size     string
}

func (item GoFileItem) getGoVersion() string {
	arr := strings.Split(strings.TrimPrefix(item.FileName, "go"), ".")
	if len(arr) < 3 || !unicode.IsNumber(rune(arr[0][0])) {
		return ""
	}
	if unicode.IsNumber(rune(arr[2][0])) {
		return fmt.Sprintf("%s.%s.%s", arr[0], arr[1], arr[2])
	}

	if unicode.IsNumber(rune(arr[1][0])) {
		return fmt.Sprintf("%s.%s", arr[0], arr[1])
	}
	return arr[0]
}

func (item GoFileItem) isSHA256File() bool {
	return strings.HasSuffix(item.FileName, ".sha256")
}

func (item GoFileItem) isPackageFile() bool {
	return strings.HasSuffix(item.FileName, ".tar.gz") ||
		strings.HasSuffix(item.FileName, ".pkg") ||
		strings.HasSuffix(item.FileName, ".zip") ||
		strings.HasSuffix(item.FileName, ".msi")
}

func (item GoFileItem) getKind() version.PackageKind {
	if strings.HasSuffix(item.FileName, ".src.tar.gz") {
		return version.SourceKind
	}
	if strings.HasSuffix(item.FileName, ".tar.gz") || strings.HasSuffix(item.FileName, ".zip") {
		return version.ArchiveKind
	}
	if strings.HasSuffix(item.FileName, ".pkg") || strings.HasSuffix(item.FileName, ".msi") {
		return version.InstallerKind
	}
	return "Unknown"
}

var osMapping = map[string]string{
	"linux":     "Linux",
	"darwin":    "macOS",
	"windows":   "Windows",
	"freebsd":   "FreeBSD",
	"netbsd":    "netbsd",
	"openbsd":   "openbsd",
	"solaris":   "solaris",
	"plan9":     "plan9",
	"aix":       "aix",
	"dragonfly": "dragonfly",
	"illumos":   "illumos",
}

func (item GoFileItem) getOS() string {
	for k, v := range osMapping {
		if strings.Contains(item.FileName, k) {
			return v
		}
	}
	return ""
}

var archMapping = map[string]string{
	"-386.":      "x86",
	"-amd64.":    "x86-64",
	"-arm.":      "ARMv6",
	"-arm64.":    "ARM64",
	"-armv6l.":   "ARMv6",
	"-ppc64.":    "ppc64",
	"-ppc64le.":  "ppc64le",
	"-mips.":     "mips",
	"-mipsle.":   "mipsle",
	"-mips64.":   "mips64",
	"-mips64le.": "mips64le",
	"-s390x.":    "s390x",
	"-riscv64.":  "riscv64",
	"-loong64.":  "loong64",
}

func (item GoFileItem) getArch() string {
	for k, v := range archMapping {
		if strings.Contains(item.FileName, k) {
			return v
		}
	}
	return ""
}

func Convert2Versions(items []*GoFileItem) (vers []*version.Version, err error) {
	pkgMap := make(map[string][]*version.Package, 20)

	for _, pitem := range items {
		ver := pitem.getGoVersion()
		if _, ok := pkgMap[ver]; !ok {
			pkgMap[ver] = make([]*version.Package, 0, 20)
		}

		if pitem.isPackageFile() {
			pkgMap[ver] = append(pkgMap[ver], &version.Package{
				FileName: pitem.FileName,
				URL:      pitem.URL,
				Kind:     pitem.getKind(),
				OS:       pitem.getOS(),
				Arch:     pitem.getArch(),
				Size:     pitem.Size,
			})
		} else if pitem.isSHA256File() {
			// Set checksum and hashing algorithm.
			for _, ppkg := range pkgMap[ver] {
				if !strings.HasPrefix(pitem.FileName, ppkg.FileName) {
					continue
				}
				ppkg.Algorithm = string(checksum.SHA256)
				ppkg.ChecksumURL = pitem.URL
			}
		}
	}

	vers = make([]*version.Version, 0, len(pkgMap))
	for vname, pkgs := range pkgMap {
		v, err := version.New(vname, version.WithPackages(pkgs))
		if err != nil {
			return nil, err
		}
		vers = append(vers, v)
	}
	sort.Sort(version.Collection(vers))
	return vers, nil
}
