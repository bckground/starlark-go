# Tests for strcit multi-value return mode.

# option:strictmultivaluereturn

load("assert.star", "assert")

# Helper functions used across tests.
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

def bare_return():
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

def test_no_return():
    v = no_return()
    assert.eq(v, None)

test_no_return()

def test_bare_return():
    v = bare_return()
    assert.eq(v, None)

test_bare_return()

def test_return_one_value():
    val = one_value()
    assert.eq(val, 42)

test_return_one_value()

# Test that implicit unpacking works for true multi-return values
def test_implicit_unpack_two():
    a, b = two_values()  # Implicit unpacking works for multiValue
    assert.eq(a, 1)
    assert.eq(b, 2)

test_implicit_unpack_two()

def test_implicit_unpack_three():
    x, y, z = three_values()  # Implicit unpacking works for multiValue
    assert.eq(x, "x")
    assert.eq(y, "y")
    assert.eq(z, "z")

test_implicit_unpack_three()

# Test that explicit unpacking also works for multi-return values
def test_explicit_unpack_two_parens():
    (a, b) = two_values()
    assert.eq(a, 1)
    assert.eq(b, 2)

test_explicit_unpack_two_parens()

def test_explicit_unpack_two_brackets():
    [a, b] = two_values()
    assert.eq(a, 1)
    assert.eq(b, 2)

test_explicit_unpack_two_brackets()

def test_explicit_unpack_three_parens():
    (x, y, z) = three_values()
    assert.eq(x, "x")
    assert.eq(y, "y")
    assert.eq(z, "z")

test_explicit_unpack_three_parens()

def test_explicit_unpack_three_brackets():
    [x, y, z] = three_values()
    assert.eq(x, "x")
    assert.eq(y, "y")
    assert.eq(z, "z")

test_explicit_unpack_three_brackets()

def test_too_few_variables_fails():
    def f():
        return 1, 2
    x = f()  # Should error: expected 1 value, got 2

assert.fails(test_too_few_variables_fails, "expected 1 value, got 2")

def test_too_many_variables_fails():
    def f():
        return 1, 2
    a, b, c = f()  # Should error: expected 3 values, got 2

assert.fails(test_too_many_variables_fails, "expected 3 values, got 2")

# Test mismatch errors: one variable for multi-return
def test_one_var():
    a, b = two_values()  # Implicit unpacking works
    c = two_values()  # Should error: expected 1 value, got 2

assert.fails(test_one_var, "expected 1 value, got 2")

# Test that splatting multiValue is not allowed
def test_nested_calls_splat_fails():
    p, q = swap(*pair())  # *pair() should fail - can't splat multiValue

# pair() returns multiValue which is not iterable
assert.fails(test_nested_calls_splat_fails, "argument after \\* must be iterable, not a multi-value return")

# Test multi-return in nested calls
def test_nested_calls():
    p, q = pair()  # Implicit unpacking works
    r, s = swap(p, q)  # Then pass as separate args
    assert.eq(r, 20)
    assert.eq(s, 10)

test_nested_calls()

# Test multi-return with different types using implicit unpacking
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

# Test that tuples from functions CANNOT be implicitly unpacked.
def test_implicit_tuple_unpack_return_fails():
    a, b = tuple_return()

assert.fails(test_implicit_tuple_unpack_return_fails, "expected 2 values, got 1")

# Test that lists from functions CANNOT be implicitly unpacked.
def test_implicit_list_unpack_return_fails():
    a, b = list_return()

assert.fails(test_implicit_list_unpack_return_fails, "expected 2 values, got 1")

# Test that tuples can be explicitly unpacked.
def test_explicit_unpack_return():
    (a, b) = tuple_return()
    assert.eq(a, 1)
    assert.eq(b, 2)

    (c, d) = list_return()
    assert.eq(c, 1)
    assert.eq(d, 2)

    [e, f] = tuple_return()
    assert.eq(e, 1)
    assert.eq(f, 2)

    [g, h] = list_return()
    assert.eq(g, 1)
    assert.eq(h, 2)    

test_explicit_unpack_return()

# Test that tuples can be assigned to a single variable.
def test_tuple_return():
    single_tuple = tuple_return()
    assert.eq(single_tuple, (1, 2))

test_tuple_return()

# Test multi-return with tuple as one of the values.
def test_tuple_multi_return():
    single_tuple, s = tuple_multi_value_return()
    assert.eq(single_tuple, (1, 2))
    assert.eq(s, "foo")

test_tuple_multi_return()

def test_no_assignment():
    bare_return()
    no_return()
    one_value()
    two_values()
    three_values()

test_no_assignment()

def test_explicit_unpack_literal():
    (a, b, c,) = (1, 2, 3) # trailing comma ok
    assert.eq(a, 1)
    assert.eq(b, 2)
    assert.eq(c, 3)

    (d, e, f,) = [1, 2, 3] # trailing comma ok
    assert.eq(d, 1)
    assert.eq(e, 2)
    assert.eq(f, 3)

    [g, h, i,] = (1, 2, 3) # trailing comma ok
    assert.eq(g, 1)
    assert.eq(h, 2)
    assert.eq(i, 3)

    [j, k, l,] = [1, 2, 3] # trailing comma ok
    assert.eq(j, 1)
    assert.eq(k, 2)
    assert.eq(l, 3)

test_explicit_unpack_literal()

# Test that implicit unpacking of explicit tuples works
def test_implicit_tuple_unpack_literal():
    a, b = (1, 2)
    assert.eq(a, 1)
    assert.eq(b, 2)

test_implicit_tuple_unpack_literal()

# Test that implicit unpacking of lists works
def test_implicit_list_unpack_literal():
    a, b = [1, 2]
    assert.eq(a, 1)
    assert.eq(b, 2)

test_implicit_list_unpack_literal()

# Test that implicit unpacking of bare tuples works
def test_implicit_bare_tuple_unpack():
    a, b = 1, 2
    assert.eq(a, 1)
    assert.eq(b, 2)

test_implicit_bare_tuple_unpack()

# Note: "x = 1, 2" (implicit tuple creation) is a compile-time error in strict mode
# and cannot be tested with assert.fails (which tests runtime errors).
# It's tested in a separate test file.

# Test that explicit tuple creation works
def test_explicit_tuple_creation():
    x = (1, 2)
    assert.eq(x, (1, 2))

test_explicit_tuple_creation()

# Test that list assignment works
def test_list_assignment():
    x = [1, 2]
    assert.eq(x, [1, 2])

test_list_assignment()
