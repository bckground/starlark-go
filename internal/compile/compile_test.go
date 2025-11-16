package compile_test

import (
	"bytes"
	"strings"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// TestSerialization verifies that a serialized program can be loaded,
// deserialized, and executed.
func TestSerialization(t *testing.T) {
	predeclared := starlark.StringDict{
		"x": starlark.String("mur"),
		"n": starlark.MakeInt(2),
	}
	const src = `
def mul(a, b):
    return a * b

y = mul(x, n)
`
	_, oldProg, err := starlark.SourceProgram("mul.star", src, predeclared.Has)
	if err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := oldProg.Write(buf); err != nil {
		t.Fatalf("oldProg.WriteTo: %v", err)
	}

	newProg, err := starlark.CompiledProgram(buf)
	if err != nil {
		t.Fatalf("CompiledProgram: %v", err)
	}

	thread := new(starlark.Thread)
	globals, err := newProg.Init(thread, predeclared)
	if err != nil {
		t.Fatalf("newProg.Init: %v", err)
	}
	if got, want := globals["y"], starlark.String("murmur"); got != want {
		t.Errorf("Value of global was %s, want %s", got, want)
		t.Logf("globals: %v", globals)
	}

	// Verify stack frame.
	predeclared["n"] = starlark.None
	_, err = newProg.Init(thread, predeclared)
	evalErr, ok := err.(*starlark.EvalError)
	if !ok {
		t.Fatalf("newProg.Init call returned err %v, want *EvalError", err)
	}
	const want = `Traceback (most recent call last):
  mul.star:5:8: in <toplevel>
  mul.star:3:14: in mul
Error: unknown binary op: string * NoneType`
	if got := evalErr.Backtrace(); got != want {
		t.Fatalf("got <<%s>>, want <<%s>>", got, want)
	}
}

func TestGarbage(t *testing.T) {
	const garbage = "This is not a compiled Starlark program."
	_, err := starlark.CompiledProgram(strings.NewReader(garbage))
	if err == nil {
		t.Fatalf("CompiledProgram did not report an error when decoding garbage")
	}
	if !strings.Contains(err.Error(), "not a compiled module") {
		t.Fatalf("CompiledProgram reported the wrong error when decoding garbage: %v", err)
	}
}

// TestStrictMultiValueReturnImplicitTuple verifies that in strict multi-value return mode,
// implicit tuple creation (e.g., "x = 1, 2") is rejected at compile time, while explicit
// forms with parentheses or brackets are allowed.
func TestStrictMultiValueReturnImplicitTuple(t *testing.T) {
	opts := &syntax.FileOptions{
		StrictMultiValueReturn: true,
	}

	// Test cases that should fail - implicit tuple creation
	invalidCases := []struct {
		name string
		src  string
	}{
		{
			name: "implicit_tuple_two_elements",
			src: `
x = 1, 2
`,
		},
		{
			name: "implicit_tuple_three_elements",
			src: `
x = 1, 2, 3
`,
		},
		{
			name: "implicit_tuple_in_function",
			src: `
def f():
    x = "a", "b"
    return x
`,
		},
		{
			name: "implicit_tuple_mixed_types",
			src: `
result = 42, "hello", True
`,
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := starlark.SourceProgramOptions(opts, "test.star", tc.src, func(string) bool { return false })
			if err == nil {
				t.Fatal("expected compilation error for implicit tuple creation, but compilation succeeded")
			}

			errMsg := err.Error()
			expectedMsg := "implicit tuple not allowed; use parentheses or brackets"
			if !strings.Contains(errMsg, expectedMsg) {
				t.Errorf("error message %q does not contain expected substring %q", errMsg, expectedMsg)
			}
		})
	}

	// Test cases that should succeed - explicit tuple/list creation or unpacking
	validCases := []struct {
		name string
		src  string
	}{
		{
			name: "explicit_tuple_with_parens",
			src: `
x = (1, 2)
`,
		},
		{
			name: "explicit_list_with_brackets",
			src: `
x = [1, 2]
`,
		},
		{
			name: "unpacking_bare_tuple",
			src: `
a, b = 1, 2
`,
		},
		{
			name: "unpacking_with_parens",
			src: `
a, b = (1, 2)
`,
		},
		{
			name: "single_value_assignment",
			src: `
x = 42
`,
		},
		{
			name: "explicit_tuple_three_elements",
			src: `
x = (1, 2, 3)
`,
		},
		{
			name: "explicit_list_in_function",
			src: `
def f():
    x = ["a", "b"]
    return x
`,
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := starlark.SourceProgramOptions(opts, "test.star", tc.src, func(string) bool { return false })
			if err != nil {
				t.Fatalf("expected successful compilation, but got error: %v", err)
			}
		})
	}
}

// TestStrictMultiValueReturnInconsistent verifies that in strict multi-value return mode,
// functions with inconsistent return counts (different numbers of values across return statements)
// are rejected at compile time.
func TestStrictMultiValueReturnInconsistent(t *testing.T) {
	opts := &syntax.FileOptions{
		StrictMultiValueReturn: true,
	}

	testCases := []struct {
		name string
		src  string
		want string // expected panic message substring
	}{
		{
			name: "one_vs_two_values",
			src: `
def inconsistent(x):
    if x:
        return 1, 2
    else:
        return 3
`,
			want: "multi-value return count mismatch: found 2 and 1 values",
		},
		{
			name: "two_vs_three_values",
			src: `
def inconsistent(mode):
    if mode == 1:
        return "a", "b"
    else:
        return "x", "y", "z"
`,
			want: "multi-value return count mismatch: found 2 and 3 values",
		},
		{
			name: "bare_return_vs_multi_value",
			src: `
def inconsistent(x):
    if x:
        return
    else:
        return 1, 2
`,
			want: "multi-value return count mismatch: found 1 and 2 values",
		},
		{
			name: "three_branches_inconsistent",
			src: `
def inconsistent(mode):
    if mode == "single":
        return "alone"
    elif mode == "pair":
        return "first", "second"
    else:
        return 1, 2, 3
`,
			want: "multi-value return count mismatch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := starlark.SourceProgramOptions(opts, "test.star", tc.src, func(string) bool { return false })
			if err == nil {
				t.Fatal("expected compilation error due to inconsistent return counts, but compilation succeeded")
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tc.want) {
				t.Errorf("error message %q does not contain expected substring %q", errMsg, tc.want)
			}
		})
	}
}
