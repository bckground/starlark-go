// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck_test

// This test pins the agreement between the static interpretation of
// type annotations (typecheck.tyFromTypeExpr) and the runtime one
// (starlark.TypeOf): both must accept the same annotations and
// display them the same way.

import (
	"fmt"
	"testing"

	"go.starlark.net/lib/typing"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func TestAnnotationAgreement(t *testing.T) {
	for _, annot := range []string{
		"int",
		"str",
		"bool",
		"float",
		"bytes",
		"range",
		"None",
		"list",
		"list[int]",
		"list[list[str]]",
		"dict[str, int]",
		"set[int]",
		"tuple[int, str]",
		"tuple[int, ...]",
		"int | None",
		"list[int] | str | None",
		"typing.Any",
		"typing.Never",
		"typing.Callable",
		"typing.Callable[[int, str], bool]",
		"typing.Iterable",
		"typing.Iterable[int]",
		"type",
		"[int, str]", // legacy union
	} {
		t.Run(annot, func(t *testing.T) {
			src := fmt.Sprintf("def f(x: %s): pass\n", annot)

			// Runtime interpretation.
			opts := &syntax.FileOptions{Types: syntax.TypesEnabled, Set: true}
			thread := new(starlark.Thread)
			predeclared := starlark.StringDict{"typing": typing.Module}
			globals, err := starlark.ExecFileOptions(opts, thread, "agree.star", src, predeclared)
			if err != nil {
				t.Fatalf("runtime rejected annotation: %v", err)
			}
			runtimeTy := globals["f"].(*starlark.Function).ParamType(0)
			if runtimeTy == nil {
				t.Fatalf("runtime recorded no type")
			}

			// Static interpretation.
			res := check(t, src, nil)
			if len(res.Errors) > 0 {
				t.Fatalf("static checker rejected annotation: %v", res.Errors)
			}
			if len(res.Approximations) > 0 {
				t.Fatalf("static checker approximated: %v", res.Approximations)
			}
			var staticTy string
			for _, line := range splitLines(res.Types.String()) {
				if len(line) > 2 && line[0] == 'x' && line[1] == ' ' {
					staticTy = line[lastEq(line)+2:]
				}
			}

			// The displays agree, modulo union ordering (the runtime
			// preserves written order; the static checker canonicalizes)
			// and the runtime's unparameterized spellings.
			want := map[string]string{
				"list":                              "list[typing.Any]",
				"tuple":                             "tuple",
				"typing.Callable[[int, str], bool]": "typing.Callable",
				"typing.Iterable[int]":              "typing.Iterable[int]",
				"[int, str]":                        "int | str",
				"list[int] | str | None":            "list[int] | None | str", // canonical order
			}
			expected := runtimeTy.String()
			if w, ok := want[annot]; ok {
				expected = w
			}
			if staticTy != expected {
				t.Errorf("static %q != expected %q (runtime %q)", staticTy, expected, runtimeTy)
			}
		})
	}
}

func splitLines(s string) []string {
	var lines []string
	for len(s) > 0 {
		i := 0
		for i < len(s) && s[i] != '\n' {
			i++
		}
		lines = append(lines, s[:i])
		if i == len(s) {
			break
		}
		s = s[i+1:]
	}
	return lines
}

func lastEq(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '=' {
			return i
		}
	}
	return -1
}
