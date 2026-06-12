# spec: spec.md#catch

errs = error_tags("E")

def may_fail(flag)!:
    if flag:
        return errs.E
    return "value"

# Value form: the fallback replaces the result on error.
assert.eq(may_fail(True) catch "fallback", "fallback")
assert.eq(may_fail(False) catch "fallback", "value")

# The fallback expression is evaluated only on the error path, and
# may be any expression, including one that begins with an identifier
# (a name, call, index, or operator chain): only an identifier
# immediately followed by ':' selects the block form.
evals = []

def fallback_value(label):
    evals.append(label)
    return label

ok = may_fail(False) catch fallback_value("a")
assert.eq(ok, "value")
assert.eq(evals, [])
caught = may_fail(True) catch fallback_value("b")
assert.eq(caught, "b")
assert.eq(evals, ["b"])

default = "n/a"
assert.eq(may_fail(True) catch default, "n/a")
assert.eq(may_fail(True) catch default + "?", "n/a?")
assert.eq(may_fail(True) catch {"k": "v"}["k"], "v")

# Block form: the suite runs with the error value bound, and must end
# with recover or return.
def describe(flag)!:
    result = may_fail(flag) catch e:
        recover "caught " + e.message
    return result

assert.eq(describe(True) catch "outer", "caught E")
assert.eq(describe(False) catch "outer", "value")

# A catch block may end with return, returning from the enclosing
# function.
def bail(flag)!:
    result = may_fail(flag) catch e:
        return "returned"
    return "fell through: " + result

assert.eq(bail(True) catch "outer", "returned")
assert.eq(bail(False) catch "outer", "fell through: value")

# catch works in ordinary (non-error-returning) functions and at
# module level.
def plain():
    return may_fail(True) catch "plain-fallback"

assert.eq(plain(), "plain-fallback")
top = may_fail(True) catch "top-fallback"
assert.eq(top, "top-fallback")

# Falling off the end of a catch block is a failure.
def fall_off():
    x = may_fail(True) catch e:
        y = 1
    return x

assert.fails(fall_off, "catch block must end with recover or return")
