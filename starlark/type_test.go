// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package starlark

import (
	"strings"
	"testing"

	"go.starlark.net/syntax"
)

// customMatcher is an embedder-defined type implementing TypeMatcher.
type customMatcher struct{ name string }

func (c *customMatcher) String() string        { return c.name }
func (c *customMatcher) Type() string          { return "custom_type" }
func (c *customMatcher) Freeze()               {}
func (c *customMatcher) Truth() Bool           { return True }
func (c *customMatcher) Hash() (uint32, error) { return 0, nil }
func (c *customMatcher) Matches(v Value) bool  { return v.Type() == "custom" }

// namedValue is an embedder-defined type implementing TypeName.
type namedValue struct{}

func (namedValue) String() string        { return "duration" }
func (namedValue) Type() string          { return "duration_factory" }
func (namedValue) Freeze()               {}
func (namedValue) Truth() Bool           { return True }
func (namedValue) Hash() (uint32, error) { return 0, nil }
func (namedValue) TypeName() string      { return "duration" }

func mustTypeOf(t *testing.T, v Value) *Type {
	t.Helper()
	ty, err := TypeOf(v)
	if err != nil {
		t.Fatalf("TypeOf(%s): %v", v, err)
	}
	return ty
}

func mustIndex(t *testing.T, x, y Value) Value {
	t.Helper()
	v, err := getIndex(x, y)
	if err != nil {
		t.Fatalf("%s[%s]: %v", x, y, err)
	}
	return v
}

