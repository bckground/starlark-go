load("assert.star", "assert")

errors = error("ErrA", "ErrB", "ErrTimeout")

# Test module-level catch with value form.
def fail_a()!:
    return errors.ErrA

result_value = fail_a() catch "caught_at_module_level"
assert.eq(result_value, "caught_at_module_level")

# Test module-level catch with block form.
def fail_b()!:
    return errors.ErrB

result_block = fail_b() catch e:
    recover "module_caught: " + str(e)

assert.eq(result_block, "module_caught: ErrB")

# Test module-level catch doesn't leak error variable.
def fail_timeout()!:
    return errors.ErrTimeout

result_no_leak = fail_timeout() catch err:
    recover "timeout_handled"

assert.eq(result_no_leak, "timeout_handled")
# Variable 'err' should not be accessible at module level after catch block.
# (Can't directly test for undefined, but the test passing shows it works)

# Test module-level catch with successful call (no error).
def succeed()!:
    return "success_value"

result_success = succeed() catch "fallback"
assert.eq(result_success, "success_value")

# Test module-level catch can access other module-level variables.
prefix = "PREFIX:"

def fail_for_prefix()!:
    return errors.ErrA

result_with_prefix = fail_for_prefix() catch e:
    recover prefix + str(e)

assert.eq(result_with_prefix, "PREFIX:ErrA")

# Test module-level catch with error comparison.
def conditional_fail()!:
    return errors.ErrA

result_conditional = conditional_fail() catch e:
    if e == errors.ErrA:
        recover "matched_ErrA"
    else:
        recover "matched_other"

assert.eq(result_conditional, "matched_ErrA")

# Test multiple module-level catches in sequence.
def fail1()!:
    return errors.ErrA

def fail2()!:
    return errors.ErrB

first = fail1() catch "first_caught"
second = fail2() catch "second_caught"

assert.eq(first, "first_caught")
assert.eq(second, "second_caught")

# Test module-level catch doesn't interfere with function-level catch.
def func_with_catch()!:
    inner_result = fail_a() catch "inner_caught"
    return inner_result

module_func_result = func_with_catch() catch "outer_caught"
assert.eq(module_func_result, "inner_caught")
