# spec: spec.md#error-returning-functions

# An error-returning function may carry annotations; the ! marker
# precedes the return annotation.
errs = error_tags("E")

def lookup(id: int)! -> str:
    if id < 0:
        return errs.E
    return "user " + str(id)

# The return annotation describes the success value; returning an
# error skips the return check.
assert.eq(lookup(42) catch "missing", "user 42")
assert.eq(lookup(-1) catch "missing", "missing")

# Argument annotations are still enforced.
def call_badly():
    return lookup("nope") catch "missing"

assert.fails(call_badly, "does not match the type annotation `int` for argument")

# A mismatched success value is still a failure.
def bad_success(flag)! -> str:
    if flag:
        return errs.E
    return 7

def call_bad_success():
    return bad_success(False) catch "fallback"

assert.eq(bad_success(True) catch "fallback", "fallback")
assert.fails(call_bad_success, "does not match the type annotation `str` for return type")

# Type-annotation mismatches are failures: catch cannot intercept
# them.
def checked(flag)!:
    x: int = "bad" if flag else 0
    return errs.E

def tries_to_catch():
    return checked(True) catch "fallback"

assert.fails(tries_to_catch, "does not match the type annotation")
assert.eq(checked(False) catch "fallback", "fallback")
