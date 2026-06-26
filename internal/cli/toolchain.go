// Copyright (c) 2024 voidint <voidint@126.com>
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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/Masterminds/semver/v3"
	"github.com/dixonwille/wlog/v3"
	"github.com/dixonwille/wmenu/v5"
	"github.com/urfave/cli/v2"
)

// ToolchainEntry records a single managed toolchain.
type ToolchainEntry struct {
	Name               string    `json:"name"`
	ModulePath         string    `json:"module_path,omitempty"`
	InstallPath        string    `json:"install_path"`
	Version            string    `json:"version"`
	InstallTime        time.Time `json:"install_time"`
	GoVersionDependent bool      `json:"go_version_dependent"`
	GoVersion          string    `json:"go_version,omitempty"`
}

// ToolAlias maps a short tool name to its full Go module path.
type ToolAlias struct {
	Name   string `json:"name"`
	Module string `json:"module"`
}

// ToolchainConfig is the on-disk format for all managed toolchains.
type ToolchainConfig struct {
	Toolchains []ToolchainEntry `json:"toolchains"`
	KnownTools []ToolAlias      `json:"known_tools,omitempty"`
}

func toolchainCfgPath() string {
	return filepath.Join(ghomeDir, "toolchain.json")
}

func loadToolchainCfg() (*ToolchainConfig, error) {
	p := toolchainCfgPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &ToolchainConfig{Toolchains: []ToolchainEntry{}}, nil
		}
		return nil, err
	}
	var cfg ToolchainConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Toolchains == nil {
		cfg.Toolchains = []ToolchainEntry{}
	}
	if cfg.KnownTools == nil {
		cfg.KnownTools = []ToolAlias{}
	}
	return &cfg, nil
}

func saveToolchainCfg(cfg *ToolchainConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(toolchainCfgPath(), data, 0644)
}

// findToolchain returns pointer to entry with given name, or nil.
func findToolchain(cfg *ToolchainConfig, name string) *ToolchainEntry {
	for i := range cfg.Toolchains {
		if cfg.Toolchains[i].Name == name {
			return &cfg.Toolchains[i]
		}
	}
	return nil
}

// copyFile copies a regular file from src to dst, creating parent directories.
func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

// moveToolchainBin moves the binary between tools/ and versions/<gover>/bin/,
// updating the entry fields accordingly.
func moveToolchainBin(entry *ToolchainEntry, toFollow bool) error {
	var src, dst string
	currentVer := inuse(goroot)

	if toFollow {
		if currentVer == "" {
			return fmt.Errorf("no Go version is currently in use")
		}
		src = entry.InstallPath
		dst = filepath.Join(versionsDir, currentVer, "bin", entry.Name)
		entry.GoVersion = currentVer
	} else {
		src = entry.InstallPath
		dst = filepath.Join(toolsDir, entry.Name)
		entry.GoVersion = ""
	}

	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source binary %q not found: %w", src, err)
	}

	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	if err := os.Remove(src); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("remove %s: %w", src, err)
	}

	entry.GoVersionDependent = toFollow
	entry.InstallPath = dst
	return nil
}

// ---------------------------------------------------------------------------
// Tool → module path resolution
// ---------------------------------------------------------------------------

// knownToolModules maps common tool names to their Go module paths.
var knownToolModules = map[string]string{
	"gopls":              "golang.org/x/tools/gopls",
	"goimports":          "golang.org/x/tools/cmd/goimports",
	"gofumpt":            "mvdan.cc/gofumpt",
	"staticcheck":        "honnef.co/go/tools/cmd/staticcheck",
	"golangci-lint":      "github.com/golangci/golangci-lint/cmd/golangci-lint",
	"dlv":                "github.com/go-delve/delve/cmd/dlv",
	"wire":               "github.com/google/wire/cmd/wire",
	"stringer":           "golang.org/x/tools/cmd/stringer",
	"impl":               "github.com/josharian/impl",
	"mockgen":            "github.com/uber-go/mock/cmd/mockgen",
	"gosec":              "github.com/securego/gosec/v2/cmd/gosec",
	"protoc-gen-go":      "google.golang.org/protobuf/cmd/protoc-gen-go",
	"protoc-gen-go-grpc": "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
	"protoc-gen-grpc-gateway":      "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway",
	"protoc-gen-openapiv2":         "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2",
	"gotests":            "github.com/cweill/gotests/gotests",
	"richgo":             "github.com/kyoh86/richgo",
	"goreleaser":         "github.com/goreleaser/goreleaser/v2",
}

