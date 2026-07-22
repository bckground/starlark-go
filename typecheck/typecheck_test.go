// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.starlark.net/resolve"
	"go.starlark.net/syntax"
	"go.starlark.net/typecheck"
)

// check parses, resolves, and typechecks src.
func check(t *testing.T, src string, loads map[string]*typecheck.Interface) *typecheck.Result {
	t.Helper()
	return checkNamed(t, "test.star", src, loads)
}

// checkNamed is check with an explicit filename for error positions.
func checkNamed(t *testing.T, filename, src string, loads map[string]*typecheck.Interface) *typecheck.Result {
	t.Helper()
	env := typecheck.UniverseEnv()
	env["typing"] = typecheck.Module("typing", nil)
	opts := &syntax.FileOptions{
		Types:          syntax.TypesEnabled,
		Set:            true,
		While:          true,
		Recursion:      true,
		PositionalOnly: true,
	}
	f, err := opts.Parse(filename, src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	isPredeclared := func(name string) bool { return false }
	isUniversal := func(name string) bool { _, ok := env[name]; return ok }
	if err := resolve.File(f, isPredeclared, isUniversal); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	res, err := typecheck.Check(f, env, loads)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	return res
}

func TestErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		want []string // expected error substrings, in order; empty => no errors
	}{
		{
			"clean",
			`
def f(x: int) -> int:
    return x + 1

y = f(1)
`,
			nil,
		},
		{
			"call_arg_mismatch",
			`
def f(x: int) -> int:
    return x

f("a")
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"return_mismatch",
			`
def f() -> str:
    return 1
`,
			[]string{"Expected type `str` but got `int`"},
		},
		{
			"implicit_return_ok",
			`
def f() -> int:
    return 1
`,
			nil,
		},
		{
			"binop_mismatch",
			`
def f():
    return 1 + "a"
`,
			[]string{"Binary operator `+` is not available on the types `int` and `str`"},
		},
		{
			"int_plus_float",
			`
def f() -> float:
    return 1 + 2.0
`,
			nil,
		},
		{
			"float_accepts_int",
			`
def f(x: float) -> float:
    return x

def g():
    f(1)
    f(1.5)
    f("a")
`,
			[]string{"Expected type `float` but got `str`"},
		},
		{
			"attr_missing",
			`
def f(x: list[int]):
    x.frobnicate()
`,
			[]string{"Object of type `list[int]` has no attribute `frobnicate`"},
		},
		{
			"call_non_callable",
			`
def f():
    x = 1
    x()
`,
			[]string{"Call to a non-callable type `int`"},
		},
		{
			"too_many_positional",
			`
def f(x):
    pass

def g():
    f(1, 2)
`,
			[]string{"Too many positional arguments"},
		},
		{
			"missing_required",
			`
def f(x, y):
    pass

def g():
    f(1)
`,
			[]string{"Missing required parameter `y`"},
		},
		{
			"unexpected_named",
			`
def f(x):
    pass

def g():
    f(x=1, z=2)
`,
			[]string{"Unexpected parameter named `z`"},
		},
		{
			"positional_only",
			`
def f(x: int, /, y: str):
    pass

def g():
    f(1, "a")
    f(1, y="a")
    f(x=1, y="a")
`,
			[]string{"Unexpected parameter named `x`"},
		},
		{
			"positional_only_kwargs",
			`
def f(x, /, **kwargs):
    pass

def g():
    f(1, x=2)
`,
			nil,
		},
		{
			"vargs_suppresses_missing",
			`
def f(x, y):
    pass

def g(args):
    f(1, *args)
`,
			nil,
		},
		{
			"not_iterable",
			`
def f():
    for x in 3:
        pass
`,
			[]string{"Type `int` is not iterable"},
		},
		{
			"index_operator",
			`
def f(x: int):
    return x[0]
`,
			[]string{"Type `int` does not have [] operator"},
		},
		{
			"dict_key_type",
			`
def f(d: dict[str, int]) -> int:
    return d["a"]

def g(d: dict[str, int]):
    return d[0]
`,
			[]string{"Type `dict[str, int]` does not have [] operator on `int`"},
		},
		{
			"annotated_assignment",
			`
def f():
    x: int = "a"
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"default_mismatch",
			`
def f(x: int = "a"):
    pass
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"union_flow",
			`
def f(x: int | None) -> int:
    return 0

def g():
    f(None)
    f(1)
    f("a")
`,
			[]string{"Expected type `int | None` but got `str`"},
		},
		{
			"alias",
			`
IntList = list[int]

def f(xs: IntList) -> int:
    return len(xs)

def g():
    f([1, 2])
    f("not a list")
`,
			[]string{"Expected type `list[int]` but got `str`"},
		},
		{
			"list_append_inference",
			`
def f():
    xs = []
    xs.append(1)
    return g(xs)

def g(xs: list[str]) -> int:
    return 0
`,
			[]string{"Expected type `list[str]` but got `list[int]`"},
		},
		{
			"string_methods",
			`
def f(s: str) -> list[str]:
    return s.split(",")

def g(s: str):
    return s.join([1, 2])
`,
			[]string{"Expected type `typing.Iterable[str]` but got `list[int]`"},
		},
		{
			"destructuring",
			`
def f(pair: tuple[int, str]) -> str:
    a, b = pair
    return a
`,
			[]string{"Expected type `str` but got `int`"},
		},
		{
			"comprehension",
			`
def f(xs: list[int]) -> list[str]:
    return [x + 1 for x in xs]
`,
			[]string{"Expected type `list[str]` but got `list[int]`"},
		},
		{
			"typing_any_disables",
			`
def f(x: typing.Any) -> int:
    return x.anything()(1)["key"]
`,
			nil,
		},
		{
			"fail_returns_never",
			`
def f(x: int) -> int:
    if x > 0:
        return x
    return fail("bad")
`,
			nil,
		},
		{
			"try_catch_types",
			`
errors = error_tags("Boom")

def may_fail(x: int)! -> int:
    if x < 0:
        return errors.Boom
    return x

def caller() -> int:
    return may_fail(1) catch 0

def caller2() -> str:
    return may_fail(1) catch 0
`,
			[]string{"Expected type `str` but got `int`"},
		},
		{
			"catch_block_recover",
			`
errors = error_tags("Boom")

def may_fail()! -> int:
    return errors.Boom

def caller() -> int:
    x = may_fail() catch e:
        recover 0
    return x

def caller2() -> str:
    x = may_fail() catch e:
        recover "fallback"
    return x
`,
			nil,
		},
		{
			"while_loop_and_augmented",
			`
def f(n: int) -> int:
    total = 0
    i = 0
    while i < n:
        total += i
        i += 1
    return total
`,
			nil,
		},
		{
			"kwargs_value_type",
			`
def f(**kwargs: int):
    pass

def g():
    f(a=1, b="x")
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"args_element_type",
			`
def f(*args: int):
    pass

def g():
    f(1, "x")
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"pointless_comparison",
			`
def f(x: int, y: str):
    return x == y
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"comparison_none_narrowing_ok",
			`
def f(x: int | None):
    return x == None
`,
			nil,
		},
		{
			"comparison_numeric_ok",
			`
def f(x: int, y: float):
    return x != y
`,
			nil,
		},
		{
			"callable_param_ok",
			`
def apply(f: typing.Callable[[int], str], x: int) -> str:
    return f(x)

def g(i: int) -> str:
    return str(i)

y = apply(g, 1)
`,
			nil,
		},
		{
			"callable_param_type_mismatch",
			`
def apply(f: typing.Callable[[int], str], x: int) -> str:
    return f(x)

def g(s: str) -> str:
    return s

y = apply(g, 1)
`,
			[]string{"Expected type `typing.Callable` but got `def g`"},
		},
		{
			"callable_result_mismatch",
			`
def apply(f: typing.Callable[[int], str]) -> str:
    return f(1)

def g(i: int) -> int:
    return i

y = apply(g)
`,
			[]string{"Expected type `typing.Callable` but got `def g`"},
		},
		{
			"callable_arity_mismatch",
			`
def apply(f: typing.Callable[[int, int], int]):
    pass

def g(i: int) -> int:
    return i

apply(g)
`,
			[]string{"but got `def g`"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			res := check(t, test.src, nil)
			var got []string
			for _, e := range res.Errors {
				got = append(got, e.Error())
			}
			if len(test.want) == 0 {
				if len(got) > 0 {
					t.Errorf("unexpected errors:\n%s", strings.Join(got, "\n"))
				}
				return
			}
			if len(got) != len(test.want) {
				t.Errorf("got %d errors, want %d:\ngot:\n%s\nwant substrings:\n%s",
					len(got), len(test.want), strings.Join(got, "\n"), strings.Join(test.want, "\n"))
				return
			}
			for i, want := range test.want {
				if !strings.Contains(got[i], want) {
					t.Errorf("error %d = %q, want substring %q", i, got[i], want)
				}
			}
		})
	}
}

