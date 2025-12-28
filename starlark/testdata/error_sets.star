load("assert.star", "assert")

# Test error() builtin creates error set with attributes.
def test_error_set_creation():
    errs = error("ErrA", "ErrB", "ErrC")
    assert.eq(str(errs.ErrA), "ErrA")
    assert.eq(str(errs.ErrB), "ErrB")
    assert.eq(str(errs.ErrC), "ErrC")

test_error_set_creation()

# Test error values are equal to themselves.
def test_error_equality():
    errs = error("Err1", "Err2")
    assert.eq(errs.Err1, errs.Err1)
    assert.eq(errs.Err2, errs.Err2)

test_error_equality()

# Test error values from same set are different.
def test_error_inequality():
    errs = error("Err1", "Err2")
    assert.ne(errs.Err1, errs.Err2)

test_error_inequality()

# Test error values from different sets are different.
def test_error_inequality_different_sets():
    errs1 = error("Err")
    errs2 = error("Err")
    # Even with same name, different sets produce different errors.
    assert.ne(errs1.Err, errs2.Err)

test_error_inequality_different_sets()

# Test error set has correct type.
def test_error_set_type():
    errs = error("Err")
    assert.eq(type(errs), "error_set")

test_error_set_type()

# Test error value has correct type.
def test_error_value_type():
    errs = error("Err")
    assert.eq(type(errs.Err), "error")

test_error_value_type()

# Test error() with no arguments.
def test_error_empty():
    errs = error()
    assert.eq(type(errs), "error_set")

test_error_empty()

# Test error() with single argument.
def test_error_single():
    errs = error("SingleErr")
    assert.eq(str(errs.SingleErr), "SingleErr")

test_error_single()

# Test accessing non-existent error attribute.
def test_error_nonexistent_attribute():
    errs = error("Err1")
    # Accessing non-existent attribute should return None or fail.
    assert.fails(lambda: errs.Err2, "no such attribute")

test_error_nonexistent_attribute()
