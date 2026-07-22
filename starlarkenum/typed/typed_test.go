// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typed_test

import (
	"strings"
	"testing"

	"go.starlark.net/resolve"
	"go.starlark.net/starlarkenum/typed"
	"go.starlark.net/syntax"
	"go.starlark.net/typecheck"
)

// check parses, resolves, and typechecks src with the enum builtin
// predeclared.
func check(t *testing.T, src string) *typecheck.Result {
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
	res, err := typecheck.Check(f, env, nil)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	return res
}

func TestEnumChecks(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		want []string // expected error substrings, in order; empty => no errors
	}{
		{
			"element_typo",
			`
Color = enum("red", "green", "blue")
x = Color.bleu
`,
			[]string{"Object of type `type[Color]` has no attribute `bleu`"},
		},
		{
			"usage_ok",
			`
Color = enum("red", "green", "blue")
c = Color("red")

def value() -> str:
    return c.value

def index() -> int:
    return c.index

def names() -> list[str]:
    return Color.values()

first = Color[0]

def h(x: Color) -> str:
    return x.value

y = h(Color.green)
n = len(Color)
`,
			nil,
		},
		{
			"annotation_rejects_string",
			`
Color = enum("red", "green")

def f(x: Color) -> str:
    return x.value

f("red")
`,
			[]string{"Expected type `Color` but got `str`"},
		},
		{
			"ctor_bad_arg",
			`
Color = enum("red", "green")
c = Color(5)
`,
			[]string{"Expected type `Color | str` but got `int`"},
		},
		{
			"iteration",
			`
Color = enum("red", "green")

def all_values() -> list[str]:
    return [c.value for c in Color]
`,
			nil,
		},
		{
			"nominal_identity",
			`
A = enum("x")
B = enum("x")

def f(a: A) -> str:
    return a.value

f(B.x)
`,
			[]string{"Expected type `A` but got `B`"},
		},
		{
			"union_with_none",
			`
Color = enum("red", "green")
MaybeColor = Color | None

def f(x: MaybeColor) -> int:
    return 0

f(None)
f(Color.red)
f(5)
`,
			[]string{"Expected type `Color | None` but got `int`"},
		},
		{
			"non_literal_elements_degrade",
			`
def pick():
    return "red"

Color = enum(pick())
c = Color("anything")
x = Color.whatever
`,
			nil,
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

// TestEnumDenoted checks the Interface channels of an enum-defining
// module: the binding's type is the enum type value, and it denotes
// the element type, so enums work across load().
func TestEnumDenoted(t *testing.T) {
	res := check(t, `
Color = enum("red", "green")
`)
	if ty, ok := res.Interface.Get("Color"); !ok || ty.String() != "type[Color]" {
		t.Errorf("Get(Color) = %s, %v; want type[Color]", ty, ok)
	}
	if ty, ok := res.Interface.Denoted("Color"); !ok || ty.String() != "Color" {
		t.Errorf("Denoted(Color) = %s, %v; want Color", ty, ok)
	}
}
