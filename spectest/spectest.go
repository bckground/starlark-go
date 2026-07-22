// Copyright 2026 bckground labs (bckground.com). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package spectest is the reference runner for the Starlark spec
// suite in the spec/ directory. It implements the runner obligations
// of spec/harness.md: it discovers spec files, gates them on the
// units this implementation supports, executes each program in a
// fresh environment with the assertion vocabulary predeclared, and
// matches the error expectations of chunked *_errors.star files.
//
// The runner also supports the contract's known-failures overlay: an
// out-of-tree list of spec files expected to fail for this
// implementation (see [ReadOverlay]). A listed file that fails is
// reported as a skip with its failures logged; a listed file that
// passes is an error, so the overlay cannot go stale.
package spectest

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"go.starlark.net/internal/chunkedfile"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
	"go.starlark.net/syntax"
)

// Supported is the set of optional units provided by this
// implementation.
var Supported = map[string]bool{
	"set":             true,
	"while":           true,
	"recursion":       true,
	"toplevelcontrol": true,
	"globalreassign":  true,
	"positionalonly":  true,
	"types":           true,
	"defer":           true,
	"error_handling":  true,
}

// A Config parameterizes a run of the spec suite.
type Config struct {
	// Supported is the set of optional units the implementation
	// provides; files requiring any other unit are skipped.
	Supported map[string]bool

	// KnownFailures is the known-failures overlay: slash-separated
	// paths, relative to the suite root, of spec files expected to
	// fail. May be nil.
	KnownFailures map[string]bool
}

// Run executes every spec file under root as a subtest of t.
func Run(t *testing.T, root string, cfg Config) {
	files := discover(t, root)
	if len(files) == 0 {
		t.Fatalf("no spec files found under %s", root)
	}
	for _, rel := range files {
		t.Run(rel, func(t *testing.T) {
			runFile(t, filepath.Join(root, filepath.FromSlash(rel)), rel, cfg)
		})
	}
}

// ReadOverlay reads a known-failures overlay file: one slash-separated
// spec file path per line, relative to the suite root; blank lines and
// lines starting with # are ignored. A nonexistent file is an empty
// overlay.
func ReadOverlay(filename string) (map[string]bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	overlay := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		overlay[line] = true
	}
	return overlay, scanner.Err()
}

// discover returns the slash-separated paths, relative to root, of
// every *.star file under root, sorted.
func discover(t *testing.T, root string) []string {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".star") {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return files
}

// A header holds the comment directives of a spec file.
type header struct {
	requires []string // units named by "# requires:"
	spec     string   // anchor named by "# spec:"
}

// parseHeader extracts the directives from the leading comment block
// of a spec file.
func parseHeader(src string) header {
	var h header
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "#") {
			break // end of leading comment block
		}
		if rest, ok := strings.CutPrefix(line, "# requires:"); ok {
			for _, unit := range strings.Split(rest, ",") {
				if unit := strings.TrimSpace(unit); unit != "" {
					h.requires = append(h.requires, unit)
				}
			}
		} else if rest, ok := strings.CutPrefix(line, "# spec:"); ok {
			h.spec = strings.TrimSpace(rest)
		}
	}
	return h
}

// requirements returns the units required by the spec file at the
// slash-separated path rel: those implied by its location plus those
// named in its header, deduplicated, in order of appearance.
func requirements(rel string, h header) []string {
	var units []string
	parts := strings.Split(rel, "/")
	switch parts[0] {
	case "optional":
		if len(parts) > 1 {
			units = append(units, parts[1])
		}
	case "interactions":
		if len(parts) > 1 {
			units = append(units, strings.Split(parts[1], "+")...)
		}
	}
	seen := make(map[string]bool)
	var result []string
	for _, unit := range append(units, h.requires...) {
		if !seen[unit] {
			seen[unit] = true
			result = append(result, unit)
		}
	}
	return result
}

// chunkMarkers returns the line numbers at which src uses the chunked
// file notation: "---" separator lines and "###" expectations. The
// harness contract permits these only in *_errors.star files.
func chunkMarkers(src string) []int {
	var lines []int
	for i, line := range strings.Split(src, "\n") {
		if strings.TrimRight(line, "\r") == "---" || strings.Contains(line, "###") {
			lines = append(lines, i+1)
		}
	}
	return lines
}

