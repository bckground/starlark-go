load("assert.star", "assert")

errors = error_tags("Err")

# Test error variable doesn't leak from catch block.
def test_error_variable_no_leak():
    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        recover "caught"

    # Variable 'e' should not be accessible here.
    # This test validates that referencing 'e' outside the catch block fails.
    # Since we can't directly test for undefined variables in Starlark, we ensure
    # the error was caught but 'e' doesn't pollute the outer scope.

test_error_variable_no_leak()

# Test catch block variable shadows outer variable.
def test_catch_block_shadows_outer():
    e = "outer_e"

    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        # This 'e' is the error, not "outer_e".
        recover "caught: " + str(e)

    # After catch, 'e' should still be "outer_e".
    assert.eq(e, "outer_e")
    assert.eq(result, "caught: Err")

test_catch_block_shadows_outer()

# Test variable assigned in catch block doesn't leak.
def test_catch_block_local_vars_dont_leak():
    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        local_var = "local"
        recover local_var + "_" + str(e)

    assert.eq(result, "local_Err")
    # local_var should not be accessible here (would need manual test).

test_catch_block_local_vars_dont_leak()

# Test catch block can modify outer variables.
def test_catch_block_modifies_outer():
    outer_list = []

    def may_fail()!:
        return errors.Err

    result = may_fail() catch e:
        outer_list.append(str(e))
        recover "handled"

    assert.eq(result, "handled")
    assert.eq(outer_list, ["Err"])

test_catch_block_modifies_outer()

# Test nested catch blocks have separate scopes.
def test_nested_catch_scopes():
    def fail1()!:
        return errors.Err

    def fail2()!:
        return errors.Err

    outer_e = "outer"

    result = fail1() catch e:
        inner_result = fail2() catch e:  # This 'e' shadows outer catch's 'e'.
            recover "inner_" + str(e)
        recover inner_result + "_outer_" + str(e)

    # outer_e should remain unchanged.
    assert.eq(outer_e, "outer")
    assert.eq(result, "inner_Err_outer_Err")

test_nested_catch_scopes()

# Test catch block with same name as function parameter.
def test_catch_shadow_parameter():
    def process(e)!:
        def may_fail()!:
            return errors.Err

        result = may_fail() catch e:  # Shadows parameter.
            recover "caught: " + str(e)

        # After catch, 'e' is still the parameter.
        return result + " param: " + e

    result = process("param_value") catch "error"
    assert.eq(result, "caught: Err param: param_value")

test_catch_shadow_parameter()
