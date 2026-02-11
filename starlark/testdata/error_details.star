load("assert.star", "assert")

# Test ErrorTag is callable with kwargs.
def test_error_tag_callable():
    tags = error_tags("Err")
    err = tags.Err(message="something went wrong")
    assert.eq(err.tag, tags.Err)
    assert.eq(err.message, "something went wrong")
    assert.eq(err.cause, None)
    assert.eq(err.details, [])

test_error_tag_callable()

# Test ErrorTag callable with no args creates bare error.
def test_error_tag_no_args():
    tags = error_tags("Err")
    err = tags.Err()
    assert.eq(err.tag, tags.Err)
    assert.eq(err.message, "Err")  # Default message is tag name
    assert.eq(err.cause, None)
    assert.eq(err.details, [])

test_error_tag_no_args()

# Test error with details.
def test_error_details():
    tags = error_tags("Err")
    err = tags.Err(message="with details", details=[1, "two", 3])
    assert.eq(err.details, [1, "two", 3])

test_error_details()

# Test default message is tag name.
def test_default_message():
    tags = error_tags("MyError")
    err = tags.MyError()
    assert.eq(err.message, "MyError")

test_default_message()

# Test str(err) returns tag name.
def test_str_returns_tag_name():
    tags = error_tags("Err")
    err = tags.Err(message="custom message")
    assert.eq(str(err), "Err")

test_str_returns_tag_name()

# Test error chain: cause linking.
def test_error_chain():
    tags = error_tags("ErrA", "ErrB")
    inner = tags.ErrA(message="inner error")
    outer = tags.ErrB(message="outer error", cause=inner)
    assert.eq(outer.tag, tags.ErrB)
    assert.eq(outer.message, "outer error")
    assert.eq(outer.cause.tag, tags.ErrA)
    assert.eq(outer.cause.message, "inner error")
    assert.eq(outer.cause.cause, None)

test_error_chain()

# Test catch block gives *Error, access .tag for comparison.
def test_catch_gives_error():
    tags = error_tags("ErrA")

    def may_fail()!:
        return tags.ErrA

    result = may_fail() catch e:
        assert.eq(type(e), "error")
        assert.eq(e.tag, tags.ErrA)
        recover "caught"

    assert.eq(result, "caught")

test_catch_gives_error()

# Test implicit wrapping: returning ErrorTag from ! function wraps it.
def test_implicit_wrapping():
    tags = error_tags("Err")

    def may_fail()!:
        return tags.Err  # Returns ErrorTag, gets wrapped

    result = may_fail() catch e:
        assert.eq(type(e), "error")
        assert.eq(e.tag, tags.Err)
        assert.eq(e.message, "Err")  # Default
        recover "ok"

    assert.eq(result, "ok")

test_implicit_wrapping()

# Test explicit error creation in ! function.
def test_explicit_error_creation():
    tags = error_tags("Err")

    def may_fail()!:
        return tags.Err(message="explicit msg")

    result = may_fail() catch e:
        assert.eq(e.message, "explicit msg")
        recover "ok"

    assert.eq(result, "ok")

test_explicit_error_creation()

# Test error with cause chain through catch.
def test_cause_chain_through_catch():
    tags = error_tags("Inner", "Outer")

    def inner()!:
        return tags.Inner

    def outer()!:
        inner() catch e:
            return tags.Outer(message="wrapped", cause=e)
        return "ok"

    result = outer() catch e:
        assert.eq(e.tag, tags.Outer)
        assert.eq(e.message, "wrapped")
        assert.eq(e.cause.tag, tags.Inner)
        assert.eq(e.cause.message, "Inner")
        recover "handled"

    assert.eq(result, "handled")

test_cause_chain_through_catch()
