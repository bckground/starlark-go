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
    # Any function with no explicit return statement returns None. Therefore,
    # this function returns None.
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

def list_return():
    return [1, 2]

def list_multi_value_return():
    return [1, 2], "foo"

def tuple_return():
    t = (1, 2)
    return t

def tuple_multi_value_return():
    t = (1, 2)
    return t, "foo"

# Test bare return
def test_bare_return():
    v = no_value()
    assert.eq(v, None)

test_bare_return()

# Test no return
def test_no_return():
    v = no_return()
    assert.eq(v, None)

test_no_return()

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

# Test that lists cannot be implicitly unpacked when returned from functions
def test_list_return_implicit_fail():
    a, b = list_return()

assert.fails(test_list_return_implicit_fail, "expected 2 values, got 1")

# Test that lists can be explicitly unpackd with brackets
def test_list_return_explicit():
    [a, b] = list_return()
    assert.eq(a, 1)
    assert.eq(b, 2)

test_list_return_explicit()

# Test that lists can be assigned to a single variable
def test_list_return():
    single_list = list_return()
    assert.eq(single_list, [1, 2])

test_list_return()

# Test multi-return with list as one of the values
def test_list_multi_return():
    single_list, s = list_multi_value_return()
    assert.eq(single_list, [1, 2])
    assert.eq(s, "foo")

test_list_multi_return()

# Test that tuples cannot be implicitly unpacked when returned from functions
def test_tuple_return_implicit_fail():
    a, b = tuple_return()

assert.fails(test_tuple_return_implicit_fail, "expected 2 values, got 1")

# Test that tuples can be explicitly unpack with parentheses
def test_tuple_return_explicit_parens():
    (a, b) = tuple_return()
    assert.eq(a, 1)
    assert.eq(b, 2)

test_tuple_return_explicit_parens()

# Test that tuples can be assigned to a single variable
def test_tuple_return():
    single_tuple = tuple_return()
    assert.eq(single_tuple, (1, 2))

test_tuple_return()

# Test multi-return with tuple as one of the values
def test_tuple_multi_return():
    single_tuple, s = tuple_multi_value_return()
    assert.eq(single_tuple, (1, 2))
    assert.eq(s, "foo")

test_tuple_multi_return()

# Test no assignments
def test_no_assignment():
    no_value()
    no_return()
    one_value()
    two_values()
    three_values()

test_no_assignment()

def test_explicit_assignment_unpack():
    (a, b, c,) = (1, 2, 3) # trailing comma ok
    (d, e, f,) = [1, 2, 3] # trailing comma ok
    [g, h, i,] = (1, 2, 3) # trailing comma ok
    [j, k, l,] = [1, 2, 3] # trailing comma ok

test_explicit_assignment_unpack()

def test_implicit_tuple_unpack():
    a, b  = (1, 2)

assert.fails(test_implicit_tuple_unpack, "expected 2 values, got 1")

def test_implicit_list_unpack():
    a, b = [1, 2]

assert.fails(test_implicit_list_unpack, "expected 2 values, got 1")
