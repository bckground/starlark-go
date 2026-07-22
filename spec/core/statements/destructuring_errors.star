# spec: spec.md#assignments

# A sequence assignment requires a sequence operand of matching
# length.
(a, b, c) = 1 ### "got int in sequence assignment"
---
(a, b) = (1,) ### "too few values to unpack"
---
(a, b) = (1, 2, 3) ### "too many values to unpack"
---
[a, b, c] = 1 ### "got int in sequence assignment"
---
[a, b] = [] ### "too few values to unpack"
---
[a, b] = [1, 2, 3] ### "too many values to unpack"
---
# The empty targets accept only empty sequences.
() = 1 ### "got int in sequence assignment"
---
() = (1,) ### "too many values to unpack"
---
[] = [1, 2] ### "too many values to unpack"
---
# The empty cases that succeed: any empty sequence matches any empty
# target, regardless of bracket style.
() = ()
[] = []
[] = ()
() = []