func TestTypeOfConversions(t *testing.T) {
	for _, test := range []struct {
		v    Value
		want string // canonical String() of the type, or "error: <prefix>"
	}{
		{None, "None"},
		{Universe["int"], "int"},
		{Universe["str"], "str"},
		{Universe["bool"], "bool"},
		{Universe["float"], "float"},
		{Universe["bytes"], "bytes"},
		{Universe["range"], "range"},
		{Universe["list"], "list"},
		{Universe["dict"], "dict"},
		{Universe["tuple"], "tuple"},
		{Universe["set"], "set"},
		{Universe["type"], "type"},
		{TypingAny, "typing.Any"},
		{TypingNever, "typing.Never"},
		{TypingCallable, "typing.Callable"},
		{TypingIterable, "typing.Iterable"},
		{Tuple{Universe["int"], Universe["str"]}, "tuple[int, str]"},
		{&customMatcher{name: "mytype"}, "mytype"},
		{namedValue{}, "duration"},
		{String("int"), "error: string literal is not allowed in type expression"},
		{MakeInt(3), "error: value of type int is not a type"},
		{Universe["len"], "error: value of type builtin_function_or_method is not a type"},
	} {
		ty, err := TypeOf(test.v)
		if wantErr, ok := strings.CutPrefix(test.want, "error: "); ok {
			if err == nil {
				t.Errorf("TypeOf(%s) = %s, want error %q", test.v, ty, wantErr)
			} else if !strings.HasPrefix(err.Error(), wantErr) {
				t.Errorf("TypeOf(%s) error %q, want prefix %q", test.v, err, wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("TypeOf(%s): %v", test.v, err)
		} else if ty.String() != test.want {
			t.Errorf("TypeOf(%s) = %s, want %s", test.v, ty, test.want)
		}
	}
}

func TestTypeParameterization(t *testing.T) {
	listOfInt := mustIndex(t, Universe["list"], Universe["int"])
	if got := listOfInt.String(); got != "list[int]" {
		t.Errorf("list[int] = %s", got)
	}
	dictStrInt := mustIndex(t, Universe["dict"], Tuple{Universe["str"], Universe["int"]})
	if got := dictStrInt.String(); got != "dict[str, int]" {
		t.Errorf("dict[str, int] = %s", got)
	}
	openTuple := mustIndex(t, Universe["tuple"], Tuple{Universe["int"], Ellipsis})
	if got := openTuple.String(); got != "tuple[int, ...]" {
		t.Errorf("tuple[int, ...] = %s", got)
	}
	callable := mustIndex(t, TypingCallable, Tuple{NewList([]Value{Universe["int"]}), Universe["str"]})
	if got := callable.String(); got != "typing.Callable[[int], str]" {
		t.Errorf("typing.Callable[[int], str] = %s", got)
	}
	callableAny := mustIndex(t, TypingCallable, Tuple{Ellipsis, Universe["str"]})
	if got := callableAny.String(); got != "typing.Callable[..., str]" {
		t.Errorf("typing.Callable[..., str] = %s", got)
	}
	iterable := mustIndex(t, TypingIterable, Universe["int"])
	if got := iterable.String(); got != "typing.Iterable[int]" {
		t.Errorf("typing.Iterable[int] = %s", got)
	}

	// arity errors
	if _, err := getIndex(Universe["dict"], Universe["int"]); err == nil {
		t.Errorf("dict[int] succeeded, want arity error")
	}
	// int is not parameterizable (rust parity)
	if _, err := getIndex(Universe["int"], Universe["int"]); err == nil {
		t.Errorf("int[int] succeeded, want error")
	}
	// misplaced ellipsis
	if _, err := getIndex(Universe["tuple"], Tuple{Ellipsis, Universe["int"]}); err == nil {
		t.Errorf("tuple[..., int] succeeded, want error")
	}
}

func TestTypeUnion(t *testing.T) {
	union, err := Binary(syntax.PIPE, Universe["int"], None)
	if err != nil {
		t.Fatalf("int | None: %v", err)
	}
	if got := union.String(); got != "int | None" {
		t.Errorf("int | None = %s", got)
	}

	// flattening and dedup
	u2, err := Binary(syntax.PIPE, union, Universe["int"])
	if err != nil {
		t.Fatalf("(int | None) | int: %v", err)
	}
	if got := u2.String(); got != "int | None" {
		t.Errorf("(int | None) | int = %s", got)
	}

	// parameterized operand
	listOfInt := mustIndex(t, Universe["list"], Universe["int"])
	u3, err := Binary(syntax.PIPE, listOfInt, Universe["str"])
	if err != nil {
		t.Fatalf("list[int] | str: %v", err)
	}
	if got := u3.String(); got != "list[int] | str" {
		t.Errorf("list[int] | str = %s", got)
	}

	// ordinary value semantics are unaffected
	if v, err := Binary(syntax.PIPE, MakeInt(5), MakeInt(3)); err != nil || v.(Int).Sign() == 0 {
		t.Errorf("5 | 3 = %v, %v", v, err)
	}
	if _, err := Binary(syntax.PIPE, MakeInt(5), None); err == nil {
		t.Errorf("5 | None succeeded, want error")
	}
}

func TestTypeMatches(t *testing.T) {
	listOfInt := mustTypeOf(t, mustIndex(t, Universe["list"], Universe["int"]))
	intOrNone, _ := Binary(syntax.PIPE, Universe["int"], None)
	openTuple := mustTypeOf(t, mustIndex(t, Universe["tuple"], Tuple{Universe["int"], Ellipsis}))
	closedTuple := mustTypeOf(t, Tuple{Universe["int"], Universe["str"]})

	for _, test := range []struct {
		ty   *Type
		v    Value
		want bool
	}{
		{mustTypeOf(t, Universe["int"]), MakeInt(1), true},
		{mustTypeOf(t, Universe["int"]), String("x"), false},
		{mustTypeOf(t, Universe["str"]), String("x"), true},
		{mustTypeOf(t, Universe["float"]), Float(1.5), true},
		{mustTypeOf(t, Universe["float"]), MakeInt(1), true}, // numeric coercion (rust parity)
		{mustTypeOf(t, Universe["float"]), String("1.5"), false},
		{mustTypeOf(t, Universe["int"]), Float(1.0), false}, // coercion is one-way
		{mustTypeOf(t, None), None, true},
		{mustTypeOf(t, None), MakeInt(0), false},
		{TypingAny, String("anything"), true},
		{TypingNever, None, false},
		{TypingCallable, Universe["len"], true},
		{TypingCallable, MakeInt(1), false},
		{TypingIterable, NewList(nil), true},
		{TypingIterable, String("abc"), false}, // strings do not iterate
		{listOfInt, NewList([]Value{MakeInt(1), MakeInt(2)}), true},
		{listOfInt, NewList([]Value{MakeInt(1), String("a")}), false}, // deep matching
		{listOfInt, NewList(nil), true},
		{listOfInt, Tuple{MakeInt(1)}, false},
		{mustTypeOf(t, intOrNone), MakeInt(1), true},
		{mustTypeOf(t, intOrNone), None, true},
		{mustTypeOf(t, intOrNone), String("x"), false},
		{openTuple, Tuple{MakeInt(1), MakeInt(2), MakeInt(3)}, true},
		{openTuple, Tuple{}, true},
		{openTuple, Tuple{String("x")}, false},
		{closedTuple, Tuple{MakeInt(1), String("x")}, true},
		{closedTuple, Tuple{MakeInt(1)}, false},
		{mustTypeOf(t, Universe["type"]), Universe["int"], true}, // isinstance(int, type)
		{mustTypeOf(t, Universe["type"]), MakeInt(1), false},     // isinstance(1, type)
		{mustTypeOf(t, Universe["type"]), listOfInt, true},       // isinstance(list[int], type)
		{mustTypeOf(t, &customMatcher{name: "m"}), String("x"), false},
	} {
		if got := test.ty.Matches(test.v); got != test.want {
			t.Errorf("(%s).Matches(%s) = %v, want %v", test.ty, test.v, got, test.want)
		}
	}
}

// TestTypeAnnotationRoundtrip checks that type annotations (and the
// ! error marker) survive program serialization.
func TestTypeAnnotationRoundtrip(t *testing.T) {
	const src = `
errors = error_tags("Boom")

def double(x: int) -> int:
    return 2 * x

def may_fail(ok)! -> str:
    if not ok:
        return errors.Boom
    return "fine"

def run():
    t: tuple[int, ...] = (1, 2)
    return double(3)

def posonly(x, /, y):
    return (x, y)
`
	opts := &syntax.FileOptions{Types: syntax.TypesEnabled, PositionalOnly: true}
	_, prog, err := SourceProgramOptions(opts, "roundtrip.star", src, func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	if err := prog.Write(&buf); err != nil {
		t.Fatal(err)
	}
	prog2, err := CompiledProgram(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatal(err)
	}

	thread := new(Thread)
	globals, err := prog2.Init(thread, nil)
	if err != nil {
		t.Fatal(err)
	}

	double := globals["double"].(*Function)
	if v, err := Call(thread, double, Tuple{MakeInt(5)}, nil); err != nil {
		t.Errorf("double(5): %v", err)
	} else if v != MakeInt(10) {
		t.Errorf("double(5) = %v", v)
	}
	if _, err := Call(thread, double, Tuple{String("a")}, nil); err == nil {
		t.Errorf("double(\"a\") succeeded, want type error")
	} else if !strings.Contains(err.Error(), "does not match the type annotation `int` for argument `x`") {
		t.Errorf("double(\"a\") error = %v", err)
	}
	if ty := double.ParamType(0); ty == nil || ty.String() != "int" {
		t.Errorf("double.ParamType(0) = %v, want int", ty)
	}
	if ty := double.ReturnType(); ty == nil || ty.String() != "int" {
		t.Errorf("double.ReturnType() = %v, want int", ty)
	}

	mayFail := globals["may_fail"].(*Function)
	if !mayFail.CanReturnError() {
		t.Errorf("may_fail.CanReturnError() = false after roundtrip, want true")
	}

	run := globals["run"].(*Function)
	if v, err := Call(thread, run, nil, nil); err != nil {
		t.Errorf("run(): %v", err)
	} else if v != MakeInt(6) {
		t.Errorf("run() = %v", v)
	}

	// Positional-only markers survive the round trip.
	posonly := globals["posonly"].(*Function)
	if n := posonly.NumPositionalOnly(); n != 1 {
		t.Errorf("posonly.NumPositionalOnly() = %d, want 1", n)
	}
	if _, err := Call(thread, posonly, nil, []Tuple{{String("x"), MakeInt(1)}, {String("y"), MakeInt(2)}}); err == nil {
		t.Errorf("posonly(x=1, y=2) succeeded, want positional-only error")
	} else if !strings.Contains(err.Error(), "positional-only parameter") {
		t.Errorf("posonly(x=1, y=2) error = %v", err)
	}
}

func TestTypeEquality(t *testing.T) {
	a := mustIndex(t, Universe["list"], Universe["int"])
	b := mustIndex(t, Universe["list"], Universe["int"])
	c := mustIndex(t, Universe["list"], Universe["str"])
	eq, err := Compare(syntax.EQL, a, b)
	if err != nil || !eq {
		t.Errorf("list[int] == list[int]: %v, %v", eq, err)
	}
	ne, err := Compare(syntax.NEQ, a, c)
	if err != nil || !ne {
		t.Errorf("list[int] != list[str]: %v, %v", ne, err)
	}
}