// resolveModulePath looks up a tool name and returns its Go module path.
// Checks the built-in known-tools map first, then falls back to user-defined
// aliases in the config file (config overrides built-in with the same name).
func resolveModulePath(name string) (modulePath string, known bool) {
	// 1. Check built-in defaults
	if mp, ok := knownToolModules[name]; ok {
		return mp, true
	}
	// 2. Check user-defined aliases from config
	cfg, err := loadToolchainCfg()
	if err == nil {
		for _, a := range cfg.KnownTools {
			if a.Name == name && a.Module != "" {
				return a.Module, true
			}
		}
	}
	return "", false
}

// ---------------------------------------------------------------------------
// Module version query via Go module proxy
// ---------------------------------------------------------------------------

// moduleVersionsResp mirrors `go list -m -versions -json <module>` output.
type moduleVersionsResp struct {
	Path     string   `json:"path"`
	Versions []string `json:"versions"`
}

// listModuleVersions queries the Go module proxy for all available versions.
func listModuleVersions(modulePath string) ([]string, error) {
	goBin := filepath.Join(goroot, "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}

	// Resolve to module root: go list -m -versions requires a module root,
	// not a package path (e.g. github.com/user/repo not .../cmd/foo).
	// Walk from longest to shortest path until go list -m -versions succeeds.
	parts := strings.Split(modulePath, "/")
	var versions []string
	var lastErr error

	for i := len(parts); i >= 2; i-- {
		candidate := strings.Join(parts[:i], "/")
		if !strings.Contains(candidate, ".") {
			continue // must have a domain-like component
		}
		cmd := exec.Command(goBin, "list", "-mod=mod", "-m", "-versions", "-json", candidate)
		cmd.Dir = downloadsDir
		output, err := cmd.Output()
		if err != nil {
			lastErr = err
			continue
		}
		var resp moduleVersionsResp
		if err := json.Unmarshal(output, &resp); err != nil {
			lastErr = err
			continue
		}
		if len(resp.Versions) > 0 {
			versions = resp.Versions
			lastErr = nil
			break
		}
	}

	if lastErr != nil && len(versions) == 0 {
		if ee, ok := lastErr.(*exec.ExitError); ok {
			return nil, fmt.Errorf("query versions for %s: %s", modulePath, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("query versions for %s: %w", modulePath, lastErr)
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for %q", modulePath)
	}
	return versions, nil
}

// latestVersion picks the newest stable (non-prerelease) version from a list,
// falling back to the newest prerelease if none found.
func latestVersion(versions []string) (string, error) {
	var stable, prerelease []*semver.Version
	for _, v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		if sv.Prerelease() == "" {
			stable = append(stable, sv)
		} else {
			prerelease = append(prerelease, sv)
		}
	}

	switch {
	case len(stable) > 0:
		sort.Sort(sort.Reverse(semver.Collection(stable)))
		return stable[0].Original(), nil
	case len(prerelease) > 0:
		sort.Sort(sort.Reverse(semver.Collection(prerelease)))
		return prerelease[0].Original(), nil
	default:
		return "", fmt.Errorf("no valid semver versions found")
	}
}

// promptVersionSelect shows an interactive menu for the user to pick a version.
func promptVersionSelect(versions []string) (string, error) {
	// Sort descending (newest first)
	items := make([]*semver.Version, 0, len(versions))
	for _, v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		items = append(items, sv)
	}
	sort.Sort(sort.Reverse(semver.Collection(items)))

	menu := wmenu.NewMenu("Select the version you want to install:")
	menu.AddColor(
		wlog.Color{Code: ct.Green},
		wlog.Color{Code: ct.Yellow},
		wlog.Color{Code: ct.Magenta},
		wlog.Color{Code: ct.Yellow},
	)

	var selected string
	menu.Action(func(opts []wmenu.Opt) error {
		selected = opts[0].Value.(string)
		return nil
	})

	const maxOptions = 30
	for i, sv := range items {
		if i >= maxOptions {
			break
		}
		label := sv.Original()
		if i == 0 {
			label += " (latest)"
			menu.Option(label, sv.Original(), true, nil)
		} else {
			menu.Option(" "+label, sv.Original(), false, nil)
		}
	}

	if err := menu.Run(); err != nil {
		return "", err
	}
	return selected, nil
}

// resolveToolVersion resolves a tool name and optional version string,
// querying the module proxy as needed. Special values:
//
//   - "latest"  → auto-pick the newest non-prerelease
//   - ""        → prompt user to select from a menu
//   - anything else → use as-is (must be a valid Go module version)
func resolveToolVersion(toolName, version string) (modulePath, resolvedVersion string, err error) {
	var mp string
	var known bool
	if strings.Contains(toolName, "/") {
		// Full module path (e.g. golang.org/x/tools/gopls), use directly
		mp = toolName
		known = true
	} else {
		mp, known = resolveModulePath(toolName)
	}
	if !known {
		return "", "", fmt.Errorf("unknown tool %q, use --from <path> to register a custom binary, or add the module mapping to the known-tools list", toolName)
	}

	versions, err := listModuleVersions(mp)
	if err != nil {
		return "", "", fmt.Errorf("query versions for %q: %w", mp, err)
	}
	if len(versions) == 0 {
		return "", "", fmt.Errorf("no versions found for %q", mp)
	}

	switch {
	case version == "latest" || version == "" && len(versions) == 1:
		resolvedVersion, err = latestVersion(versions)
		if err != nil {
			return "", "", fmt.Errorf("resolve latest version for %q: %w", mp, err)
		}
	case version == "":
		resolvedVersion, err = promptVersionSelect(versions)
		if err != nil {
			return "", "", err
		}
		if resolvedVersion == "" {
			return "", "", fmt.Errorf("no version selected")
		}
	default:
		resolvedVersion = version
	}

	return mp, resolvedVersion, nil
}

// installModule runs `go install <module>@<version>` with GOBIN set to gobinDir.
func installModule(modulePath, version, gobinDir string) error {
	goBin := filepath.Join(goroot, "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}

	if err := os.MkdirAll(gobinDir, 0755); err != nil {
		return err
	}

	spec := fmt.Sprintf("%s@%s", modulePath, version)
	cmd := exec.Command(goBin, "install", "-mod=mod", spec)
	cmd.Dir = downloadsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOBIN=%s", gobinDir),
		"GO111MODULE=on",
	)

	fmt.Printf("Installing %s ...\n", spec)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install %s: %w", spec, err)
	}
	return nil
}

