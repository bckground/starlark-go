# Tests for single-argument load() with implicit module binding.
# option:loadmodulebinding

load("assert.star", "assert")

# load("json.star") is shorthand for load("json.star", "json")
load("json.star")
assert.eq(type(json), "module")
assert.eq(json.encode({"a": 1}), '{"a":1}')

# load("math.star") is shorthand for load("math.star", "math")
load("math.star")
assert.eq(type(math), "module")
assert.eq(math.floor(1.7), 1)

# load("time.star") is shorthand for load("time.star", "time")
load("time.star")
assert.eq(type(time), "module")

# load("noext") is shorthand for load("noext", "noext") — no extension
load("noext")
assert.eq(noext, "loaded_without_extension")
