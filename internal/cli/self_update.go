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
	"bufio"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/gzzyj/g/internal/build"
	"github.com/gzzyj/g/internal/pkg/checksum"
	"github.com/gzzyj/g/internal/pkg/errs"
	httppkg "github.com/gzzyj/g/internal/pkg/http"
	"github.com/gzzyj/g/internal/pkg/sdk/github"
	"github.com/urfave/cli/v2"
)

func selfUpdate(*cli.Context) (err error) {
	up := github.NewReleaseUpdater()

	// 检查更新
	latest, yes, err := up.CheckForUpdates(semver.MustParse(build.ShortVersion), "voidint", "g")
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}
	if !yes {
		fmt.Printf("You are up to date! g v%s is the latest version.\n", build.ShortVersion)
		return nil
	}
	fmt.Printf("A new version of g(%s) is available\n", latest.TagName)

	// 应用更新
	if err = up.Apply(latest, findAsset, findChecksum); err != nil {
		return cli.Exit(errstring(err), 1)
	}
	fmt.Println("Update completed")
	return nil
}

func findAsset(items []github.Asset) (idx int) {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	suffix := fmt.Sprintf("%s-%s.%s", runtime.GOOS, runtime.GOARCH, ext)
	for i := range items {
		if strings.HasSuffix(items[i].BrowserDownloadURL, suffix) {
			return i
		}
	}
	return -1
}

func findChecksum(items []github.Asset) (algo checksum.Algorithm, expectedChecksum string, err error) {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	suffix := fmt.Sprintf("%s-%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	var checksumFileURL string
	for i := range items {
		if items[i].Name == "sha256sum.txt" {
			checksumFileURL = items[i].BrowserDownloadURL
			break
		}
	}
	if checksumFileURL == "" {
		return checksum.SHA256, "", errs.ErrChecksumFileNotFound
	}

	resp, err := http.Get(checksumFileURL)
	if err != nil {
		return checksum.SHA256, "", err
	}
	defer resp.Body.Close()

	if !httppkg.IsSuccess(resp.StatusCode) {
		return "", "", errs.NewURLUnreachableError(checksumFileURL, fmt.Errorf("%d", resp.StatusCode))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasSuffix(line, suffix) {
			continue
		}
		return checksum.SHA256, strings.Fields(line)[0], nil
	}
	if err = scanner.Err(); err != nil {
		return checksum.SHA256, "", err
	}
	return checksum.SHA256, "", errs.ErrChecksumFileNotFound
}