func TestTypeMap(t *testing.T) {
	res := check(t, `
def f(xs: list[int]):
    total = 0
    for x in xs:
        total += x
    pair = (1, "a")
    ys = []
    ys.append("s")
    d = {"k": 1}
`, nil)
	want := map[string]string{
		"total": "int",
		"x":     "int",
		"pair":  "tuple[int, str]",
		"ys":    "list[str]",
		"d":     "dict[str, int]",
	}
	tmStr := res.Types.String()
	for name, ty := range want {
		needle := name + " (test.star:"
		found := false
		for _, line := range strings.Split(tmStr, "\n") {
			if strings.HasPrefix(line, needle) {
				found = true
				if !strings.HasSuffix(line, "= "+ty) {
					t.Errorf("binding %s: got %q, want type %s", name, line, ty)
				}
			}
		}
		if !found {
			t.Errorf("binding %s not in TypeMap:\n%s", name, tmStr)
		}
	}
}

// TestPreciseBuiltinResults exercises the argument-aware result types
// of the universe signature table (universe.go's specialFuncs).
func TestPreciseBuiltinResults(t *testing.T) {
	for _, test := range []struct {
		src  string // must bind a global x
		want string // expected type of x
	}{
		{`x = sorted([3, 1, 2])`, "list[int]"},
		{`x = sorted(["b", "a"], reverse=True)`, "list[str]"},
		{`x = reversed([1.0])`, "list[float]"},
		{`x = max(1, 2)`, "int"},
		{`x = max(1, "a")`, "int | str"},
		{`x = min([1.0, 2.0])`, "float"},
		{`x = abs(-1)`, "int"},
		{`x = abs(1.5)`, "float"},
		{`x = list("abc".elems())`, "list[str]"},
		{`x = list()`, "list[typing.Never]"},
		{`x = set([1])`, "set[int]"},
		{`x = tuple([1])`, "tuple[int, ...]"},
		{`x = zip([1], ["a"])`, "list[tuple[int, str]]"},
		{`x = enumerate(["a"])`, "list[tuple[int, str]]"},
		{`x = dict(a=1)`, "dict[str, int]"},
		{`x = dict([(1, "a")])`, "dict[int, str]"},
		{`x = dict({1: "a"}, b=2)`, "dict[int | str, int | str]"},
		{`x = {1: "a"}.get(1)`, "None | str"},
		{`x = {1: "a"}.get(1, "z")`, "str"},
		{`x = {1: "a"}.get(1, 0)`, "int | str"},
		{`x = {1: "a"}.pop(1)`, "str"},
		{`x = {1: "a"}.setdefault(1)`, "None | str"},
		// *args call forms fall back to the declared (lenient) result.
		{`args = ([1], [2])` + "\n" + `x = zip(*args)`, "list[tuple]"},
	} {
		t.Run(test.want, func(t *testing.T) {
			res := check(t, test.src, nil)
			if len(res.Errors) > 0 {
				t.Fatalf("unexpected errors: %v", res.Errors)
			}
			ty, ok := res.Interface.Get("x")
			if !ok {
				t.Fatalf("no global x")
			}
			if ty.String() != test.want {
				t.Errorf("x = %s, want %s", ty, test.want)
			}
		})
	}
}

