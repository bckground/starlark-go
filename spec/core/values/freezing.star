# spec: spec.md#freezing-a-value

# (freeze is provided by the test harness; in production, values are
# frozen when their module finishes executing.)

# Freezing is recursive: it applies to all values reachable from the
# frozen value.
outer_list = [[1], {"k": [2]}]
frozen = freeze(outer_list)

# freeze returns its operand.
assert.eq(frozen, outer_list)

# Mutation of a frozen value, at any depth, is an error.
assert.fails(lambda: outer_list.append(3), "cannot append to frozen list")
assert.fails(lambda: outer_list[0].append(3), "cannot append to frozen list")

def mutate_nested_dict():
    outer_list[1]["k2"] = 1

assert.fails(mutate_nested_dict, "cannot insert into frozen hash table")

def mutate_deep_list():
    outer_list[1]["k"].clear()

assert.fails(mutate_deep_list, "cannot clear frozen list")

# Reading a frozen value is unaffected.
assert.eq(outer_list[0][0], 1)
assert.eq(len(outer_list), 2)

# Freezing an immutable value is a no-op.
assert.eq(freeze("s"), "s")
assert.eq(freeze(3), 3)
