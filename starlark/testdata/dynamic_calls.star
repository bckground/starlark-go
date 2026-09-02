# Tests of calls to error-returning functions through a dynamic target.

load("assert.star", "assert")

errors = error_tags("ErrMissing")

def read_all(path)!:
    return errors.ErrMissing(message="no such file: " + path)

def read_ok(path)!:
    return "contents of " + path

# An embedder-style module: the error-returning function is reached
# through an attribute, so the resolver cannot prove the call target
# error-returning and the requirement to guard the call with try or
# catch is enforced at runtime instead.
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

# A guarded attribute call that returns no error yields its value.
def test_success_through_attr():
    assert.eq(fs.read_ok("//x.yml") catch "unexpected", "contents of //x.yml")

test_success_through_attr()

# The requirement is on the call site, not on the outcome: an
# unguarded attribute call is a failure even when the callee would
# have returned a value rather than an error.
def unguarded_success():
    x = fs.read_ok("//x.yml")
    return "no failure"

assert.fails(unguarded_success, 'call to error-returning function "read_ok" must be handled with try or catch')

# The callee does not run: the failure is raised at the call site.
ran = []

def records(path)!:
    ran.append(path)
    return "ok"

rec = struct(records=records)

def unguarded_does_not_run():
    x = rec.records("//x.yml")
    return "no failure"

assert.fails(unguarded_does_not_run, "must be handled with try or catch")
assert.eq(ran, [])

# An unhandled error from an attribute call is a failure, not a
# silently dropped error and a None result - both when the enclosing
# function returns...
def drops_on_return():
    x = fs.read_all("//x.yml")
    return "no failure"

assert.fails(drops_on_return, 'call to error-returning function "read_all" must be handled with try or catch')

# ...and when a later call would overwrite it.
def drops_on_later_call():
    x = fs.read_all("//x.yml")
    y = fs.read_ok("//y.yml")
    return "no failure"

assert.fails(drops_on_later_call, 'call to error-returning function "read_all" must be handled with try or catch')
