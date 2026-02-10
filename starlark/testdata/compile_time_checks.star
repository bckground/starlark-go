load("assert.star", "assert")

# Test compile-time validation of error handling.
# Most tests here should fail at compile/resolve time, documented with comments.

errors = error_tags("Err")

# Valid: Calling ! function with try.
def test_valid_try():
    def may_fail()!:
        return "value"

    def caller()!:
        x = try may_fail()
        return x

    result = caller() catch "error"
    assert.eq(result, "value")

test_valid_try()

# Valid: Calling ! function with catch.
def test_valid_catch():
    def may_fail()!:
        return errors.Err

    def caller():
        x = may_fail() catch "default"
        return x

    result = caller()
    assert.eq(result, "default")

test_valid_catch()

# Valid: ! function calling another ! function with try (propagates).
def test_valid_propagation():
    def inner()!:
        return errors.Err

    def outer()!:
        x = try inner()
        return x

    result = outer() catch "caught"
    assert.eq(result, "caught")

test_valid_propagation()

# Valid: Non-! function can be called without error handling.
def test_valid_normal_function():
    def normal():
        return "value"

    def caller():
        x = normal()
        return x

    result = caller()
    assert.eq(result, "value")

test_valid_normal_function()

# INVALID: Calling ! function without try/catch in non-! function.
# This should fail at compile time with error like:
# "must handle error from function may_fail with try/catch"
#
# def test_invalid_unhandled():
#     def may_fail()!:
#         return errors.Err
#
#     def caller():  # Non-! function.
#         x = may_fail()  # ERROR: Unhandled error-returning function call.
#         return x

# NOTE: Using try in non-! function is now a compile-time error.
# This is tested in resolve/testdata/resolve.star.

# INVALID: Using errdefer in non-! function.
# def test_invalid_errdefer_in_normal():
#     def cleanup():
#         pass
#
#     def caller():  # Non-! function.
#         errdefer cleanup()  # ERROR: errdefer only in ! functions.
#         return "value"

# INVALID: Using recover outside catch block.
# def test_invalid_recover_outside_catch():
#     def caller()!:
#         recover "value"  # ERROR: recover only in catch blocks.
#         return "value"

# INVALID: Using recover in regular defer or errdefer.
# def test_invalid_recover_in_defer():
#     def may_fail()!:
#         errdefer recover "value"  # ERROR: recover only in catch blocks.
#         return errors.Err

# Valid: Calling ! builtin function (if any) with try/catch.
# (Assuming we mark some builtins as !).
# def test_valid_builtin_error_handling():
#     # Hypothetical: open() is a ! builtin.
#     def caller()!:
#         file = try open("file.txt")
#         return file

# Note: Most invalid tests are commented out because they should cause
# compile-time errors. The test framework would need to be enhanced
# to test for expected compilation failures.

# Test that error handling is properly required.
def test_error_handling_required():
    def may_fail()!:
        return errors.Err

    # This should work: catch handles the error.
    result = may_fail() catch "handled"
    assert.eq(result, "handled")

    # This should work: calling from ! function with try propagates.
    def propagate()!:
        return try may_fail()

    result2 = propagate() catch "caught"
    assert.eq(result2, "caught")

test_error_handling_required()

# Test that ! is required for functions that return errors.
def test_bang_required_for_error_return():
    # This should work: ! function can return errors.
    def can_return_error_tags()!:
        return errors.Err

    result = can_return_error_tags() catch "caught"
    assert.eq(result, "caught")

    # INVALID: Non-! function returning error.
    # def cannot_return_error_tags():
    #     return errors.Err  # ERROR: non-! function cannot return errors.

test_bang_required_for_error_return()
