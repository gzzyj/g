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
	"io"
	"os"

	"github.com/urfave/cli/v2"
)

var envNames = []string{
	homeEnv,
	mirrorEnv,
	experimentalEnv,
}

func showEnv(ctx *cli.Context) (err error) {
	printEnv(os.Stdout, envNames)
	_, _ = fmt.Fprintf(os.Stdout, "%s=%q\n", toolsEnv, toolsDir)

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "# To add tools directory to PATH and block 'go install',")
	fmt.Fprintln(os.Stderr, "# add the following to your shell rc file (e.g. ~/.bashrc, ~/.zshrc):")
	fmt.Fprintln(os.Stderr, "#")
	fmt.Fprintf(os.Stderr, "# export %s=%q\n", toolsEnv, toolsDir)
	fmt.Fprintln(os.Stderr, "# export PATH=\"$PATH:$G_TOOLS\"")
	fmt.Fprintln(os.Stderr, "#")
	fmt.Fprintln(os.Stderr, "# go() {")
	fmt.Fprintln(os.Stderr, "#     if [[ \"$1\" == \"install\" ]]; then")
	fmt.Fprintln(os.Stderr, "#         echo \"Error: 'go install' is disabled. Use 'g toolchain add' to manage toolchains.\" >&2")
	fmt.Fprintln(os.Stderr, "#         return 1")
	fmt.Fprintln(os.Stderr, "#     fi")
	fmt.Fprintln(os.Stderr, "#     command go \"$@\"")
	fmt.Fprintln(os.Stderr, "# }")
	return nil
}

func printEnv(w io.Writer, envNames []string) {
	for _, eName := range envNames {
		_, _ = fmt.Fprintf(w, "%s=%q\n", eName, os.Getenv(eName))
	}
}
