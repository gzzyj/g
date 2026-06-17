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

package fancyindex

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gzzyj/g/internal/collector/internal"
	"github.com/gzzyj/g/internal/pkg/errs"
	httppkg "github.com/gzzyj/g/internal/pkg/http"
	"github.com/gzzyj/g/internal/version"
)

const (
	// Name Collector name
	Name = "fancyindex"
)

// Collector Nginx fancyindex collector
type Collector struct {
	url  string
	pURL *url.URL
	doc  *goquery.Document
}

// NewCollector Get the collector instance
func NewCollector(downloadPageURL string) (*Collector, error) {
	if downloadPageURL == "" {
		return nil, errs.ErrEmptyURL
	}

	pURL, err := url.Parse(downloadPageURL)
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

// StableVersions Return all stable versions
func (c *Collector) StableVersions() (items []*version.Version, err error) {
	return make([]*version.Version, 0), nil // Unable to determine which versions are stable
}

// UnstableVersions Return all stable versions
func (c *Collector) UnstableVersions() (items []*version.Version, err error) {
	return make([]*version.Version, 0), nil // Unable to determine which versions are unstable
}

// ArchivedVersions Return all archived versions
func (c *Collector) ArchivedVersions() (items []*version.Version, err error) {
	return make([]*version.Version, 0), nil // Unable to determine which versions are archived
}

// AllVersions Return all versions
func (c *Collector) AllVersions() (vers []*version.Version, err error) {
	items := c.findGoFileItems(c.doc.Find("table").First())
	if len(items) == 0 {
		return make([]*version.Version, 0), nil
	}
	if vers, err = internal.Convert2Versions(items); err != nil {
		return nil, err
	}
	return vers, nil
}

func (c *Collector) findGoFileItems(table *goquery.Selection) (items []*internal.GoFileItem) {
	trs := table.Find("tbody").Find("tr")
	items = make([]*internal.GoFileItem, 0, trs.Length())

	trs.Each(func(j int, tr *goquery.Selection) {
		tds := tr.Find("td")
		anchor := tds.Filter(".link").Find("a")

		href := anchor.AttrOr("href", "")
		if !strings.HasPrefix(href, "go") || strings.HasSuffix(href, "/") {
			return
		}

		items = append(items, &internal.GoFileItem{
			FileName: anchor.Text(),
			URL:      c.url + href,
			Size:     strings.TrimSpace(tds.Filter(".size").Text()),
		})
	})
	return items
}
