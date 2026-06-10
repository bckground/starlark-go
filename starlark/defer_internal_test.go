package starlark

import (
	"fmt"
	"strings"
	"testing"
)

func TestRunDeferredErrorIgnored(t *testing.T) {
	thread := &Thread{}
	fr := &frame{callable: NewBuiltin("test", nil)} // the frame being torn down
	thread.stack = append(thread.stack, fr)
	tag := NewErrorTag("E")
	errFn := NewBuiltinCanReturnError("errfn", func(_ *Thread, _ *Builtin, _ Tuple, _ []Tuple) (Value, error) {
		return tag, nil // returns an error: Call transfers it onto the caller frame (fr)
	})
	trigger := NewError(NewErrorTag("TRIG"), nil, nil, nil)
	fr.pendingError = trigger

	got := thread.runDeferred(fr, []deferredCall{{fn: errFn}}, nil)
	if got != nil {
		t.Errorf("got failure %v, want nil (error from cleanup must be ignored)", got)
	}
	if fr.pendingError != trigger {
		t.Errorf("fr.pendingError = %v, want the preserved trigger", fr.pendingError)
	}
}

func TestRunDeferredFailuresChainedAllRun(t *testing.T) {
	thread := &Thread{}
	fr := &frame{callable: NewBuiltin("test", nil)}
	thread.stack = append(thread.stack, fr)
	var ran []string
	mk := func(tag string) *Builtin {
		return NewBuiltin(tag, func(_ *Thread, _ *Builtin, _ Tuple, _ []Tuple) (Value, error) {
			ran = append(ran, tag)
			return nil, fmt.Errorf("e-%s", tag)
		})
	}
	stack := []deferredCall{{fn: mk("A")}, {fn: mk("B")}}
	got := thread.runDeferred(fr, stack, fmt.Errorf("trigger"))
	if strings.Join(ran, ",") != "B,A" {
		t.Errorf("ran = %v, want [B A] (LIFO, all run)", ran)
	}
	msg := got.Error()
	for _, want := range []string{"trigger", "e-B", "e-A"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q (all failures must be chained)", msg, want)
		}
	}
}
