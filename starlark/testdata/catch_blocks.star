load("assert.star", "assert")

errors = error_tags("ErrA", "ErrB", "ErrTimeout")
network_errors = error_tags("Timeout", "Disconnected")

# Test catch block with error binding.
def test_catch_block_binding():
    def may_fail()!:
        return errors.ErrA

    result = may_fail() catch e:
        recover "caught: " + str(e)

    assert.eq(result, "caught: ErrA")

test_catch_block_binding()

# Test catch block can inspect error.
def test_catch_block_inspect_error_tags():
    def may_fail()!:
        return network_errors.Timeout

    result = may_fail() catch e:
        if e.tag == network_errors.Timeout:
            recover "timeout_handled"
        else:
            recover "other_error"

    assert.eq(result, "timeout_handled")

test_catch_block_inspect_error_tags()

# Test catch block with multiple statements.
def test_catch_block_multiple_statements():
    def may_fail()!:
        return errors.ErrA

    log = []
    result = may_fail() catch e:
        log.append("caught")
        log.append(str(e))
        recover "handled"

    assert.eq(result, "handled")
    assert.eq(log, ["caught", "ErrA"])

test_catch_block_multiple_statements()

# Test catch block can return from outer function.
def test_catch_block_return():
    def outer()!:
        def inner()!:
            return errors.ErrA

        value = inner() catch e:
            return errors.ErrB  # Return from outer.

        return "should_not_reach"

    result = outer() catch "final_catch"
    assert.eq(result, "final_catch")

test_catch_block_return()

# Test catch block with conditional logic.
def test_catch_block_conditional():
    def may_fail(which)!:
        if which == 1:
            return errors.ErrA
        else:
            return errors.ErrB

    result1 = may_fail(1) catch e:
        if e.tag == errors.ErrA:
            recover "handled_A"
        else:
            recover "handled_other"

    result2 = may_fail(2) catch e:
        if e.tag == errors.ErrA:
            recover "handled_A"
        else:
            recover "handled_B"

    assert.eq(result1, "handled_A")
    assert.eq(result2, "handled_B")

test_catch_block_conditional()

# Test nested catch blocks.
def test_nested_catch_blocks():
    def inner_fail()!:
        return errors.ErrA

    def outer_fail()!:
        return errors.ErrB

    result = outer_fail() catch e1:
        value = inner_fail() catch e2:
            recover "inner_caught: " + str(e2)
        recover value + " outer: " + str(e1)

    assert.eq(result, "inner_caught: ErrA outer: ErrB")

test_nested_catch_blocks()

# Test catch block has access to outer scope.
def test_catch_block_outer_scope():
    prefix = "prefix:"

    def may_fail()!:
        return errors.ErrA

    result = may_fail() catch e:
        recover prefix + str(e)

    assert.eq(result, "prefix:ErrA")

test_catch_block_outer_scope()

---
# Test: catch block without recover or return is a runtime error.
errors = error_tags("Err")

def may_fail()!:
    return errors.Err

result = may_fail() catch e:  ### "catch block must end with recover or return"
    x = 1
