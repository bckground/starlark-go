# spec: spec.md#annotation-semantics

# Arguments of a deferred call are evaluated when the defer statement
# executes, but the callee's annotations are checked when the
# deferred call runs -- at function exit, after the body completes.
trace = []

def sink(x: int):
    trace.append(x)

def ok():
    defer sink(1)
    trace.append("body")
    return "done"

assert.eq(ok(), "done")
assert.eq(trace, ["body", 1])

# A deferred call whose eagerly-captured argument violates the
# annotation fails when the function exits, not at the defer
# statement; the body still runs to completion first.
body_ran = []

def bad():
    defer sink("mismatch")
    body_ran.append(True)
    return "unobserved"

assert.fails(bad, "does not match the type annotation `int` for argument `x`")
assert.eq(body_ran, [True])

# The check applies the value captured at the defer statement, not
# the variable's value at exit.
def captures():
    v = 7
    defer sink(v)
    v = "now a string"
    return v

assert.eq(captures(), "now a string")
assert.eq(trace, ["body", 1, 7])
