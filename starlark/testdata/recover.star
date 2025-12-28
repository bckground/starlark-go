load("assert.star", "assert")

errors = error("Err", "ErrTimeout")

# Test recover assigns value and resumes normal execution.
def test_basic_recover():
    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        recover "recovered"

    assert.eq(result, "recovered")

test_basic_recover()

# Test recover with computed value.
def test_recover_computed_value():
    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        recover "recovered_" + str(e)

    assert.eq(result, "recovered_Err")

test_recover_computed_value()

# Test recover stops error propagation.
def test_recover_stops_propagation():
    def inner()!:
        return errors.Err

    def outer()!:
        value = inner() catch e:
            recover "recovered"
        return value + "_outer"

    result = outer() catch "outer_caught"
    # Recover in inner catch stops error, so outer succeeds.
    assert.eq(result, "recovered_outer")

test_recover_stops_propagation()

# Test recover vs return in catch block.
def test_recover_vs_return():
    def may_fail()!:
        return errors.Err

    def test_recover()!:
        value = may_fail() catch e:
            recover "recovered"
        return value + "_after"

    def test_return()!:
        value = may_fail() catch e:
            return errors.ErrTimeout
        return "unreachable"

    result1 = test_recover() catch "caught"
    assert.eq(result1, "recovered_after")

    result2 = test_return() catch "caught"
    assert.eq(result2, "caught")

test_recover_vs_return()

# Test conditional recover.
def test_conditional_recover():
    def may_fail(which)!:
        if which == 1:
            return errors.Err
        else:
            return errors.ErrTimeout

    result1 = may_fail(1) catch e:
        if e == errors.Err:
            recover "recovered_Err"
        else:
            recover "recovered_other"

    result2 = may_fail(2) catch e:
        if e == errors.Err:
            recover "recovered_Err"
        else:
            recover "recovered_Timeout"

    assert.eq(result1, "recovered_Err")
    assert.eq(result2, "recovered_Timeout")

test_conditional_recover()

# Test recover with multiple statements.
def test_recover_with_statements():
    log = []

    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        log.append("caught: " + str(e))
        log.append("processing")
        recover "final"

    assert.eq(result, "final")
    assert.eq(log, ["caught: Err", "processing"])

test_recover_with_statements()

# Test recover only valid in catch block.
# This should fail at compile time.
# def test_recover_outside_catch():
#     def test()!:
#         recover "value"  # Compile error: recover outside catch block.
#         return "unreachable"

# Test recover in nested catch.
def test_recover_nested_catch():
    def inner_fail()!:
        return errors.Err

    def outer_fail()!:
        return errors.ErrTimeout

    result = outer_fail() catch e1:
        inner_result = inner_fail() catch e2:
            recover "inner_recovered"
        recover inner_result + "_outer_recovered"

    assert.eq(result, "inner_recovered_outer_recovered")

test_recover_nested_catch()

# Test recover with error variable shadowing.
def test_recover_shadow():
    e = "outer_e"

    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        recover "recovered: " + str(e)

    # Outer 'e' unchanged.
    assert.eq(e, "outer_e")
    assert.eq(result, "recovered: Err")

test_recover_shadow()

# Test multiple recovers in same catch (first one wins - rest is dead code).
def test_multiple_recovers():
    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        recover "first"
        recover "second"  # Dead code - never executed
        recover "third"   # Dead code - never executed

    # First recover executes and jumps away, rest is unreachable.
    assert.eq(result, "first")

test_multiple_recovers()
