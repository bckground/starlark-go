# Tests for multi-return values feature
# This file tests the true multi-return semantics (not tuple packing)
# option:strictmultivaluereturn

load("assert.star", "assert")

# Helper functions used across tests
def two_values():
    return 1, 2

def three_values():
    return "x", "y", "z"

def one_value():
    return 42

def no_return():
    # This function returns None.
    42

def no_value():
    return

def swap(x, y):
    return y, x

def pair():
    return 10, 20

def mixed_types():
    return 1, "two", 3.0, True, None

def dynamic_return(x):
    if x:
        return 1, 2
    else:
        return 3

def complex_dynamic(mode):
    if mode == "single":
        return "alone"
    elif mode == "pair":
        return "first", "second"
    else:
        return 1, 2, 3

def array_return():
    return [1, 2]

def array_multi_value_return():
    return [1, 2], "foo"

# Test basic multi-return with 2 values
def test_two_values():
    a, b = two_values()
    assert.eq(a, 1)
    assert.eq(b, 2)

test_two_values()

# Test multi-return with 3 values
def test_three_values():
    x, y, z = three_values()
    assert.eq(x, "x")
    assert.eq(y, "y")
    assert.eq(z, "z")

test_three_values()

# Test single return value
def test_single_return():
    val = one_value()
    assert.eq(val, 42)

test_single_return()

# Test return with no explicit value (returns None)
def test_no_value_return():
    none_val = no_value()
    assert.eq(none_val, None)

test_no_value_return()

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

# Test multi-return in nested calls with unpacking
def test_nested_calls():
    p, q = swap(*pair())
    assert.eq(p, 20)
    assert.eq(q, 10)

test_nested_calls()

# Test multi-return with different types
def test_mixed_types():
    int_val, str_val, float_val, bool_val, none_val = mixed_types()
    assert.eq(int_val, 1)
    assert.eq(str_val, "two")
    assert.eq(float_val, 3.0)
    assert.eq(bool_val, True)
    assert.eq(none_val, None)

test_mixed_types()

# Test dynamic return counts (when called with True, returns 2 values)
def test_dynamic_return_two():
    a, b = dynamic_return(True)
    assert.eq(a, 1)
    assert.eq(b, 2)

test_dynamic_return_two()

# Test dynamic return counts (when called with False, returns 1 value)
def test_dynamic_return_one():
    c = dynamic_return(False)
    assert.eq(c, 3)

test_dynamic_return_one()

# Test strict multi-return mode enforces count matching for dynamic returns
def test_dynamic_single_var():
    # Expects 1 value but gets 2 - should fail even for dynamic returns
    x = dynamic_return(True)

assert.fails(test_dynamic_single_var, "expected 1 value, got 2")

# Test unpacking validates count with clear error messages
def test_dynamic_unpack_mismatch():
    # Expects 2 values but gets 1 (an int)
    x, y = dynamic_return(False)

assert.fails(test_dynamic_unpack_mismatch, "expected 2 values, got 1")

# Test complex dynamic return with multiple branches
def test_complex_dynamic():
    single_val = complex_dynamic("single")
    assert.eq(single_val, "alone")

    pair_a, pair_b = complex_dynamic("pair")
    assert.eq(pair_a, "first")
    assert.eq(pair_b, "second")

    triple_x, triple_y, triple_z = complex_dynamic("triple")
    assert.eq(triple_x, 1)
    assert.eq(triple_y, 2)
    assert.eq(triple_z, 3)

test_complex_dynamic()

# Test that arrays are single unpackable values in strict mode
def test_array_return_mismatch():
    a, b = array_return()

assert.fails(test_array_return_mismatch, "expected 2 values, got 1")

# Test that arrays can be assigned to a single variable
def test_array_return():
    single_array = array_return()
    assert.eq(single_array, [1, 2])

test_array_return()

# Test multi-return with array as one of the values
def test_array_multi_return():
    single_array, s = array_multi_value_return()
    assert.eq(single_array, [1, 2])
    assert.eq(s, "foo")

test_array_multi_return()

# Test no assignments
def test_no_assignment():
  no_value()
  no_return()
  one_value()
  two_values()
  three_values()

test_no_assignment()
