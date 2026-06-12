// Copyright 2026 bckground labs (bckground.com). All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spectest

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"go.starlark.net/starlarktest"
	"go.starlark.net/syntax"
)

// TestSuite runs the spec suite against this implementation.
func TestSuite(t *testing.T) {
	overlay, err := ReadOverlay(starlarktest.DataFile("spectest", "known_failures.txt"))
	if err != nil {
		t.Fatal(err)
	}
	Run(t, starlarktest.DataFile("spectest", "../spec"), Config{
		Supported:     Supported,
		KnownFailures: overlay,
	})
}

func TestParseHeader(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		want header
	}{
		{
			"both",
			"# requires: error_handling, defer\n# spec: spec.md#errdefer\nx = 1\n",
			header{requires: []string{"error_handling", "defer"}, spec: "spec.md#errdefer"},
		},
		{
			"spec_only",
			"# spec: spec.md#integers\n\nassert.eq(1, 1)\n",
			header{spec: "spec.md#integers"},
		},
		{
			"none",
			"x = 1\n",
			header{},
		},
		{
			"directives_after_code_ignored",
			"x = 1\n# spec: spec.md#late\n",
			header{},
		},
		{
			"blank_lines_in_block",
			"# comment\n\n# spec: spec.md#a\nx = 1\n",
			header{spec: "spec.md#a"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := parseHeader(test.src); !reflect.DeepEqual(got, test.want) {
				t.Errorf("parseHeader = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestRequirements(t *testing.T) {
	for _, test := range []struct {
		rel  string
		h    header
		want []string
	}{
		{"core/int/division.star", header{}, nil},
		{"optional/error_handling/try/basic.star", header{}, []string{"error_handling"}},
		{
			"optional/error_handling/errdefer/cleanup.star",
			header{requires: []string{"defer"}},
			[]string{"error_handling", "defer"},
		},
		{
			"interactions/error_handling+types/annotated.star",
			header{},
			[]string{"error_handling", "types"},
		},
		{
			// header repeating the implicit unit is deduplicated
			"optional/set/literals.star",
			header{requires: []string{"set"}},
			[]string{"set"},
		},
	} {
		if got := requirements(test.rel, test.h); !reflect.DeepEqual(got, test.want) {
			t.Errorf("requirements(%q, %+v) = %v, want %v", test.rel, test.h, got, test.want)
		}
	}
}

func TestChunkMarkers(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		want []int
	}{
		{"none", "x = 1\ny = 2\n", nil},
		{"separator", "x = 1\n---\ny = 2\n", []int{2}},
		{"expectation", "x = 1 // 0 ### \"division by zero\"\n", []int{1}},
		{"dashes_not_alone", "s = '---'\n", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := chunkMarkers(test.src); !reflect.DeepEqual(got, test.want) {
				t.Errorf("chunkMarkers = %v, want %v", got, test.want)
			}
		})
	}
}

func TestReadOverlay(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "known_failures.txt")
	const overlay = `
# a comment
core/int/floored_division.star

optional/error_handling/try/basic.star
`
	if err := os.WriteFile(filename, []byte(overlay), 0666); err != nil {
		t.Fatal(err)
	}
	got, err := ReadOverlay(filename)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"core/int/floored_division.star":         true,
		"optional/error_handling/try/basic.star": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadOverlay = %v, want %v", got, want)
	}

	// A nonexistent overlay is empty, not an error.
	got, err = ReadOverlay(filepath.Join(t.TempDir(), "missing.txt"))
	if err != nil || got != nil {
		t.Errorf("ReadOverlay(missing) = %v, %v; want nil, nil", got, err)
	}
}

// TestKnownFailureRecording checks that a recorder captures every
// kind of failure a spec program can produce: assertion failures,
// aborts in plain files, and chunk-expectation mismatches.
func TestKnownFailureRecording(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) string {
		filename := filepath.Join(dir, name)
		if err := os.WriteFile(filename, []byte(src), 0666); err != nil {
			t.Fatal(err)
		}
		return filename
	}

	for _, test := range []struct {
		name, src    string
		wantFailures int
	}{
		{"passing.star", "# spec: spec.md#x\nassert.eq(1, 1)\n", 0},
		{"asserts.star", "# spec: spec.md#x\nassert.eq(1, 2)\nassert.eq(2, 3)\n", 2},
		{"abort.star", "# spec: spec.md#x\nx = 1 // 0\n", 1},
		{"unmet_errors.star", "# spec: spec.md#x\nx = 1 ### \"an error that does not occur\"\n", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			filename := write(test.name, test.src)
			rec := new(recorder)
			execFile(t, rec, &syntax.FileOptions{}, filename, test.name, test.src)
			if len(rec.failures) != test.wantFailures {
				t.Errorf("got %d failures, want %d: %q", len(rec.failures), test.wantFailures, rec.failures)
			}
		})
	}
}
