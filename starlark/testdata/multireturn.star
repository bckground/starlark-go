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

# Test dynamic return counts (inconsistent across branches)
# These use tuple packing/unpacking with runtime validation
def dynamic_return(x):
    if x:
        return 1, 2
    else:
        return 3

# When called with True, returns 2 values
dyn_a, dyn_b = dynamic_return(True)
assert.eq(dyn_a, 1)
assert.eq(dyn_b, 2)

# When called with False, returns 1 value
dyn_c = dynamic_return(False)
assert.eq(dyn_c, 3)

# With strict multi-return mode, dynamic returns also enforce count matching
def test_dynamic_single_var():
    # Expects 1 value but gets 2 - should fail even for dynamic returns
    x = dynamic_return(True)

assert.fails(test_dynamic_single_var, "expected 1 value, got 2")

# Unpacking validates count with clear error messages
def test_dynamic_unpack_mismatch():
    # Expects 2 values but gets 1 (an int)
    x, y = dynamic_return(False)

assert.fails(test_dynamic_unpack_mismatch, "expected 2 values, got 1")

# More complex dynamic return
def complex_dynamic(mode):
    if mode == "single":
        return "alone"
    elif mode == "pair":
        return "first", "second"
    else:
        return 1, 2, 3

single_val = complex_dynamic("single")
assert.eq(single_val, "alone")

pair_a, pair_b = complex_dynamic("pair")
assert.eq(pair_a, "first")
assert.eq(pair_b, "second")

triple_x, triple_y, triple_z = complex_dynamic("triple")
assert.eq(triple_x, 1)
assert.eq(triple_y, 2)
assert.eq(triple_z, 3)

def array_return():
    return [1, 2]

def test_array_return_mismatch():
    a, b = array_return()

assert.fails(test_array_return_mismatch, "expected 2 values, got 1")

def test_array_return():
    single_array = array_return()
    assert.eq(single_array, [1, 2])

test_array_return()

def array_multi_value_return():
    return [1, 2], "foo"

def test_array_multi_return():
    single_array, s = array_multi_value_return()
    assert.eq(single_array, [1, 2])
    assert.eq(s, "foo")

test_array_multi_return()
