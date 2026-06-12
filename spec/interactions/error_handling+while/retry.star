# spec: spec.md#catch

errs = error_tags("Flaky")

# The classic retry loop: attempt until the call stops returning an
# error.
attempts = []

def flaky()!:
    attempts.append(True)
    if len(attempts) < 3:
        return errs.Flaky
    return "ok"

def retry():
    while True:
        result = flaky() catch None
        if result != None:
            return result

assert.eq(retry(), "ok")
assert.eq(len(attempts), 3)

# A bounded retry that re-raises after too many failures.
def stubborn()!:
    return errs.Flaky(message="still down")

def retry_up_to(n)!:
    remaining = n
    while remaining > 0:
        remaining -= 1
        result = stubborn() catch e:
            if remaining == 0:
                return e
            recover None
        if result != None:
            return result
    return errs.Flaky

final = retry_up_to(3) catch e:
    recover "gave up: " + e.message

assert.eq(final, "gave up: still down")

# catch in the loop condition (parenthesized, so the ':' that ends
# the condition is unambiguous): keep looping while the call errors.
polls = []

def poll()!:
    polls.append(True)
    if len(polls) < 4:
        return errs.Flaky
    return False  # success: nothing left to do

def drain():
    while (poll() catch True):
        pass
    return len(polls)

assert.eq(drain(), 4)
