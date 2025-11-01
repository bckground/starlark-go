load("assert.star", "assert")

# Test basic defer execution.
def test_basic_defer():
    result = []
    def append_value(x):
        result.append(x)

    append_value(1)
    defer append_value(2)
    append_value(3)
    # Execution order: 1, 3, 2 (deferred runs last)
    return result

final_result = test_basic_defer()
assert.eq(final_result, [1, 3, 2])

# Test multiple defers (LIFO order).
def test_multiple_defers():
    result = []
    def append(x):
        result.append(x)

    defer append("first")
    defer append("second")
    defer append("third")
    append("body")
    # Defers execute in reverse order: third, second, first.
    return result

result2 = test_multiple_defers()
assert.eq(result2, ["body", "third", "second", "first"])

# Test defer with arguments evaluated immediately.
def test_argument_evaluation():
    result = []
    def append(x):
        result.append(x)

    i = 0
    defer append(i)  # Captures i=0 immediately
    i = 1
    defer append(i)  # Captures i=1 immediately
    i = 2
    append(i)        # Appends i=2 immediately
    # Final i=2, but defers captured 0 and 1
    return result

result3 = test_argument_evaluation()
assert.eq(result3, [2, 1, 0])

# Test defer on normal return.
def test_defer_on_return():
    result = []
    def cleanup():
        result.append("cleaned")

    def with_defer():
        defer cleanup()
        result.append("body")
        return "done"

    ret = with_defer()
    assert.eq(ret, "done")
    assert.eq(result, ["body", "cleaned"])

test_defer_on_return()

# Test defer on early return.
def test_defer_on_early_return():
    result = []
    def cleanup():
        result.append("cleaned")

    def with_early_return(flag):
        defer cleanup()
        if flag:
            result.append("early")
            return "early"
        result.append("normal")
        return "normal"

    # Test early return path.
    result = []
    ret = with_early_return(True)
    assert.eq(ret, "early")
    assert.eq(result, ["early", "cleaned"])

    # Test normal return path.
    result = []
    ret = with_early_return(False)
    assert.eq(ret, "normal")
    assert.eq(result, ["normal", "cleaned"])

test_defer_on_early_return()

# Test defer in nested functions.
def test_defer_in_nested_functions():
    result = []
    def append(x):
        result.append(x)

    def outer():
        defer append("outer")
        def inner():
            defer append("inner")
            append("inner-body")
        inner()
        append("outer-body")

    outer()
    # Order: inner-body, inner, outer-body, outer.
    assert.eq(result, ["inner-body", "inner", "outer-body", "outer"])

test_defer_in_nested_functions()

# Test defer with closures.
def test_defer_with_closures():
    result = []

    def make_appender(x):
        defer (lambda: result.append(x))()
        result.append(x * 2)

    make_appender(5)
    # Order: 10 (body), then 5 (deferred lambda).
    assert.eq(result, [10, 5])

test_defer_with_closures()

# Test defer with function arguments.
def test_defer_with_args():
    result = []

    def log(name, value):
        result.append((name, value))

    def process():
        defer log("status", "done")
        result.append(("started", True))

    process()
    assert.eq(result, [("started", True), ("status", "done")])

test_defer_with_args()

# Test defer with keyword arguments.
def test_defer_with_kwargs():
    result = []

    def log(name, value):
        result.append((name, value))

    def process():
        defer log(name="operation", value="complete")
        result.append("working")

    process()
    assert.eq(result, ["working", ("operation", "complete")])

test_defer_with_kwargs()

# Test defer with mixed positional and keyword arguments.
def test_defer_with_mixed_args():
    result = []

    def log(a, b, c=None):
        result.append((a, b, c))

    def process():
        defer log(1, 2, c=3)
        result.append("done")

    process()
    assert.eq(result, ["done", (1, 2, 3)])

test_defer_with_mixed_args()

# Test error: defer outside function.
# This should be caught by the resolver, but we can't test that easily here.

# Test error: defer non-call expression.
# This should be caught by the parser.
