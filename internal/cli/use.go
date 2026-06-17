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
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func use(ctx *cli.Context) error {
	vname := ctx.Args().First()
	if vname == "" {
		// Uses go.mod if available and version is omitted
		goModData, err := os.ReadFile("go.mod")
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cli.Exit(wrapstring("No go.mod file found"), 1)
			}
			return cli.Exit(errstring(err), 1)
		}

		goDirective := getGoDirective(goModData)
		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		if goDirective == "" {
			return cli.Exit(wrapstring(fmt.Sprintf("Go directive does not exist in %q files", filepath.Join(wd, "go.mod"))), 1)
		}
		fmt.Printf("Found %q with version <%s>\n", filepath.Join(wd, "go.mod"), goDirective)
		vname = goDirective
	}

	versions, err := listLocalVersions(versionsDir)
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	// Try to match the version number strictly first
	for i := range versions {
		if versions[i].Name() != vname {
			continue
		}
		if err = switchVersion(versions[i].Name()); err != nil {
			return cli.Exit(errstring(err), 1)
		}
		return nil
	}

	// Try fuzzy matching the version number again
	cs, err := semver.NewConstraint(vname)
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	for j := len(versions) - 1; j >= 0; j-- {
		if !versions[j].MatchConstraint(cs) {
			continue
		}
		if err = switchVersion(versions[j].Name()); err != nil {
			return cli.Exit(errstring(err), 1)
		}
		return nil
	}

	return cli.Exit(wrapstring(fmt.Sprintf("The %q version does not exist, please install it first.", vname)), 1)
}

var goDirectiveReg = regexp.MustCompile(`(?m)^go\s+(\d+\.\d+(?:\.\d+)?(?:beta\d+|rc\d+)?)\s*(?:$|//.*)`)

// getGoDirective Extract the go directive from the go.mod file.
func getGoDirective(goModData []byte) string {
	// https://go.dev/ref/mod#go-mod-file-go
	match := goDirectiveReg.FindStringSubmatch(string(goModData))
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
