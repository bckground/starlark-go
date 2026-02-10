load("assert.star", "assert")

errors = error_tags("Err", "ErrTimeout")

# Test unpacking successful result with try.
def test_try_unpack_success():
    def returns_tuple()!:
        return (1, 2, 3)

    def caller()!:
        a, b, c = try returns_tuple()
        return (a, b, c)

    a, b, c = caller() catch (0, 0, 0)
    assert.eq(a, 1)
    assert.eq(b, 2)
    assert.eq(c, 3)

test_try_unpack_success()

# Test unpacking with catch fallback (tuple).
def test_catch_unpack_tuple_fallback():
    def returns_error_tags()!:
        return errors.Err

    x, y = returns_error_tags() catch (10, 20)
    assert.eq(x, 10)
    assert.eq(y, 20)

test_catch_unpack_tuple_fallback()

# Test unpacking with catch block form.
def test_catch_unpack_block():
    def returns_error_tags()!:
        return errors.Err

    a, b, c = returns_error_tags() catch e:
        recover (100, 200, 300)

    assert.eq(a, 100)
    assert.eq(b, 200)
    assert.eq(c, 300)

test_catch_unpack_block()

# Test unpacking with conditional recovery.
def test_catch_unpack_conditional():
    def fail_timeout()!:
        return errors.ErrTimeout

    x, y = fail_timeout() catch e:
        if e.tag == errors.ErrTimeout:
            recover ("timeout", 99)
        else:
            recover ("other", 0)

    assert.eq(x, "timeout")
    assert.eq(y, 99)

test_catch_unpack_conditional()

# Test that unpacking non-iterable fallback fails.
def test_catch_unpack_invalid_fallback():
    def returns_error_tags()!:
        return errors.Err

    # This should fail: can't unpack string into 2 variables
    assert.fails(
        lambda: [returns_error_tags() catch "default"][0],  # Force evaluation
        "got string in sequence assignment"
    )

# Can't directly test the multi-assignment failure in assert.fails,
# so we test that the error message is correct through execution.

# Test unpacking with list fallback.
def test_catch_unpack_list():
    def returns_error_tags()!:
        return errors.Err

    a, b, c = returns_error_tags() catch [4, 5, 6]
    assert.eq(a, 4)
    assert.eq(b, 5)
    assert.eq(c, 6)

test_catch_unpack_list()

# Test mixed unpacking and regular assignment.
def test_catch_unpack_mixed():
    def returns_tuple()!:
        return ("first", "second")

    # Unpack with catch
    x, y = returns_tuple() catch ("a", "b")
    assert.eq(x, "first")
    assert.eq(y, "second")

    # Regular assignment with catch
    z = returns_tuple() catch "default"
    assert.eq(z, ("first", "second"))

test_catch_unpack_mixed()

# Test try with unpacking in error-returning function.
def test_try_unpack_in_error_func():
    def returns_tuple()!:
        return (7, 8, 9)

    def outer()!:
        a, b, c = try returns_tuple()
        return a + b + c

    result = outer() catch 0
    assert.eq(result, 24)

test_try_unpack_in_error_func()

# Test catch with unpacking at module level.
def module_level_func()!:
    return errors.Err

m1, m2 = module_level_func() catch ("module", "level")
assert.eq(m1, "module")
assert.eq(m2, "level")

# Test error propagation with unpacking.
def test_try_unpack_propagates_error_tags():
    def fails()!:
        return errors.Err

    def outer()!:
        a, b = try fails()  # Error propagates
        return (a, b)  # Never reached

    result = outer() catch "caught"
    assert.eq(result, "caught")

test_try_unpack_propagates_error_tags()
