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

package http

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/gzzyj/g/internal/build"
	"github.com/gzzyj/g/internal/pkg/errs"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

// Download saves the remote resource to local file with progress support.
func Download(srcURL string, filename string, flag int, perm fs.FileMode, withProgress bool) (size int64, err error) {
	req, err := http.NewRequest(http.MethodGet, srcURL, nil)
	if err != nil {
		return 0, errs.NewDownloadError(srcURL, err)
	}
	req.Header.Set("User-Agent", "g/"+build.ShortVersion) // Custom User-Agent avoids redirection issues when downloading from some mirrors
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, errs.NewDownloadError(srcURL, err)
	}
	defer resp.Body.Close()

	if !IsSuccess(resp.StatusCode) {
		return 0, errs.NewURLUnreachableError(srcURL, fmt.Errorf("%d", resp.StatusCode))
	}

	f, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		return 0, errs.NewDownloadError(srcURL, err)
	}
	defer f.Close()

	var dst io.Writer
	if withProgress {
		bar := progressbar.NewOptions64(
			resp.ContentLength,
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription("Downloading"),
			progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
			progressbar.OptionShowBytes(true),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				_, _ = fmt.Fprint(ansi.NewAnsiStdout(), "\n")
			}),
			// progressbar.OptionSpinnerType(35),
			// progressbar.OptionFullWidth(),
		)
		_ = bar.RenderBlank()
		dst = io.MultiWriter(f, bar)

	} else {
		dst = f
	}
	return io.Copy(dst, resp.Body)
}

// DownloadAsBytes fetches the resource and returns its raw byte content.
func DownloadAsBytes(srcURL string) (data []byte, err error) {
	resp, err := http.Get(srcURL)
	if err != nil {
		return nil, errs.NewDownloadError(srcURL, err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// IsSuccess determines if the HTTP status code indicates successful response.
func IsSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
