load("assert.star", "assert")

errors = error("ErrInner", "ErrOuter")

# Test try propagates through multiple levels.
def test_deep_propagation():
    def level3()!:
        return errors.ErrInner

    def level2()!:
        return try level3()

    def level1()!:
        return try level2()

    result = level1() catch "caught"
    assert.eq(result, "caught")

test_deep_propagation()

# Test try propagation stops at catch.
def test_propagation_stops_at_catch():
    def inner()!:
        return errors.ErrInner

    def middle()!:
        # Catch here stops propagation.
        return inner() catch "handled_in_middle"

    def outer()!:
        return try middle()

    result = outer() catch "handled_in_outer"
    # Error was caught in middle, so outer gets "handled_in_middle".
    assert.eq(result, "handled_in_middle")

test_propagation_stops_at_catch()

# Test try propagation with non-error-returning function.
def test_try_non_error_function():
    def normal_func():
        return "value"

    def caller()!:
        # Calling non-! function with try should work (no error to propagate).
        x = try normal_func()
        return x

    result = caller() catch "error"
    assert.eq(result, "value")

test_try_non_error_function()

# Test error propagates through try chain.
def test_error_chain():
    def step1()!:
        return errors.ErrInner

    def step2()!:
        a = try step1()
        return a + "_modified"  # Should not reach here.

    def step3()!:
        b = try step2()
        return b + "_more"  # Should not reach here.

    result = step3() catch "caught"
    assert.eq(result, "caught")

test_error_chain()

# Test successful propagation through chain.
def test_success_chain():
    def step1()!:
        return "one"

    def step2()!:
        a = try step1()
        return a + "_two"

    def step3()!:
        b = try step2()
        return b + "_three"

    result = step3() catch "error"
    assert.eq(result, "one_two_three")

test_success_chain()

# Test mixing try and catch in chain.
def test_mixed_try_catch():
    def may_fail(fail)!:
        if fail:
            return errors.ErrInner
        return "ok"

    def process()!:
        a = may_fail(False) catch "a_fallback"
        b = may_fail(True) catch "b_fallback"
        c = try may_fail(False)
        return a + " " + b + " " + c

    result = process() catch "error"
    assert.eq(result, "ok b_fallback ok")

test_mixed_try_catch()
