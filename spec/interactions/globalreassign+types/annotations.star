# spec: spec.md#annotation-semantics

# An assignment annotation belongs to the statement, not the
# variable: with globalreassign, a later unannotated reassignment of
# an annotated global is not checked.
x: int = 1
x = "now a string"
assert.eq(x, "now a string")

# Each annotated reassignment is checked independently, and may use a
# different type.
x: str = "checked as str"
x: int = 42
assert.eq(x, 42)

# A reassigned annotated statement is re-checked each time it
# executes (here: per loop iteration via a function).
def assign(v):
    y: int = v
    return y

assert.eq(assign(1), 1)
assert.fails(lambda: assign("bad"),
             "does not match the type annotation `int` for assignment `y`")

# Annotations are evaluated when their def executes: reassigning a
# type alias afterwards does not change functions already defined.
Alias = int

def takes_alias(v: Alias):
    return v

Alias = str

assert.eq(takes_alias(1), 1)
assert.fails(lambda: takes_alias("s"),
             "does not match the type annotation `int` for argument `v`")

# A def executed after the reassignment captures the new alias.
def takes_new_alias(v: Alias):
    return v

assert.eq(takes_new_alias("s"), "s")
assert.fails(lambda: takes_new_alias(1),
             "does not match the type annotation `str` for argument `v`")
