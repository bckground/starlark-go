load("assert.star", "assert")

# Integration tests for complex error handling scenarios.

error_set = error("ErrA", "ErrB")
network_errors = error("Timeout", "Disconnected")

# Test complex example from the original plan.
def test_complex_scenario():
    log = []

    def some_func(should_fail)!:
        if should_fail:
            return error_set.ErrA
        return "success"

    def acquire_lock(name)!:
        log.append("acquired: " + name)
        return {"id": name, "release": lambda: log.append("released: " + name)}

    def on_error(lock_id):
        log.append("error_cleanup: " + lock_id)

    def example(fail_at)!:
        # Test catch with block and return.
        a = some_func(fail_at == 1) catch e:
            return e

        # Test try.
        b = try some_func(fail_at == 2)

        # Acquire lock with try.
        lock = try acquire_lock("foo")
        defer lock["release"]()

        # Errdefer.
        errdefer on_error(lock["id"])

        # Catch with value.
        c = some_func(fail_at == 3) catch 42

        # Catch with block and conditional recover.
        d = some_func(fail_at == 4) catch e:
            if e == network_errors.Timeout:
                recover "timeout_handled"
            return error_set.ErrB

        return "completed: " + a + " " + str(b) + " " + str(c) + " " + str(d)

    # Test success case.
    log = []
    result = example(0) catch "error"
    assert.eq(result, "completed: success success 42 success")
    # Defer should have run, but not errdefer.
    assert.true("released: foo" in log)
    assert.true("error_cleanup: foo" not in log)

    # Test error at first catch.
    log = []
    result = example(1) catch "error"
    assert.eq(result, "error")
    # Lock not acquired yet, no cleanup.
    assert.eq(log, [])

    # Test error at try.
    log = []
    result = example(2) catch "error"
    assert.eq(result, "error")
    # Lock not acquired yet, no cleanup.
    assert.eq(log, [])

    # Test error after lock acquired (errdefer should run).
    log = []
    result = example(3) catch "error"
    # fail_at == 3 means the catch with 42 has an error, but it catches it.
    # So this should complete successfully.
    assert.eq(result, "completed: success success 42 success")

    # Test error in last catch that returns error.
    log = []
    result = example(4) catch "error"
    assert.eq(result, "error")
    # Lock was acquired, errdefer and defer should run.
    assert.true("acquired: foo" in log)
    assert.true("error_cleanup: foo" in log)
    assert.true("released: foo" in log)

test_complex_scenario()

# Test error handling with resource management.
def test_resource_management():
    log = []

    def open_file(name, fail)!:
        if fail:
            return error_set.ErrA
        log.append("opened: " + name)
        return {"name": name, "close": lambda: log.append("closed: " + name)}

    def process()!:
        file1 = try open_file("file1", False)
        defer file1["close"]()

        file2 = try open_file("file2", True)  # This fails.
        defer file2["close"]()  # Never reached.

        return "unreachable"

    log = []
    result = process() catch "error"
    assert.eq(result, "error")
    # Only file1 should be closed (file2 failed to open).
    assert.eq(log, ["opened: file1", "closed: file1"])

test_resource_management()

# Test nested error handling with multiple levels.
def test_nested_levels():
    def level3(fail)!:
        if fail == 3:
            return error_set.ErrA
        return "level3"

    def level2(fail)!:
        result = level3(fail) catch e:
            if fail == 3:
                recover "level3_recovered"
            return e
        if fail == 2:
            return error_set.ErrB
        return result + "_level2"

    def level1(fail)!:
        result = try level2(fail)
        return result + "_level1"

    # No errors.
    result = level1(0) catch "caught"
    assert.eq(result, "level3_level2_level1")

    # Error at level3, recovered.
    result = level1(3) catch "caught"
    assert.eq(result, "level3_recovered_level2_level1")

    # Error at level2, propagates to level1.
    result = level1(2) catch "caught"
    assert.eq(result, "caught")

test_nested_levels()

# Test interleaved defer and errdefer with recovery.
def test_defer_errdefer_recovery():
    log = []

    def append(x):
        log.append(x)

    def process(should_fail)!:
        defer append("defer1")
        errdefer append("errdefer1")
        defer append("defer2")

        if should_fail:
            return error_set.ErrA

        errdefer append("errdefer2")
        defer append("defer3")
        return "success"

    # Success case: all defers run, no errdefers.
    log = []
    result = process(False) catch "error"
    assert.eq(result, "success")
    assert.eq(log, ["defer3", "defer2", "defer1"])

    # Error case: errdefers and defers run.
    log = []
    result = process(True) catch "error"
    assert.eq(result, "error")
    # errdefer1 runs, defer2 and defer1 run, but errdefer2 and defer3 not registered yet.
    assert.eq(log, ["errdefer1", "defer2", "defer1"])

test_defer_errdefer_recovery()

# Test error handling in loops.
def test_error_in_loop():
    results = []

    def process_item(item)!:
        if item == "fail":
            return error_set.ErrA
        return "ok_" + item

    def process_all(items)!:
        for item in items:
            result = process_item(item) catch "caught_" + item
            results.append(result)
        return "done"

    result = process_all(["a", "fail", "b"]) catch "error"
    assert.eq(result, "done")
    assert.eq(results, ["ok_a", "caught_fail", "ok_b"])

test_error_in_loop()
