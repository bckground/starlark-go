# Tests of calls to error-returning functions through a dynamic target.

load("assert.star", "assert")

errors = error_tags("ErrMissing")

def read_all(path)!:
    return errors.ErrMissing(message="no such file: " + path)

def read_ok(path)!:
    return "contents of " + path

# An embedder-style module: the error-returning function is reached
# through an attribute, so the resolver cannot prove the call target
# error-returning and the requirement to handle the error is enforced
# at runtime instead.
fs = struct(read_all=read_all, read_ok=read_ok)

# Both catch forms and try work through an attribute target.
def test_catch_value_through_attr():
    assert.eq(fs.read_all("//x.yml") catch "fallback", "fallback")

test_catch_value_through_attr()

def test_catch_block_through_attr():
    x = fs.read_all("//x.yml") catch e:
        recover e.tag
    assert.eq(x, errors.ErrMissing)

test_catch_block_through_attr()

def propagates()!:
    return try fs.read_all("//x.yml")

def test_try_through_attr():
    assert.eq(propagates() catch "propagated", "propagated")

test_try_through_attr()

# An attribute call that returns no error is unaffected.
def test_success_through_attr():
    assert.eq(fs.read_ok("//x.yml"), "contents of //x.yml")

test_success_through_attr()

# An unhandled error from an attribute call is a failure, not a
# silently dropped error and a None result - both when the enclosing
# function returns...
def drops_on_return():
    x = fs.read_all("//x.yml")
    return "no failure"

assert.fails(drops_on_return, "ErrMissing: no such file")

# ...and when a later call would overwrite it.
def drops_on_later_call():
    x = fs.read_all("//x.yml")
    y = fs.read_ok("//y.yml")
    return "no failure"

assert.fails(drops_on_later_call, "ErrMissing: no such file")