// buildInstallAction installs a toolchain from either a local binary path
// (--from) or automatically from the Go module proxy. Returns the entry
// fields and the install path on success.
func buildInstallAction(name, version, sourcePath string, goDependent bool, goVersion string) (*ToolchainEntry, error) {
	var (
		modulePath string
		installOK  bool
	)

	// Determine destination directory
	var destPath string
	if goDependent {
		if goVersion == "" {
			return nil, fmt.Errorf("no Go version in use; switch to a version first, or use --from to register a pre-built binary")
		}
		destPath = filepath.Join(versionsDir, goVersion, "bin", name)
	} else {
		destPath = filepath.Join(toolsDir, name)
	}

	if sourcePath != "" {
		// If sourcePath is a directory, join with tool name
		if fi, stErr := os.Stat(sourcePath); stErr == nil && fi.IsDir() {
			sourcePath = filepath.Join(sourcePath, name)
		}
		// Manual binary registration — copy to managed location
		if err := copyFile(sourcePath, destPath); err != nil {
			return nil, err
		}
		installOK = true
	} else {
		// Automatic module install
		mp, resolvedVersion, err := resolveToolVersion(name, version)
		if err != nil {
			return nil, err
		}
		modulePath = mp
		version = resolvedVersion

		if err := installModule(modulePath, version, filepath.Dir(destPath)); err != nil {
			return nil, err
		}
		installOK = true
	}

	if !installOK {
		return nil, fmt.Errorf("failed to install toolchain %s", name)
	}

	return &ToolchainEntry{
		Name:               name,
		ModulePath:         modulePath,
		InstallPath:        destPath,
		Version:            version,
		InstallTime:        time.Now(),
		GoVersionDependent: goDependent,
		GoVersion:          goVersion,
	}, nil
}

