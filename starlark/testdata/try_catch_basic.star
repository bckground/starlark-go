load("assert.star", "assert")

errors = error("ErrFail", "ErrTimeout")

# Test function that can return an error.
def may_fail(should_fail)!:
    if should_fail:
        return errors.ErrFail
    return "success"

# Test catch with value form - error case.
def test_catch_value_error():
    result = may_fail(True) catch "default"
    assert.eq(result, "default")

test_catch_value_error()

# Test catch with value form - success case.
def test_catch_value_success():
    result = may_fail(False) catch "default"
    assert.eq(result, "success")

test_catch_value_success()

# Test catch with numeric fallback.
def test_catch_numeric_value():
    result = may_fail(True) catch 42
    assert.eq(result, 42)

test_catch_numeric_value()

# Test catch with expression as fallback.
def test_catch_expression_value():
    result = may_fail(True) catch (10 + 20)
    assert.eq(result, 30)

test_catch_expression_value()

# Test nested catch expressions.
def test_nested_catch():
    def inner()!:
        return errors.ErrFail
    def outer()!:
        return inner() catch "inner_caught"

    result = outer() catch "outer_caught"
    # Inner catch should handle it.
    assert.eq(result, "inner_caught")

test_nested_catch()

# Test catch doesn't trigger on success.
def test_no_catch_on_success():
    call_count = [0]
    def fallback_fn():
        call_count[0] = call_count[0] + 1
        return "fallback"

    # This should succeed, fallback should not be evaluated.
    result = may_fail(False) catch "should_not_see_this"
    assert.eq(result, "success")

test_no_catch_on_success()

# Test try keyword propagates error.
def test_try_propagates():
    def inner()!:
        return errors.ErrFail

    def outer()!:
        x = try inner()
        return x

    result = outer() catch "caught"
    assert.eq(result, "caught")

test_try_propagates()

# Test try with successful call.
def test_try_success():
    def inner()!:
        return "value"

    def outer()!:
        x = try inner()
        return x

    result = outer() catch "caught"
    assert.eq(result, "value")

test_try_success()

# Test multiple try in sequence.
def test_multiple_try():
    def call1()!:
        return "first"
    def call2()!:
        return "second"

    def outer()!:
        a = try call1()
        b = try call2()
        return a + " " + b

    result = outer() catch "error"
    assert.eq(result, "first second")

test_multiple_try()

# Test try stops on first error.
def test_try_stops_on_error():
    call_log = []

    def call1()!:
        call_log.append("call1")
        return errors.ErrFail

    def call2()!:
        call_log.append("call2")
        return "value"

    def outer()!:
        a = try call1()
        b = try call2()  # Should not be called.
        return b

    result = outer() catch "caught"
    assert.eq(result, "caught")
    # Only call1 should have been called.
    assert.eq(call_log, ["call1"])

test_try_stops_on_error()
