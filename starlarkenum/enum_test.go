// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package starlarkenum_test

import (
	"fmt"
	"maps"
	"path/filepath"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkenum"
	"go.starlark.net/starlarktest"
	"go.starlark.net/syntax"
)

func Test(t *testing.T) {
	testdata := starlarktest.DataFile("starlarkenum", ".")
	thread := &starlark.Thread{Load: load}
	starlarktest.SetReporter(thread, t)
	filename := filepath.Join(testdata, "testdata/enum.star")
	predeclared := make(starlark.StringDict)
	maps.Copy(predeclared, starlarkenum.Predeclared)
	opts := &syntax.FileOptions{Types: syntax.TypesEnabled}
	if _, err := starlark.ExecFileOptions(opts, thread, filename, nil, predeclared); err != nil {
		if err, ok := err.(*starlark.EvalError); ok {
			t.Fatal(err.Backtrace())
		}
		t.Fatal(err)
	}
}

// load implements the 'load' operation as used in the evaluator tests.
func load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	if module == "assert.star" {
		return starlarktest.LoadAssertModule()
	}
	return nil, fmt.Errorf("load not implemented")
}