// hasArgFlag checks if a flag appears anywhere in os.Args.
// This works around urfave/cli not parsing flags after positional args.
func hasArgFlag(flag string) bool {
	for _, a := range os.Args {
		if a == flag {
			return true
		}
	}
	return false
}

// getArgFlagValue scans os.Args for "--flag value" and returns value.
func getArgFlagValue(flag string) string {
	for i, a := range os.Args {
		if a == flag && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// CLI command handlers
// ---------------------------------------------------------------------------

func toolchainAdd(ctx *cli.Context) error {
	arg := ctx.Args().First()
	if arg == "" {
		return cli.ShowSubcommandHelp(ctx)
	}

	name, ver := arg, ""
	if idx := strings.LastIndex(arg, "@"); idx > 0 {
		name = arg[:idx]
		ver = arg[idx+1:]
	}

	goDependent := ctx.Bool("go-dependent")
	if !goDependent {
		goDependent = hasArgFlag("--go-dependent")
	}
	sourcePath := ctx.String("from")
	if sourcePath == "" {
		sourcePath = getArgFlagValue("--from")
	}

	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	if findToolchain(cfg, name) != nil {
		return cli.Exit(wrapstring(fmt.Sprintf("toolchain %q already exists", name)), 1)
	}

	currentVer := inuse(goroot)
	if goDependent && currentVer == "" {
		return cli.Exit(wrapstring("no Go version in use; switch to a version first, or omit --go-dependent"), 1)
	}

	entry, err := buildInstallAction(name, ver, sourcePath, goDependent, currentVer)
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	cfg.Toolchains = append(cfg.Toolchains, *entry)

	if err := saveToolchainCfg(cfg); err != nil {
		_ = os.Remove(entry.InstallPath)
		return cli.Exit(errstring(err), 1)
	}

	fmt.Printf("Added toolchain %s@%s", entry.Name, entry.Version)
	if entry.GoVersionDependent && currentVer != "" {
		fmt.Printf(" (follows Go %s)", currentVer)
	}
	fmt.Println()
	return nil
}

func toolchainRemove(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		return cli.ShowSubcommandHelp(ctx)
	}

	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	entry := findToolchain(cfg, name)
	if entry == nil {
		return cli.Exit(wrapstring(fmt.Sprintf("toolchain %q not found", name)), 1)
	}

	if err := os.Remove(entry.InstallPath); err != nil && !os.IsNotExist(err) {
		return cli.Exit(errstring(err), 1)
	}

	filtered := make([]ToolchainEntry, 0, len(cfg.Toolchains)-1)
	for i := range cfg.Toolchains {
		if cfg.Toolchains[i].Name != name {
			filtered = append(filtered, cfg.Toolchains[i])
		}
	}
	cfg.Toolchains = filtered

	if err := saveToolchainCfg(cfg); err != nil {
		return cli.Exit(errstring(err), 1)
	}

	fmt.Printf("Removed toolchain %s\n", name)
	return nil
}

func toolchainList(ctx *cli.Context) error {
	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	if len(cfg.Toolchains) == 0 {
		fmt.Println("No toolchains registered")
		return nil
	}

	sort.Slice(cfg.Toolchains, func(i, j int) bool {
		return cfg.Toolchains[i].Name < cfg.Toolchains[j].Name
	})

	fmt.Printf("%-24s %-24s %-16s %-30s %-8s %s\n",
		"NAME", "MODULE", "VERSION", "INSTALL_PATH", "FOLLOWS", "GO_VERSION")
	fmt.Println(strings.Repeat("─", 120))
	for _, tc := range cfg.Toolchains {
		follows := "no"
		gv := "-"
		if tc.GoVersionDependent {
			follows = "yes"
			gv = tc.GoVersion
		}
		ver := tc.Version
		if ver == "" {
			ver = "-"
		}
		mp := tc.ModulePath
		if mp == "" {
			mp = "-"
		}
		fmt.Printf("%-24s %-24s %-16s %-30s %-8s %s\n",
			tc.Name, mp, ver, tc.InstallPath, follows, gv)
	}
	return nil
}

// ── Tool alias commands ──────────────────────────────────────────────

func toolchainAliasAdd(ctx *cli.Context) error {
	name := ctx.Args().Get(0)
	module := ctx.Args().Get(1)
	if name == "" || module == "" {
		return cli.Exit(wrapstring("usage: g toolchain alias add <name> <module>"), 1)
	}

	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	// Check for duplicate
	for i := range cfg.KnownTools {
		if cfg.KnownTools[i].Name == name {
			cfg.KnownTools[i].Module = module
			if err := saveToolchainCfg(cfg); err != nil {
				return cli.Exit(errstring(err), 1)
			}
			fmt.Printf("Updated alias %s → %s\n", name, module)
			return nil
		}
	}

	cfg.KnownTools = append(cfg.KnownTools, ToolAlias{Name: name, Module: module})
	if err := saveToolchainCfg(cfg); err != nil {
		return cli.Exit(errstring(err), 1)
	}
	fmt.Printf("Added alias %s → %s\n", name, module)
	return nil
}

func toolchainAliasRemove(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		return cli.Exit(wrapstring("usage: g toolchain alias remove <name>"), 1)
	}

	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	found := false
	filtered := make([]ToolAlias, 0, len(cfg.KnownTools))
	for i := range cfg.KnownTools {
		if cfg.KnownTools[i].Name == name {
			found = true
		} else {
			filtered = append(filtered, cfg.KnownTools[i])
		}
	}
	if !found {
		return cli.Exit(wrapstring(fmt.Sprintf("alias %q not found", name)), 1)
	}
	cfg.KnownTools = filtered

	if err := saveToolchainCfg(cfg); err != nil {
		return cli.Exit(errstring(err), 1)
	}
	fmt.Printf("Removed alias %s\n", name)
	return nil
}