// TestPartialEval exercises the module-level partial evaluator:
// type aliases under conditionals, defined inside functions, or
// reassigned, beyond the literal Name = <type-expression> shape.
func TestPartialEval(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		want []string // expected error substrings, in order; empty => no errors
	}{
		{
			"cond_alias_agree",
			`
FLAG = True
T = int if FLAG else int

def f(x: T) -> int:
    return x

f("a")
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"cond_alias_disagree",
			`
FLAG = True
T = int if FLAG else float

def f(x: T) -> int:
    return 0

f("a")
`,
			nil, // T is unknown: typing.Any
		},
		{
			"if_else_alias_agree",
			`
def h() -> int:
    if True:
        T = int
    else:
        T = int
    def g(x: T) -> int:
        return x
    return g("a")
`,
			[]string{"Expected type `int` but got `str`"},
		},
		{
			"if_else_alias_disagree",
			`
def h() -> int:
    if True:
        T = int
    else:
        T = str
    def g(x: T) -> int:
        return 0
    return g(3.5)
`,
			nil, // branches disagree: T is unknown
		},
		{
			"function_scope_alias",
			`
def outer() -> int:
    IntList = list[int]
    def inner(xs: IntList) -> int:
        return len(xs)
    return inner("nope")
`,
			[]string{"Expected type `list[int]` but got `str`"},
		},
		{
			"alias_chain",
			`
Row = list[int]
Matrix = list[Row]

def f(m: Matrix) -> int:
    return len(m)

f([["a"]])
`,
			[]string{"Expected type `list[list[int]]` but got `list[list[str]]`"},
		},
		{
			"alias_reassigned_conflict",
			`
def f() -> int:
    T = int
    T = "x"
    def g(y: T) -> int:
        return 0
    return g(3.5)
`,
			nil, // reassignment disagrees: T is unknown
		},
		{
			"alias_shadowed_by_loop_var",
			`
def f() -> int:
    T = int
    for T in ["a"]:
        pass
    def g(y: T) -> int:
        return 0
    return g(3.5)
`,
			nil, // loop variable rebinding: T is unknown
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			res := check(t, test.src, nil)
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

// TestErrorTagTyping exercises the static typing of error tag sets:
// the binding's type carries the exact tag set as its attribute
// table, so misspelled tags are caught.
func TestErrorTagTyping(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		want []string
	}{
		{
			"tag_typo",
			`
errors = error_tags("NotFound", "Timeout")

def f()! -> int:
    return errors.NotFonud
`,
			[]string{"Object of type `error_tags` has no attribute `NotFonud`"},
		},
		{
			"tag_ok",
			`
errors = error_tags("NotFound", "Timeout")

def f(x: int)! -> int:
    if x < 0:
        return errors.NotFound
    return x

def g() -> int:
    return f(1) catch 0
`,
			nil,
		},
		{
			"tag_call_constructs_error",
			`
errors = error_tags("Boom")

def f()! -> int:
    return errors.Boom(message="exploded")
`,
			nil,
		},
		{
			"tag_set_merge",
			`
a = error_tags("A")
b = error_tags("B")
c = a | b

def f()! -> int:
    return c.A
`,
			nil, // merged sets degrade to Any: lenient
		},
		{
			"success_return_still_checked",
			`
errors = error_tags("Boom")

def f()! -> int:
    return "not an int"
`,
			[]string{"Expected type `error | error_tag | int` but got `str`"},
		},
		{
			"tag_set_via_variable",
			`
errors = error_tags("NotFound")
errs2 = errors

def f()! -> int:
    return errs2.NotFonud
`,
			[]string{"has no attribute `NotFonud`"},
		},
		{
			"non_literal_tags_unknown",
			`
NAME = "NotFound"

def name() -> str:
    return "x"

errors = error_tags(name())

def f()! -> int:
    return errors.Whatever
`,
			nil, // non-literal arguments: tag set is unknown
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			res := check(t, test.src, nil)
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

// TestDenotedAcrossLoads checks that type aliases flow across load()
// through the Interface's denoted-type channel.
func TestDenotedAcrossLoads(t *testing.T) {
	lib := check(t, `
IntList = list[int]
`, nil)
	if len(lib.Errors) > 0 {
		t.Fatalf("lib errors: %v", lib.Errors)
	}
	if ty, ok := lib.Interface.Denoted("IntList"); !ok || ty.String() != "list[int]" {
		t.Fatalf("IntList denotes %s (ok=%v), want list[int]", ty, ok)
	}

	main := check(t, `
load("lib.star", "IntList")

def f(xs: IntList) -> int:
    return len(xs)

f("nope")
`, map[string]*typecheck.Interface{"lib.star": lib.Interface})
	var got []string
	for _, e := range main.Errors {
		got = append(got, e.Msg)
	}
	if len(got) != 1 || !strings.Contains(got[0], "Expected type `list[int]` but got `str`") {
		t.Errorf("errors = %q", got)
	}
}

func TestInterfaceAndLoads(t *testing.T) {
	// Check a library module, then a dependent module using its Interface.
	lib := check(t, `
def double(x: int) -> int:
    return 2 * x

VERSION = "1.0"
`, nil)
	if len(lib.Errors) > 0 {
		t.Fatalf("lib errors: %v", lib.Errors)
	}
	if ty, ok := lib.Interface.Get("VERSION"); !ok || ty.String() != "str" {
		t.Errorf("VERSION = %s, want str", ty)
	}

	main := check(t, `
load("lib.star", "double", "VERSION")

a = double(2)
b = double("x")
c = VERSION + 1
`, map[string]*typecheck.Interface{"lib.star": lib.Interface})
	var got []string
	for _, e := range main.Errors {
		got = append(got, e.Msg)
	}
	if len(got) != 2 ||
		!strings.Contains(got[0], "Expected type `int` but got `str`") ||
		!strings.Contains(got[1], "Binary operator `+` is not available on the types `str` and `int`") {
		t.Errorf("errors = %q", got)
	}
}

// TestCorpus runs the checker over the interpreter's entire testdata
// corpus: it must not panic, and must not report errors for programs
// that execute successfully (modulo the intentionally-failing chunks,
// so only panics and gross over-reporting are detected here).
func TestCorpus(t *testing.T) {
	for _, dir := range []string{"../starlark/testdata", "../syntax/testdata"} {
		files, err := filepath.Glob(filepath.Join(dir, "*.star"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			src := string(data)
			opts := &syntax.FileOptions{
				Set:             strings.Contains(src, "option:set"),
				While:           strings.Contains(src, "option:while"),
				TopLevelControl: strings.Contains(src, "option:toplevelcontrol"),
				GlobalReassign:  strings.Contains(src, "option:globalreassign"),
				Recursion:       strings.Contains(src, "option:recursion"),
				Types:           syntax.TypesEnabled,
			}
			// Many corpus files are chunked with --- separators and
			// contain intentional errors; we only check the chunks
			// that parse and resolve.
			for _, chunk := range strings.Split(src, "\n---\n") {
				f, err := opts.Parse(file, chunk, 0)
				if err != nil {
					continue
				}
				if err := resolve.File(f, func(string) bool { return true }, func(string) bool { return true }); err != nil {
					continue
				}
				func() {
					defer func() {
						if x := recover(); x != nil {
							t.Errorf("%s: panic: %v\nchunk:\n%s", file, x, chunk)
						}
					}()
					if _, err := typecheck.Check(f, typecheck.UniverseEnv(), nil); err != nil {
						t.Errorf("%s: %v", file, err)
					}
				}()
			}
		}
	}
}
