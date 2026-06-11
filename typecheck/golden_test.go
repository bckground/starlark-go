// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck_test

// A golden-file harness in the style of starlark-rust's
// typing/tests/golden: each testdata/golden/NAME.star input is
// checked, and the reported errors must match testdata/golden/
// NAME.golden. Regenerate the golden files with:
//
//	go test ./typecheck -run TestGolden -update

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

func TestGolden(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "golden", "*.star"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no golden test inputs in testdata/golden")
	}
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".star")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			res := checkNamed(t, filepath.Base(file), string(data), nil)

			var report strings.Builder
			if len(res.Errors) == 0 {
				report.WriteString("No errors.\n")
			} else {
				for _, e := range res.Errors {
					fmt.Fprintf(&report, "%s\n", e)
				}
			}

			goldenFile := strings.TrimSuffix(file, ".star") + ".golden"
			if *update {
				if err := os.WriteFile(goldenFile, []byte(report.String()), 0666); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("missing golden file (regenerate with -update): %v", err)
			}
			if report.String() != string(want) {
				t.Errorf("golden mismatch for %s (regenerate with -update if intended)\ngot:\n%swant:\n%s",
					file, report.String(), want)
			}
		})
	}
}