func toolchainAliasList(ctx *cli.Context) error {
	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	// Merge built-in + config, config overrides
	all := make(map[string]string)
	for k, v := range knownToolModules {
		all[k] = v
	}
	for _, a := range cfg.KnownTools {
		all[a.Name] = a.Module // override
	}

	fmt.Printf("%-24s %s\n", "NAME", "MODULE_PATH")
	fmt.Println(strings.Repeat("─", 80))
	var names []string
	for k := range all {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		mark := ""
		for _, a := range cfg.KnownTools {
			if a.Name == n {
				mark = " (user)"
				break
			}
		}
		fmt.Printf("%-24s %s%s\n", n, all[n], mark)
	}
	return nil
}

func toolchainFollow(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		return cli.ShowSubcommandHelp(ctx)
	}

	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	entry := findToolchain(cfg, name)
	if entry == nil {
		return cli.Exit(wrapstring(fmt.Sprintf("toolchain %q not found", name)), 1)
	}
	if entry.GoVersionDependent {
		fmt.Printf("Toolchain %s already follows Go version\n", name)
		return nil
	}

	if err := moveToolchainBin(entry, true); err != nil {
		return cli.Exit(errstring(err), 1)
	}
	if err := saveToolchainCfg(cfg); err != nil {
		return cli.Exit(errstring(err), 1)
	}

	fmt.Printf("Toolchain %s now follows Go %s\n", name, entry.GoVersion)
	return nil
}

