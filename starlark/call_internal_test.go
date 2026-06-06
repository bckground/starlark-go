package starlark

import (
	"errors"
	"testing"
)

// TestCallInBuiltinLeavesFrameClean verifies the structural property:
// when a Go builtin invokes a !-function via Call and that function
// returns an error, the error is delivered to the builtin on Call's
// error result (as a *ReturnedError) and is NOT deposited on the
// builtin's frame. Because the frame stays clean, an error the
// builtin ignores cannot implicitly propagate when the builtin
// returns.
func TestCallInBuiltinLeavesFrameClean(t *testing.T) {
	tag := NewErrorTag("E")
	errfn := NewBuiltinCanReturnError("errfn", func(*Thread, *Builtin, Tuple, []Tuple) (Value, error) {
		return tag, nil // returns an error value
	})

	var (
		framePendingAfterCall *Error // the builtin's own frame.pendingError after Call
		innerErr              error  // what Call returned to the builtin
		innerVal              Value
	)
	// probe ignores the error errfn returns and returns its own clean value.
	probe := NewBuiltin("probe", func(th *Thread, _ *Builtin, _ Tuple, _ []Tuple) (Value, error) {
		innerVal, innerErr = Call(th, errfn, nil, nil)
		framePendingAfterCall = th.frameAt(0).pendingError // probe's frame
		return String("clean"), nil
	})

	outerVal, outerErr := Call(&Thread{}, probe, nil, nil)

	// 1. The error was delivered to the builtin via the error channel, not lost.
	var re *ReturnedError
	if !errors.As(innerErr, &re) {
		t.Errorf("Call inside builtin returned (%v, %v); want a ReturnedError", innerVal, innerErr)
	} else if re.Value.Tag() != tag {
		t.Errorf("ReturnedError tag = %v, want E", re.Value.Tag())
	}

	// 2. The error was NOT deposited on the builtin's frame.
	if framePendingAfterCall != nil {
		t.Errorf("builtin frame.pendingError = %v after Call; want nil (no deposit on a non-Function caller)", framePendingAfterCall)
	}

	// 3. The ignored error did not implicitly propagate out of the builtin: the
	//    outer Call sees the builtin's clean value and no error.
	if outerErr != nil {
		t.Errorf("outer Call err = %v; want nil (an ignored error must not leak past the builtin)", outerErr)
	}
	if outerVal != String("clean") {
		t.Errorf("outer Call value = %v; want \"clean\"", outerVal)
	}
}
