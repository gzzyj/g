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

package build

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	// ShortVersion short version number
	ShortVersion = "1.8.0"
)

// The value of variables come form `gb build -ldflags '-X "build.Built=xxxxx" -X "build.CommitID=xxxx"' `
var (
	// Built build time
	Built string
	// GitBranch current git branch
	GitBranch string
	// GitCommit git commit id
	GitCommit string
)

// Version returns the version information.
func Version() string {
	var buf strings.Builder
	buf.WriteString(ShortVersion)

	if Built != "" {
		buf.WriteString(fmt.Sprintf("\n%-15s%s", "Built:", Built))
	}
	if GitBranch != "" {
		buf.WriteString(fmt.Sprintf("\n%-15s%s", "Git branch:", GitBranch))
	}
	if GitCommit != "" {
		buf.WriteString(fmt.Sprintf("\n%-15s%s", "Git commit:", GitCommit))
	}
	buf.WriteString(fmt.Sprintf("\n%-15s%s", "Go version:", runtime.Version()))
	buf.WriteString(fmt.Sprintf("\n%-15s%s/%s", "OS/Arch:", runtime.GOOS, runtime.GOARCH))
	buf.WriteString(fmt.Sprintf("\n%-15s%t", "Experimental:", strings.EqualFold(os.Getenv("G_EXPERIMENTAL"), "true")))
	return buf.String()
}