// optionsFor maps a spec file's required units to this
// implementation's dialect options. The defer and error_handling
// units are unconditionally enabled in this implementation and need
// no option.
func optionsFor(units []string) *syntax.FileOptions {
	opts := &syntax.FileOptions{}
	for _, unit := range units {
		switch unit {
		case "set":
			opts.Set = true
		case "while":
			opts.While = true
		case "recursion":
			opts.Recursion = true
		case "toplevelcontrol":
			opts.TopLevelControl = true
		case "globalreassign":
			opts.GlobalReassign = true
		case "positionalonly":
			opts.PositionalOnly = true
		case "types":
			opts.Types = syntax.TypesEnabled
		}
	}
	return opts
}

// A reporter receives the failures of a single spec program. It is
// satisfied by *testing.T (chunkedfile errors arrive via Errorf,
// assertion failures via Error) and by recorder.
type reporter interface {
	Errorf(format string, args ...any)
	Error(args ...any)
}

// A recorder is a reporter that accumulates failures instead of
// reporting them, for running known-failure files.
type recorder struct {
	failures []string
}

func (r *recorder) Errorf(format string, args ...any) {
	r.failures = append(r.failures, fmt.Sprintf(format, args...))
}

func (r *recorder) Error(args ...any) {
	r.failures = append(r.failures, fmt.Sprint(args...))
}

func runFile(t *testing.T, filename, rel string, cfg Config) {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	src := string(data)

	// Suite-integrity checks: these report on t even for files in the
	// known-failures overlay, which excuses implementation failures,
	// not malformed spec files.
	h := parseHeader(src)
	if h.spec == "" {
		t.Errorf("%s: missing '# spec:' header", filename)
	}
	if !strings.HasSuffix(rel, "_errors.star") {
		for _, line := range chunkMarkers(src) {
			t.Errorf("%s:%d: chunked file notation is permitted only in *_errors.star files", filename, line)
		}
	}

	units := requirements(rel, h)
	for _, unit := range units {
		if !cfg.Supported[unit] {
			t.Skipf("requires unit %q", unit)
		}
	}
	opts := optionsFor(units)

	if cfg.KnownFailures[rel] {
		rec := new(recorder)
		execFile(t, rec, opts, filename, rel, src)
		if len(rec.failures) == 0 {
			t.Errorf("%s: expected failure, but the file passed; remove it from the known-failures overlay", filename)
			return
		}
		for _, failure := range rec.failures {
			t.Logf("known failure: %s", failure)
		}
		t.Skipf("known failure")
	}
	execFile(t, t, opts, filename, rel, src)
}

// execFile executes the spec file as one or more programs, reporting
// failures to rep. t is used only for errors of the harness itself.
func execFile(t *testing.T, rep reporter, opts *syntax.FileOptions, filename, rel, src string) {
	predeclared, err := starlarktest.SpecPredeclared()
	if err != nil {
		t.Fatal(err)
	}

	if strings.HasSuffix(rel, "_errors.star") {
		for _, chunk := range chunkedfile.Read(filename, rep) {
			thread := &starlark.Thread{Name: rel}
			starlarktest.SetReporter(thread, rep)
			_, err := starlark.ExecFileOptions(opts, thread, filename, chunk.Source, predeclared)
			switch err := err.(type) {
			case nil:
				// success
			case *starlark.EvalError:
				// Report the error at the innermost frame in this file.
				found := false
				for i := range err.CallStack {
					posn := err.CallStack.At(i).Pos
					if posn.Filename() == filename {
						chunk.GotError(int(posn.Line), err.Error())
						found = true
						break
					}
				}
				if !found {
					rep.Error(err.Backtrace())
				}
			case resolve.ErrorList:
				for _, e := range err {
					chunk.GotError(int(e.Pos.Line), e.Msg)
				}
			case syntax.Error:
				chunk.GotError(int(err.Pos.Line), err.Msg)
			default:
				rep.Errorf("\n%s", err)
			}
			chunk.Done()
		}
	} else {
		thread := &starlark.Thread{Name: rel}
		starlarktest.SetReporter(thread, rep)
		_, err := starlark.ExecFileOptions(opts, thread, filename, src, predeclared)
		switch err := err.(type) {
		case nil:
			// success
		case *starlark.EvalError:
			rep.Errorf("\n%s", err.Backtrace())
		default:
			rep.Errorf("\n%s", err)
		}
	}
}
