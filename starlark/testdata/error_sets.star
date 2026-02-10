load("assert.star", "assert")

# Test error_tags() builtin creates error set with attributes.
def test_error_set_creation():
    errs = error_tags("ErrA", "ErrB", "ErrC")
    assert.eq(str(errs.ErrA), "ErrA")
    assert.eq(str(errs.ErrB), "ErrB")
    assert.eq(str(errs.ErrC), "ErrC")

test_error_set_creation()

# Test error values are equal to themselves.
def test_error_equality():
    errs = error_tags("Err1", "Err2")
    assert.eq(errs.Err1, errs.Err1)
    assert.eq(errs.Err2, errs.Err2)

test_error_equality()

# Test error values from same set are different.
def test_error_inequality():
    errs = error_tags("Err1", "Err2")
    assert.ne(errs.Err1, errs.Err2)

test_error_inequality()

# Test error values from different sets are different.
def test_error_inequality_different_sets():
    errs1 = error_tags("Err")
    errs2 = error_tags("Err")
    # Even with same name, different sets produce different errors.
    assert.ne(errs1.Err, errs2.Err)

test_error_inequality_different_sets()

# Test error set has correct type.
def test_error_set_type():
    errs = error_tags("Err")
    assert.eq(type(errs), "error_set")

test_error_set_type()

# Test error value has correct type.
def test_error_value_type():
    errs = error_tags("Err")
    assert.eq(type(errs.Err), "error_tag")

test_error_value_type()

# Test error_tags() with no arguments.
def test_error_empty():
    errs = error_tags()
    assert.eq(type(errs), "error_set")

test_error_empty()

# Test error_tags() with single argument.
def test_error_single():
    errs = error_tags("SingleErr")
    assert.eq(str(errs.SingleErr), "SingleErr")

test_error_single()

# Test accessing non-existent error attribute.
def test_error_nonexistent_attribute():
    errs = error_tags("Err1")
    # Accessing non-existent attribute should fail with a helpful message.
    assert.fails(lambda: errs.Err2, "has no attribute")

test_error_nonexistent_attribute()

# Test merging error sets with | and +.
def test_error_set_merge():
    errs1 = error_tags("ErrA", "ErrB")
    errs2 = error_tags("ErrC", "ErrD")
    # Test | operator
    merged_pipe = errs1 | errs2
    assert.eq(type(merged_pipe), "error_set")
    assert.eq(str(merged_pipe.ErrA), "ErrA")
    assert.eq(str(merged_pipe.ErrB), "ErrB")
    assert.eq(str(merged_pipe.ErrC), "ErrC")
    assert.eq(str(merged_pipe.ErrD), "ErrD")
    # Test + operator
    merged_plus = errs1 + errs2
    assert.eq(type(merged_plus), "error_set")
    assert.eq(str(merged_plus.ErrA), "ErrA")
    assert.eq(str(merged_plus.ErrB), "ErrB")
    assert.eq(str(merged_plus.ErrC), "ErrC")
    assert.eq(str(merged_plus.ErrD), "ErrD")

test_error_set_merge()

# Test merging error sets preserves identity from left operand.
def test_error_set_merge_preserves_left():
    errs1 = error_tags("Err")
    errs2 = error_tags("Err")  # Same name, different set
    # Test | operator
    merged_pipe = errs1 | errs2
    assert.eq(merged_pipe.Err, errs1.Err)
    assert.ne(merged_pipe.Err, errs2.Err)
    # Test + operator
    merged_plus = errs1 + errs2
    assert.eq(merged_plus.Err, errs1.Err)
    assert.ne(merged_plus.Err, errs2.Err)

test_error_set_merge_preserves_left()

# Test merging with empty error set.
def test_error_set_merge_empty():
    errs = error_tags("Err")
    empty = error_tags()
    # Test | operator
    assert.eq(str((errs | empty).Err), "Err")
    assert.eq(str((empty | errs).Err), "Err")
    # Test + operator
    assert.eq(str((errs + empty).Err), "Err")
    assert.eq(str((empty + errs).Err), "Err")

test_error_set_merge_empty()

# Test chained merging.
def test_error_set_merge_chain():
    errs1 = error_tags("A")
    errs2 = error_tags("B")
    errs3 = error_tags("C")
    # Test | operator
    merged_pipe = errs1 | errs2 | errs3
    assert.eq(str(merged_pipe.A), "A")
    assert.eq(str(merged_pipe.B), "B")
    assert.eq(str(merged_pipe.C), "C")
    # Test + operator
    merged_plus = errs1 + errs2 + errs3
    assert.eq(str(merged_plus.A), "A")
    assert.eq(str(merged_plus.B), "B")
    assert.eq(str(merged_plus.C), "C")

test_error_set_merge_chain()
