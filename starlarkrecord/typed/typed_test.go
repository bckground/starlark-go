// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typed_test

import (
	"strings"
	"testing"

	"go.starlark.net/resolve"
	"go.starlark.net/starlarkrecord/typed"
	"go.starlark.net/syntax"
	"go.starlark.net/typecheck"
)

// check parses, resolves, and typechecks src with the record and
// field builtins predeclared.
func check(t *testing.T, src string) *typecheck.Result {
	return checkLoads(t, src, nil)
}

// checkLoads is check with the Interfaces of loadable modules.
func checkLoads(t *testing.T, src string, loads map[string]*typecheck.Interface) *typecheck.Result {
	t.Helper()
	uni := typecheck.UniverseEnv()
	env := typecheck.UniverseEnv()
	typed.AddTypes(env)
	opts := &syntax.FileOptions{Types: syntax.TypesEnabled}
	f, err := opts.Parse("test.star", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	isUniversal := func(name string) bool { _, ok := uni[name]; return ok }
	isPredeclared := func(name string) bool {
		if _, ok := uni[name]; ok {
			return false
		}
		_, ok := env[name]
		return ok
	}
	if err := resolve.File(f, isPredeclared, isUniversal); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	res, err := typecheck.Check(f, env, loads)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	return res
}

func TestRecordChecks(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		want []string // expected error substrings, in order; empty => no errors
	}{
		{
			"field_typo",
			`
Rec = record(host=str, port=field(int, 80))
r = Rec(host="localhost")
x = r.hosst
`,
			[]string{"Object of type `Rec` has no attribute `hosst`"},
		},
		{
			"field_access_ok",
			`
Rec = record(host=str, port=field(int, 80))
r = Rec(host="localhost")

def f() -> int:
    return r.port + len(r.host)
`,
			nil,
		},
		{
			"ctor_field_type",
			`
Rec = record(host=str)
r = Rec(host=8)
`,
			[]string{"Expected type `str` but got `int`"},
		},
		{
			"ctor_missing_required",
			`
Rec = record(host=str, port=field(int, 80))
r = Rec()
`,
			[]string{"Missing required parameter `host`"},
		},
		{
			"ctor_default_not_required",
			`
Rec = record(host=str, port=field(int, 80))
r = Rec(host="h")
`,
			nil,
		},
		{
			"ctor_unknown_field",
			`
Rec = record(host=str)
r = Rec(host="h", extra=1)
`,
			[]string{"Unexpected parameter named `extra`"},
		},
		{
			"annotation_precision",
			`
Rec = record(host=str)

def serve(x: Rec) -> str:
    return x.host

r = Rec(host="h")
serve(r)
serve(5)
`,
			[]string{"Expected type `Rec` but got `int`"},
		},
		{
			"nominal_identity",
			`
A = record(x=int)
B = record(x=int)

def f(a: A) -> int:
    return a.x

b = B(x=1)
f(b)
`,
			[]string{"Expected type `A` but got `B`"},
		},
		{
			"alias_field_type",
			`
IntList = list[int]
Rec = record(values=IntList)
r = Rec(values="nope")
`,
			[]string{"Expected type `list[int]` but got `str`"},
		},
		{
			"union_with_none",
			`
Rec = record(host=str)
MaybeRec = Rec | None

def f(x: MaybeRec) -> int:
    return 0

f(None)
f(Rec(host="h"))
f(5)
`,
			[]string{"Expected type `Rec | None` but got `int`"},
		},
		{
			"unknown_args_degrade",
			`
def pick_type():
    return int

Rec = record(host=pick_type())
r = Rec(host=8, whatever=1)
x = r.anything
`,
			nil, // non-constant arguments: Rec is unknown, everything is Any
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			res := check(t, test.src)
			var got []string
			for _, e := range res.Errors {
				got = append(got, e.Error())
			}
			if len(got) != len(test.want) {
				t.Fatalf("got %d errors, want %d:\ngot:\n%s\nwant substrings:\n%s",
					len(got), len(test.want), strings.Join(got, "\n"), strings.Join(test.want, "\n"))
			}
			for i, want := range test.want {
				if !strings.Contains(got[i], want) {
					t.Errorf("error %d = %q, want substring %q", i, got[i], want)
				}
			}
		})
	}
}

// TestRecordAcrossLoads checks that a record type flows across
// load(): the constructor through the Interface's type channel, the
// instance type through its denoted channel, with nominal identity
// preserved.
func TestRecordAcrossLoads(t *testing.T) {
	lib := check(t, `
Rec = record(host=str)
`)
	if len(lib.Errors) > 0 {
		t.Fatalf("lib errors: %v", lib.Errors)
	}
	main := checkLoads(t, `
load("lib.star", "Rec")

def serve(x: Rec) -> str:
    return x.host

r = Rec(host="h")
serve(r)
bad = r.hosst
serve(5)
`, map[string]*typecheck.Interface{"lib.star": lib.Interface})
	var got []string
	for _, e := range main.Errors {
		got = append(got, e.Msg)
	}
	if len(got) != 2 ||
		!strings.Contains(got[0], "has no attribute `hosst`") ||
		!strings.Contains(got[1], "Expected type `Rec` but got `int`") {
		t.Errorf("errors = %q", got)
	}
}

// TestRecordDenoted checks the Interface channels of a record-defining
// module: the binding's type is the constructor, and it denotes the
// instance type, so records work across load().
func TestRecordDenoted(t *testing.T) {
	res := check(t, `
Rec = record(host=str)
`)
	if ty, ok := res.Interface.Get("Rec"); !ok || ty.String() != "type[Rec]" {
		t.Errorf("Get(Rec) = %s, %v; want type[Rec]", ty, ok)
	}
	if ty, ok := res.Interface.Denoted("Rec"); !ok || ty.String() != "Rec" {
		t.Errorf("Denoted(Rec) = %s, %v; want Rec", ty, ok)
	}
}
