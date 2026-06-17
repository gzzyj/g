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
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/gzzyj/g/internal/collector"
	"github.com/gzzyj/g/internal/version"
	"github.com/k0kubun/go-ansi"
	"github.com/urfave/cli/v2"
)

const (
	stableChannel   = "stable"
	unstableChannel = "unstable"
	archivedChannel = "archived"
)

func listRemote(ctx *cli.Context) (err error) {
	vname := ctx.Args().First()

	var cs *semver.Constraints
	if vname != "" && vname != stableChannel && vname != unstableChannel && vname != archivedChannel && vname != version.Latest {
		if cs, err = semver.NewConstraint(vname); err != nil {
			return cli.Exit(errstring(err), 1)
		}
	}

	c, err := collector.NewCollector(strings.Split(os.Getenv(mirrorEnv), mirrorSep)...)
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	var vs []*version.Version
	switch vname {
	case stableChannel:
		vs, err = c.StableVersions()
	case unstableChannel:
		vs, err = c.UnstableVersions()
	case archivedChannel:
		vs, err = c.ArchivedVersions()
	default:
		vs, err = c.AllVersions()
		if err == nil {
			if vname == version.Latest {
				vs = []*version.Version{vs[len(vs)-1]}
			}

			if vname != "" && vname != version.Latest {
				var newVs []*version.Version
				for _, v := range vs {
					if v.MatchConstraint(cs) {
						newVs = append(newVs, v)
					}
				}
				vs = newVs
			}
		}

	}
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	var renderMode uint8
	switch ctx.String("output") {
	case "json":
		renderMode = jsonMode
	default:
		renderMode = textMode
	}

	render(renderMode, installed(), vs, ansi.NewAnsiStdout())
	return nil
}
