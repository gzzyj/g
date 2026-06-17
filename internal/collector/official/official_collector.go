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

package official

import (
	"fmt"
	"net/http"
	stdurl "net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gzzyj/g/internal/pkg/errs"
	httppkg "github.com/gzzyj/g/internal/pkg/http"
	"github.com/gzzyj/g/internal/version"
)

const (
	// Name Collector name
	Name = "official"
)

// Collector collects Go versions from official download page.
type Collector struct {
	url  string
	pURL *stdurl.URL
	doc  *goquery.Document
}

// NewCollector creates a new collector instance for official Go downloads.
func NewCollector(downloadPageURL string) (*Collector, error) {
	if downloadPageURL == "" {
		return nil, errs.ErrEmptyURL
	}

	pURL, err := stdurl.Parse(downloadPageURL)
	if err != nil {
		return nil, err
	}

	c := Collector{
		url:  downloadPageURL,
		pURL: pURL,
	}
	if err = c.loadDocument(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Name Collector name
func (c *Collector) Name() string {
	return Name
}

func (c *Collector) loadDocument() (err error) {
	resp, err := http.Get(c.url)
	if err != nil {
		return errs.NewURLUnreachableError(c.url, err)
	}
	defer resp.Body.Close()

	if !httppkg.IsSuccess(resp.StatusCode) {
		return errs.NewURLUnreachableError(c.url, fmt.Errorf("%d", resp.StatusCode))
	}

	c.doc, err = goquery.NewDocumentFromReader(resp.Body)
	return err
}

func (c *Collector) findPackages(table *goquery.Selection) (pkgs []*version.Package) {
	alg := strings.TrimSuffix(table.Find("thead").Find("th").Last().Text(), " Checksum")

	table.Find("tr").Not(".first").Each(func(j int, tr *goquery.Selection) {
		td := tr.Find("td")
		href := td.Eq(0).Find("a").AttrOr("href", "")
		if strings.HasPrefix(href, "/") { // relative paths
			href = fmt.Sprintf("%s://%s%s", c.pURL.Scheme, c.pURL.Host, href)
		}
		pkgs = append(pkgs, &version.Package{
			FileName:  td.Eq(0).Find("a").Text(),
			URL:       href,
			Kind:      version.PackageKind(td.Eq(1).Text()),
			OS:        td.Eq(2).Text(),
			Arch:      td.Eq(3).Text(),
			Size:      td.Eq(4).Text(),
			Checksum:  td.Eq(5).Text(),
			Algorithm: alg,
		})
	})
	return pkgs
}

// hasUnstableVersions checks if unstable versions exist in document.
func (c *Collector) hasUnstableVersions() bool {
	return c.doc.Find("#unstable").Length() > 0
}

// StableVersions retrieves all stable Go versions from official releases.
func (c *Collector) StableVersions() (items []*version.Version, err error) {
	var divs *goquery.Selection
	if c.hasUnstableVersions() {
		divs = c.doc.Find("#stable").NextUntil("#unstable")
	} else {
		divs = c.doc.Find("#stable").NextUntil("#archive")
	}

	divs.EachWithBreak(func(i int, div *goquery.Selection) bool {
		vname, ok := div.Attr("id")
		if !ok {
			return true
		}

		var v *version.Version
		if v, err = version.New(
			strings.TrimPrefix(vname, "go"),
			version.WithPackages(c.findPackages(div.Find("table").First())),
		); err != nil {
			return false
		}

		items = append(items, v)
		return true
	})

	if err != nil {
		return nil, err
	}
	sort.Sort(version.Collection(items))
	return items, nil
}

// UnstableVersions fetches pre-release and development builds of Go.
func (c *Collector) UnstableVersions() (items []*version.Version, err error) {
	c.doc.Find("#unstable").NextUntil("#archive").EachWithBreak(func(i int, div *goquery.Selection) bool {
		vname, ok := div.Attr("id")
		if !ok {
			return true
		}

		var v *version.Version
		if v, err = version.New(
			strings.TrimPrefix(vname, "go"),
			version.WithPackages(c.findPackages(div.Find("table").First())),
		); err != nil {
			return false
		}

		items = append(items, v)
		return true
	})

	if err != nil {
		return nil, err
	}
	sort.Sort(version.Collection(items))
	return items, nil
}

// ArchivedVersions provides historical Go versions no longer supported.
func (c *Collector) ArchivedVersions() (items []*version.Version, err error) {
	c.doc.Find("#archive").Find("div.toggle").EachWithBreak(func(i int, div *goquery.Selection) bool {
		vname, ok := div.Attr("id")
		if !ok {
			return true
		}

		var v *version.Version
		if v, err = version.New(
			strings.TrimPrefix(vname, "go"),
			version.WithPackages(c.findPackages(div.Find("table").First())),
		); err != nil {
			return false
		}

		items = append(items, v)
		return true
	})

	if err != nil {
		return nil, err
	}
	sort.Sort(version.Collection(items))
	return items, nil
}

// AllVersions returns all known Go versions including stable/unstable/archived.
func (c *Collector) AllVersions() (items []*version.Version, err error) {
	items, err = c.StableVersions()
	if err != nil {
		return nil, err
	}
	archives, err := c.ArchivedVersions()
	if err != nil {
		return nil, err
	}
	items = append(items, archives...)

	unstables, err := c.UnstableVersions()
	if err != nil {
		return nil, err
	}
	items = append(items, unstables...)
	sort.Sort(version.Collection(items))
	return items, nil
}
