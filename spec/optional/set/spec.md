# Unit: set

This unit adds the *set* data type of the core specification
(spec.md#sets), which the core specification makes optional. An
implementation that adopts this unit predeclares the `set` built-in
function.

## The set type

A set is a mutable collection of distinct hashable values. The
[type](spec.md#type) of a set is `"set"`. A set is iterable and
preserves insertion order. The built-in `len` returns the number of
elements; `x in s` tests membership; sets are considered True if
non-empty.

Attempting to insert an unhashable value is an error. Sets themselves
are not hashable. A frozen set cannot be mutated.

## The set function

`set(x)` returns a new set containing the elements of the iterable
`x`, deduplicated, in first-insertion order. With no argument, it
returns a new empty set.

## Operators

Given sets `x` and `y`:

- `x | y` — union
- `x & y` — intersection
- `x - y` — difference
- `x ^ y` — symmetric difference

Each yields a new set and leaves the operands unmodified. Two sets
are equal if they contain the same elements, regardless of order.
Sets are not ordered by `<`.

## Methods

A set has these methods: `add`, `clear`, `difference`, `discard`,
`intersection`, `issubset`, `issuperset`, `pop`, `remove`,
`symmetric_difference`, `union`, `update`.

`remove` of an absent element is an error; `discard` is not. `pop`
removes and returns the first element; on an empty set it is an
error.