func toolchainUnfollow(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		return cli.ShowSubcommandHelp(ctx)
	}

	cfg, err := loadToolchainCfg()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	entry := findToolchain(cfg, name)
	if entry == nil {
		return cli.Exit(wrapstring(fmt.Sprintf("toolchain %q not found", name)), 1)
	}
	if !entry.GoVersionDependent {
		fmt.Printf("Toolchain %s already does not follow Go version\n", name)
		return nil
	}

	if err := moveToolchainBin(entry, false); err != nil {
		return cli.Exit(errstring(err), 1)
	}
	if err := saveToolchainCfg(cfg); err != nil {
		return cli.Exit(errstring(err), 1)
	}

	fmt.Printf("Toolchain %s no longer follows Go version\n", name)
	return nil
}

// installAllDependentToolchains installs all go-version-dependent toolchains
// into the specified Go version's bin directory, using the module path and
// version recorded in the config. Updates the config entries to point to the
// new Go version.
func installAllDependentToolchains(goVersion string) error {
	cfg, err := loadToolchainCfg()
	if err != nil {
		return err
	}

	var installed int
	gobinDir := filepath.Join(versionsDir, goVersion, "bin")

	for i := range cfg.Toolchains {
		tc := &cfg.Toolchains[i]
		if !tc.GoVersionDependent || tc.ModulePath == "" || tc.Version == "" {
			continue
		}

		if err := installModule(tc.ModulePath, tc.Version, gobinDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to install %s: %v\n", tc.Name, err)
			continue
		}

		tc.GoVersion = goVersion
		tc.InstallPath = filepath.Join(gobinDir, tc.Name)
		installed++
	}

	if installed > 0 {
		if err := saveToolchainCfg(cfg); err != nil {
			return fmt.Errorf("save config after installing toolchains: %w", err)
		}
		fmt.Printf("Installed %d toolchain(s) for Go %s\n", installed, goVersion)
	} else {
		fmt.Println("No following toolchains configured to install")
	}
	return nil
}

// migrateDependentToolchains moves go-version-dependent toolchains
// from the old Go version's bin directory to the new one.
func migrateDependentToolchains(oldVersion, newVersion string) error {
	cfg, err := loadToolchainCfg()
	if err != nil {
		return err
	}

	for i := range cfg.Toolchains {
		tc := &cfg.Toolchains[i]
		if !tc.GoVersionDependent || tc.GoVersion != oldVersion {
			continue
		}

		oldPath := filepath.Join(versionsDir, oldVersion, "bin", tc.Name)
		newPath := filepath.Join(versionsDir, newVersion, "bin", tc.Name)

		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			tc.GoVersion = newVersion
			tc.InstallPath = newPath
			continue
		}

		if err := copyFile(oldPath, newPath); err != nil {
			return fmt.Errorf("migrate %s: %w", tc.Name, err)
		}
		if err := os.Remove(oldPath); err != nil {
			_ = os.Remove(newPath)
			return fmt.Errorf("migrate %s: remove %s: %w", tc.Name, oldPath, err)
		}

		tc.GoVersion = newVersion
		tc.InstallPath = newPath
	}

	return saveToolchainCfg(cfg)
}
