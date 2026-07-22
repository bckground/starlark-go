# spec: spec.md#failures

errs = error_tags("E")

# catch does not intercept failures: fail() aborts through a catch.
def fails_hard()!:
    fail("boom")
    return errs.E

def tries_to_catch():
    return fails_hard() catch "fallback"

assert.fails(tries_to_catch, "fail: boom")

# Implicit faults are failures too.
def divides()!:
    x = 1 // 0
    return errs.E

def catches_divide():
    return divides() catch "fallback"

assert.fails(catches_divide, "floored division by zero")

# A catch block does not run for a failure.
trace = []

def block_does_not_run():
    x = fails_hard() catch e:
        trace.append("ran")
        recover "r"
    return x

assert.fails(block_does_not_run, "fail: boom")
assert.eq(trace, [])

# try does not mask failures either.
def tries()!:
    return try fails_hard()

def catches_try():
    return tries() catch "fallback"

assert.fails(catches_try, "fail: boom")

# fail with a single error value turns a caught error into a failure
# carrying it; the abort message uses the tag's name. (This is the
# handling that a module-level try is specified to perform.)
def errors_softly()!:
    return errs.E(message="ignored by fail")

def fail_with_error():
    e = errors_softly() catch err:
        recover err
    fail(e)

assert.fails(fail_with_error, "fail: E")

# A bare error tag is wrapped, as in a ! function's return.
assert.fails(lambda: fail(errs.E), "fail: E")

# An error or error tag must be the sole argument: mixing it with
# message arguments, or passing several, is itself a failure.
def mixed():
    e = errors_softly() catch err:
        recover err
    fail("context:", e)

assert.fails(mixed, "an error or error tag must be the sole argument")
assert.fails(lambda: fail(errs.E, errs.E), "an error or error tag must be the sole argument")
assert.fails(lambda: fail("context:", errs.E), "an error or error tag must be the sole argument")
assert.fails(lambda: fail(errs.E, "trailing context"), "an error or error tag must be the sole argument")
