# Tests for defer/errdefer behaviour on the error and failure paths.
#
# An "error" is a '!' function returning an error value: it is recoverable
# (catchable with try/catch/recover).
# A "failure" is an unrecoverable abort that unwinds the frame and cannot be
# intercepted by catch. A failure is either explicit — a call to fail() — or
# implicit — a runtime fault such as 1//0 or an arity mismatch.
#
# Many tests use assert.fails(f, pattern), which runs f(), swallows the failure
# at the harness level (so Starlark execution resumes), and matches the message
# against pattern. Because the swallowing happens in Go, closure state mutated by
# the deferred calls that ran during the unwind is observable afterwards. Patterns
# beginning with (?s) let '.' span the newlines that separate chained messages.

load("assert.star", "assert")

errors = error_tags("E")

def bad():  # takes no arguments; calling it with args is an arity error
    pass

# ---------------------------------------------------------------------------
# Both errors and failures run every stacked errdefer and defer (errdefers
# first, each group in LIFO order).
# ---------------------------------------------------------------------------

def test_error_runs_all_stacked_cleanup():
    log = []
    def f()!:
        defer log.append("d1")
        errdefer log.append("ed1")
        defer log.append("d2")
        errdefer log.append("ed2")
        log.append("body")
        return errors.E  # error trigger
    result = f() catch "caught"
    assert.eq(result, "caught")
    assert.eq(log, ["body", "ed2", "ed1", "d2", "d1"])

test_error_runs_all_stacked_cleanup()

def test_failure_runs_all_stacked_cleanup():
    log = []
    def f()!:
        defer log.append("d1")
        errdefer log.append("ed1")
        defer log.append("d2")
        errdefer log.append("ed2")
        log.append("body")
        fail("boom")  # failure trigger (called dynamically, so no try/catch site)
    assert.fails(f, "boom")
    assert.eq(log, ["body", "ed2", "ed1", "d2", "d1"])

test_failure_runs_all_stacked_cleanup()

# ---------------------------------------------------------------------------
# A deferred call that itself fails.
# ---------------------------------------------------------------------------

# On a clean return, a failing deferred call turns success into a failure.
def test_defer_failure_turns_success_into_failure():
    def f():
        defer bad(1, 2)  # arity failure, raised at unwind
        return "ok"
    assert.fails(f, "function bad accepts no arguments .2 given.")

test_defer_failure_turns_success_into_failure()

# A deferred call may itself call fail().
def test_defer_calls_fail():
    def f():
        defer fail("boom in defer")
        return "ok"
    assert.fails(f, "boom in defer")

test_defer_calls_fail()

# ---------------------------------------------------------------------------
# Failures raised by deferred calls are chained, and the remaining deferred
# calls still run.
# ---------------------------------------------------------------------------

# Two failing defers on a clean return: both failures chained, LIFO order.
def test_cleanup_failures_chained_on_clean_return():
    def boom(tag):
        fail("boom-" + tag)
    def f():
        defer boom("A")
        defer boom("B")
        return "ok"
    assert.fails(f, "(?s)boom-B.*boom-A")  # B (last deferred) runs first

test_cleanup_failures_chained_on_clean_return()

# A failing deferred call is chained onto the failure that triggered the unwind
# in the first place.
def test_cleanup_failure_chained_to_triggering_failure():
    def boom(tag):
        fail("boom-" + tag)
    def f():
        defer boom("A")
        x = 1 // 0  # triggering (implicit) failure
        return "unreachable"
    assert.fails(f, "(?s)division by zero.*boom-A")

test_cleanup_failure_chained_to_triggering_failure()

# A failing deferred call does not stop the remaining deferred calls.
def test_remaining_cleanup_runs_after_one_fails():
    log = []
    def boom(tag):
        log.append(tag)
        fail("boom-" + tag)
    def f():
        defer log.append("d_bottom")
        defer boom("mid")
        defer log.append("d_top")
        return "ok"
    assert.fails(f, "boom-mid")
    assert.eq(log, ["d_top", "mid", "d_bottom"])  # all three ran

test_remaining_cleanup_runs_after_one_fails()

# ---------------------------------------------------------------------------
# A failing errdefer on the error path replaces the (catchable) error value with
# a failure and BYPASSES catch: g() would evaluate to "caught" if the errdefer
# had not failed, but instead the arity failure propagates, uncaught.
# ---------------------------------------------------------------------------

def test_failing_errdefer_bypasses_catch():
    def f()!:
        errdefer bad(1, 2)
        return errors.E
    def g():
        return f() catch "caught"
    assert.fails(g, "function bad accepts no arguments .2 given.")

test_failing_errdefer_bypasses_catch()

# ---------------------------------------------------------------------------
# Deferring an error-returning function is allowed; an error it returns is
# ignored (like Go ignores the return values of a deferred call). A failure it
# raises is still chained.
# ---------------------------------------------------------------------------

# defer a '!' function whose returned error is discarded: the outer call succeeds.
def test_defer_bang_error_ignored():
    log = []
    def cleanup()!:
        log.append("cleanup")
        return errors.E
    def f():
        defer cleanup()
        log.append("body")
        return "ok"
    f()
    assert.eq(log, ["body", "cleanup"])

test_defer_bang_error_ignored()

# errdefer a '!' function on the error path: its returned error is ignored, and
# the original (catchable) error still propagates to the catch.
def test_errdefer_bang_error_ignored():
    log = []
    def cleanup()!:
        log.append("cleanup")
        return errors.E
    def f()!:
        errdefer cleanup()
        log.append("body")
        return errors.E
    result = f() catch "caught"
    assert.eq(result, "caught")
    assert.eq(log, ["body", "cleanup"])

test_errdefer_bang_error_ignored()

# A failure raised inside a deferred error-returning function is NOT ignored;
# it is chained like any other deferred failure.
def test_defer_bang_failure_chained():
    def cleanup()!:
        x = 1 // 0
        return "unreachable"
    def f():
        defer cleanup()
        return "ok"
    assert.fails(f, "division by zero")

test_defer_bang_failure_chained()
