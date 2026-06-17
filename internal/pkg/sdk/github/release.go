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

package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/gzzyj/g/internal/go-update"
	"github.com/gzzyj/g/internal/pkg/checksum"
	"github.com/gzzyj/g/internal/pkg/errs"
	httppkg "github.com/gzzyj/g/internal/pkg/http"
	"github.com/mholt/archiver/v3"
)

// Release represents a software version release.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset contains downloadable resource files.
type Asset struct {
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// IsCompressedFile checks if the file is in compressed format.
func (a Asset) IsCompressedFile() bool {
	return a.ContentType == "application/zip" || a.ContentType == "application/x-gzip"
}

// ReleaseUpdater handles version update checks and operations.
type ReleaseUpdater struct {
}

// NewReleaseUpdater creates a release update handler instance.
func NewReleaseUpdater() *ReleaseUpdater {
	return new(ReleaseUpdater)
}

// CheckForUpdates verifies if newer version exists.
func (up ReleaseUpdater) CheckForUpdates(current *semver.Version, owner, repo string) (rel *Release, yes bool, err error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if !httppkg.IsSuccess(resp.StatusCode) {
		return nil, false, errs.NewURLUnreachableError(url, fmt.Errorf("%d", resp.StatusCode))
	}

	var latest Release
	if err = json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return nil, false, err
	}

	latestVersion, err := semver.NewVersion(latest.TagName)
	if err != nil {
		return nil, false, err
	}
	if latestVersion.GreaterThan(current) {
		return &latest, true, nil
	}
	return nil, false, nil
}

// Apply performs version update to specified release.
func (up ReleaseUpdater) Apply(rel *Release,
	findAsset func([]Asset) (idx int),
	findChecksum func([]Asset) (algo checksum.Algorithm, expectedChecksum string, err error),
) error {
	// findDownloadLink locates asset download URL.
	idx := findAsset(rel.Assets)
	if idx < 0 {
		return errs.ErrAssetNotFound
	}

	// findChecksum verifies file integrity hash.
	algo, expectedChecksum, err := findChecksum(rel.Assets)
	if err != nil {
		return err
	}

	// downloadFile fetches remote resource.
	tmpDir, err := os.MkdirTemp("", strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	url := rel.Assets[idx].BrowserDownloadURL
	srcFilename := filepath.Join(tmpDir, filepath.Base(url))
	dstFilename := srcFilename
	if _, err = httppkg.Download(url, srcFilename, os.O_WRONLY|os.O_CREATE, 0644, true); err != nil {
		return err
	}

	// verifyChecksum validates file hash.
	fmt.Println("Computing checksum with", algo)
	if err = checksum.VerifyFile(algo, expectedChecksum, srcFilename); err != nil {
		return err
	}
	fmt.Println("Checksums matched")

	// extractFile handles archive decompression.
	if rel.Assets[idx].IsCompressedFile() {
		if dstFilename, err = up.unarchive(srcFilename, tmpDir); err != nil {
			return err
		}
	}

	// updateBinary replaces old executable.
	dstFile, err := os.Open(dstFilename)
	if err != nil {
		return nil
	}
	defer dstFile.Close()
	return update.Apply(dstFile, update.Options{})
}

// unarchive extracts compressed files to target directory and returns first extracted file.
func (up ReleaseUpdater) unarchive(srcFile, dstDir string) (dstFile string, err error) {
	if err = archiver.Unarchive(srcFile, dstDir); err != nil {
		return "", err
	}
	// locateTargetFile finds the main executable after extraction.
	fis, _ := os.ReadDir(dstDir)
	for _, fi := range fis {
		if strings.HasSuffix(srcFile, fi.Name()) {
			continue
		}
		return filepath.Join(dstDir, fi.Name()), nil
	}
	return "", nil
}
