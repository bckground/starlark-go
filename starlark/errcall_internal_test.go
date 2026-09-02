package starlark

import (
	"testing"

	"go.starlark.net/internal/compile"
	"go.starlark.net/syntax"
)

// TestGuardOpcodeFollowsCall pins the codegen invariant that the interpreter's
// unguarded-call check relies on (see the CALL case in interp.go): a TRY or
// CATCH_CHECK produced by a try/catch expression is always emitted immediately
// after the call instruction it guards, so the interpreter can decide whether a
// call site is guarded by peeking at the next opcode, before running the callee.
//
// The invariant holds because the resolver requires the operand of try and catch
// to be a call expression (resolve.checkErrorCalls), so the compiler never emits
// anything between the operand's call and the guard.
func TestGuardOpcodeFollowsCall(t *testing.T) {
	// Every shape of guarded call: the four call forms, both catch forms, try
	// inside a ! function and at module level, direct and dynamic targets, and a
	// guarded call nested in an argument list, a comprehension, and a larger
	// expression.
	const src = `
def ok(*args, **kwargs)!: return "ok"
tbl = {"ok": ok}

def a()!: return try ok()
def b(): return ok() catch "d"
def c():
    x = tbl["ok"]() catch e:
        recover 1
    return x
def d()!: return (try ok()) + 1
def e()!: return try ok(1, k=2)
def g()!: return try ok(*[1], **{})
def h()!: return try ok(1, k=3, *[2], **{})
def i()!: return len(try ok())
def j()!: return [try ok() for _ in range(1)]
def k(f)!: return try f()
top = ok() catch "d"
`
	_, prog, err := SourceProgramOptions(syntax.LegacyFileOptions(), "guard.star", src, func(string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}

	funcodes := append([]*compile.Funcode{prog.compiled.Toplevel}, prog.compiled.Functions...)
	guards := 0
	for _, fn := range funcodes {
		var prev compile.Opcode
		var prevPC uint32
		for pc := uint32(0); int(pc) < len(fn.Code); {
			at := pc
			op := compile.Opcode(fn.Code[pc])
			pc++
			if op >= compile.OpcodeArgMin {
				for {
					b := fn.Code[pc]
					pc++
					if b < 0x80 {
						break
					}
				}
			}
			if op == compile.TRY || op == compile.CATCH_CHECK {
				guards++
				switch prev {
				case compile.CALL, compile.CALL_VAR, compile.CALL_KW, compile.CALL_VAR_KW:
					// ok
				default:
					t.Errorf("%s: %s at pc=%d is preceded by %s at pc=%d, want a call instruction",
						fn.Name, op, at, prev, prevPC)
				}
			}
			prev, prevPC = op, at
		}
	}
	if want := 11; guards != want {
		t.Errorf("found %d try/catch_check instructions, want %d (did the test source stop compiling as intended?)", guards, want)
	}
}
