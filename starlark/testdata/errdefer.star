load("assert.star", "assert")

errors = error_tags("Err")

# Test errdefer runs only on error.
def test_errdefer_runs_on_error_tags():
    log = []

    def cleanup():
        log.append("cleanup")

    def may_fail(should_fail)!:
        errdefer cleanup()
        log.append("body")
        if should_fail:
            return errors.Err
        return "success"

    # Test error case - errdefer should run.
    log = []
    result = may_fail(True) catch "caught"
    assert.eq(result, "caught")
    assert.eq(log, ["body", "cleanup"])

test_errdefer_runs_on_error_tags()

# Test errdefer doesn't run on success.
def test_errdefer_not_on_success():
    log = []

    def cleanup():
        log.append("cleanup")

    def may_fail(should_fail)!:
        errdefer cleanup()
        log.append("body")
        if should_fail:
            return errors.Err
        return "success"

    # Test success case - errdefer should NOT run.
    log = []
    result = may_fail(False) catch "caught"
    assert.eq(result, "success")
    assert.eq(log, ["body"])  # No cleanup.

test_errdefer_not_on_success()

# Test multiple errdefers execute in LIFO order.
def test_multiple_errdefers():
    log = []

    def append(x):
        log.append(x)

    def test()!:
        errdefer append("first")
        errdefer append("second")
        errdefer append("third")
        log.append("body")
        return errors.Err

    result = test() catch "caught"
    assert.eq(log, ["body", "third", "second", "first"])

test_multiple_errdefers()

# Test errdefer and defer together.
def test_errdefer_and_defer():
    log = []

    def append(x):
        log.append(x)

    def test_error_tags()!:
        defer append("defer1")
        errdefer append("errdefer1")
        defer append("defer2")
        errdefer append("errdefer2")
        log.append("body")
        return errors.Err

    # On error: errdefers run, then defers.
    log = []
    result = test_error_tags() catch "caught"
    # Errdefers run first (LIFO), then regular defers (LIFO).
    assert.eq(log, ["body", "errdefer2", "errdefer1", "defer2", "defer1"])

test_errdefer_and_defer()

# Test defer runs on success, errdefer doesn't.
def test_defer_vs_errdefer_on_success():
    log = []

    def append(x):
        log.append(x)

    def test_success()!:
        defer append("defer")
        errdefer append("errdefer")
        log.append("body")
        return "success"

    log = []
    result = test_success() catch "caught"
    # Only defer runs on success.
    assert.eq(log, ["body", "defer"])

test_defer_vs_errdefer_on_success()

# Test errdefer with arguments evaluated immediately.
def test_errdefer_argument_evaluation():
    log = []

    def append(x):
        log.append(x)

    def test()!:
        i = 1
        errdefer append(i)  # Captures i=1 immediately.
        i = 2
        errdefer append(i)  # Captures i=2 immediately.
        i = 3
        log.append(i)       # Appends i=3.
        return errors.Err

    result = test() catch "caught"
    # Errdefers capture 1 and 2, run in LIFO order.
    assert.eq(log, [3, 2, 1])

test_errdefer_argument_evaluation()

# Test errdefer only in error-returning function.
# This should fail at compile time, but we'll document it in comments.
# def test_errdefer_requires_error_function():
#     def normal_func():
#         errdefer print("cleanup")  # Compile error: errdefer in non-! function.
#         return "value"

# Test errdefer in nested functions.
def test_errdefer_nested_functions():
    log = []

    def append(x):
        log.append(x)

    def outer()!:
        errdefer append("outer_errdefer")
        log.append("outer_body")

        def inner()!:
            errdefer append("inner_errdefer")
            log.append("inner_body")
            return errors.Err

        result = inner() catch e:
            log.append("inner_caught")
            return errors.Err  # Still errors in outer.

        return "unreachable"

    log = []
    result = outer() catch "outer_caught"
    # Inner errdefer runs when inner errors, outer errdefer runs when outer errors.
    assert.eq(log, ["outer_body", "inner_body", "inner_errdefer", "inner_caught", "outer_errdefer"])

test_errdefer_nested_functions()

# Test errdefer runs when error propagates via try.
def test_errdefer_with_try_propagation():
    log = []

    def append(x):
        log.append(x)

    def inner()!:
        errdefer append("inner_errdefer")
        log.append("inner_body")
        return errors.Err

    def outer()!:
        errdefer append("outer_errdefer")
        result = try inner()  # error propagates via try
        return result

    log = []
    result = outer() catch "caught"
    assert.eq(result, "caught")
    # Both inner and outer errdefers must run.
    assert.eq(log, ["inner_body", "inner_errdefer", "outer_errdefer"])

test_errdefer_with_try_propagation()
