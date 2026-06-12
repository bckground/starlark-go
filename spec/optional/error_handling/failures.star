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
