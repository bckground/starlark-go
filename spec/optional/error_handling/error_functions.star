# spec: spec.md#error-returning-functions

errs = error_tags("Bad")

# An error-returning function that returns an ordinary value behaves
# like a normal function call under try/catch.
def succeed()!:
    return 42

assert.eq(succeed() catch "fallback", 42)

def none_result()!:
    pass

assert.eq(none_result() catch "fallback", None)

# Returning an error tag or error value takes the error channel: the
# caller observes the error, not a result.
def fail_tag(flag)!:
    if flag:
        return errs.Bad
    return "ok"

assert.eq(fail_tag(True) catch "handled", "handled")
assert.eq(fail_tag(False) catch "handled", "ok")

# The error decision is dynamic: the same return statement may yield
# a value or an error depending on what it evaluates to.
def conditional(value)!:
    return value

assert.eq(conditional(7) catch "handled", 7)
assert.eq(conditional(errs.Bad) catch "handled", "handled")
