# spec: spec.md#while-loops

# A while loop executes its body while the condition is true.
def countdown(n):
    out = []
    while n > 0:
        out.append(n)
        n -= 1
    return out

assert.eq(countdown(3), [3, 2, 1])
assert.eq(countdown(0), [])

# The condition is a truth test.
def drain(stack):
    out = []
    while stack:
        out.append(stack.pop())
    return out

assert.eq(drain([1, 2]), [2, 1])

# break and continue work as in for loops.
def first_factor(n):
    k = 2
    while True:
        if n % k == 0:
            break
        k += 1
    return k

assert.eq(first_factor(35), 5)

def odd_sum(limit):
    total = 0
    i = 0
    while i < limit:
        i += 1
        if i % 2 == 0:
            continue
        total += i
    return total

assert.eq(odd_sum(5), 9)
