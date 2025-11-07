# Tests for multi-return values feature
# This file tests the true multi-return semantics (not tuple packing)
# option:multireturn

load("assert.star", "assert")

# Basic multi-return with 2 values
def two_values():
    return 1, 2

a, b = two_values()
assert.eq(a, 1)
assert.eq(b, 2)

# Multi-return with 3 values
def three_values():
    return "x", "y", "z"

x, y, z = three_values()
assert.eq(x, "x")
assert.eq(y, "y")
assert.eq(z, "z")

# Single return value
def one_value():
    return 42

val = one_value()
assert.eq(val, 42)

# Return with no explicit value (returns None)
def no_value():
    return

none_val = no_value()
assert.eq(none_val, None)

# Test mismatch errors: too few variables
def test_too_few():
    def f():
        return 1, 2
    x = f()  # Should error: expected 1 value, got 2

assert.fails(test_too_few, "expected 1 value, got 2")

# Test mismatch errors: too many variables
def test_too_many():
    def f():
        return 1, 2
    a, b, c = f()  # Should error: expected 3 values, got 2

assert.fails(test_too_many, "expected 3 values, got 2")

# Test mismatch errors: one variable for multi-return
def test_one_var():
    a, b = two_values()
    c = two_values()  # Should error: expected 1 value, got 2

assert.fails(test_one_var, "expected 1 value, got 2")

# Multi-return in nested calls
def swap(x, y):
    return y, x

def pair():
    return 10, 20

p, q = swap(*pair())
assert.eq(p, 20)
assert.eq(q, 10)

# Multi-return with different types
def mixed_types():
    return 1, "two", 3.0, True, None

int_val, str_val, float_val, bool_val, none_val2 = mixed_types()
assert.eq(int_val, 1)
assert.eq(str_val, "two")
assert.eq(float_val, 3.0)
assert.eq(bool_val, True)
assert.eq(none_val2, None)

# Note: Inconsistent return counts (e.g., returning different numbers of values
# in different branches) are now caught at compile time and will cause a
# compilation error.

print("All multi-return tests passed!")
