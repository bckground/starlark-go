// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// This file defines Ty, the static type model: a canonical union of
// Basic alternatives, mirroring starlark-rust's Ty/TyBasic
// (starlark/src/typing/ty.rs).

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"go.starlark.net/syntax"
)

// A Ty is a static type: a union of basic (non-union) alternatives.
// The zero value is Never, the empty union.
//
// Ty values are canonical: alternatives are sorted, deduplicated, and
// adjacent container types are merged (list[a] | list[b] = list[a|b]).
// A union containing Any collapses to Any.
type Ty struct {
	alts []Basic
}

// A Basic is one non-union alternative of a Ty.
type Basic interface {
	String() string
	// basicSortKey is a total order used for canonicalization;
	// equal keys imply equal types.
	basicSortKey() string
}

// Constructors.

// Never returns the type with no values (the empty union).
func Never() Ty { return Ty{} }

// Any returns the dynamic type, which intersects every type.
func Any() Ty { return Ty{alts: []Basic{anyBasic{}}} }

// Prim returns a type matched by its runtime type name,
// e.g. "int", "string", "NoneType", "bool", "error".
func Prim(name string) Ty { return Ty{alts: []Basic{primBasic{name}}} }

// None returns the type of None.
func None() Ty { return Prim("NoneType") }

// List returns the type list[elem].
func List(elem Ty) Ty { return Ty{alts: []Basic{listBasic{elem}}} }

// Dict returns the type dict[key, val].
func Dict(key, val Ty) Ty { return Ty{alts: []Basic{dictBasic{key, val}}} }

// Set returns the type set[elem].
func Set(elem Ty) Ty { return Ty{alts: []Basic{setBasic{elem}}} }

// Tuple returns a fixed-arity tuple type.
func Tuple(elems ...Ty) Ty {
	if elems == nil {
		elems = []Ty{}
	}
	return Ty{alts: []Basic{tupleBasic{elems: elems}}}
}

// TupleOf returns the type of tuples of unknown arity whose elements
// all have type item (tuple[T, ...]).
func TupleOf(item Ty) Ty { return Ty{alts: []Basic{tupleBasic{of: &item}}} }

// AnyTuple returns the type of tuples of unknown arity and element type.
func AnyTuple() Ty { return TupleOf(Any()) }

// Iter returns the type of iterables yielding item.
func Iter(item Ty) Ty { return Ty{alts: []Basic{iterBasic{item}}} }

// Callable returns a callable type with the given parameters and result.
func Callable(params *ParamSpec, result Ty) Ty {
	return Ty{alts: []Basic{callableBasic{name: "", params: params, result: result}}}
}

// Function returns a named callable type, as bound by a def statement
// or a builtin with a known signature.
func Function(name string, params *ParamSpec, result Ty) Ty {
	return Ty{alts: []Basic{callableBasic{name: name, params: params, result: result}}}
}

// AnyCallable returns the type of callables about which nothing more
// is known: any arguments are accepted and the result is Any.
func AnyCallable() Ty { return Callable(AnyParams(), Any()) }

// AnyFunction returns a named callable accepting any arguments.
func AnyFunction(name string) Ty { return Function(name, AnyParams(), Any()) }

// TypeType returns the type of type values (the result of eval_type,
// list[int], etc.).
func TypeType() Ty { return Ty{alts: []Basic{typeBasic{}}} }

// Module returns an opaque module-like type with the given attributes.
// Attribute lookups not present in attrs yield Any.
func Module(name string, attrs map[string]Ty) Ty {
	return Ty{alts: []Basic{moduleBasic{name, attrs}}}
}

// Union returns the canonical union of the given types.
func Union(tys ...Ty) Ty {
	var alts []Basic
	for _, ty := range tys {
		alts = append(alts, ty.alts...)
	}
	return makeUnion(alts)
}

func makeUnion(alts []Basic) Ty {
	if len(alts) == 0 {
		return Never()
	}
	// Any absorbs everything.
	for _, a := range alts {
		if _, ok := a.(anyBasic); ok {
			return Any()
		}
	}
	// Sort and deduplicate. The order is case-insensitive so that
	// displays read naturally ("int | None", not "None | int").
	sorted := slices.Clone(alts)
	sort.SliceStable(sorted, func(i, j int) bool {
		ki, kj := sorted[i].basicSortKey(), sorted[j].basicSortKey()
		if li, lj := strings.ToLower(ki), strings.ToLower(kj); li != lj {
			return li < lj
		}
		return ki < kj
	})
	deduped := sorted[:0]
	var prev string
	for i, a := range sorted {
		key := a.basicSortKey()
		if i == 0 || key != prev {
			deduped = append(deduped, a)
		}
		prev = key
	}
	// Merge adjacent containers of the same shape:
	// list[a] | list[b] => list[a|b], and similarly dict/set/iter.
	merged := make([]Basic, 0, len(deduped))
	for _, a := range deduped {
		if len(merged) > 0 {
			if m, ok := mergePair(merged[len(merged)-1], a); ok {
				merged[len(merged)-1] = m
				continue
			}
		}
		merged = append(merged, a)
	}
	return Ty{alts: merged}
}

func mergePair(a, b Basic) (Basic, bool) {
	switch a := a.(type) {
	case listBasic:
		if b, ok := b.(listBasic); ok {
			return listBasic{Union(a.elem, b.elem)}, true
		}
	case setBasic:
		if b, ok := b.(setBasic); ok {
			return setBasic{Union(a.elem, b.elem)}, true
		}
	case dictBasic:
		if b, ok := b.(dictBasic); ok {
			return dictBasic{Union(a.key, b.key), Union(a.val, b.val)}, true
		}
	case iterBasic:
		if b, ok := b.(iterBasic); ok {
			return iterBasic{Union(a.item, b.item)}, true
		}
	}
	return nil, false
}

// Predicates and accessors.

// IsAny reports whether t is the dynamic type Any.
func (t Ty) IsAny() bool {
	if len(t.alts) != 1 {
		return false
	}
	_, ok := t.alts[0].(anyBasic)
	return ok
}

// IsNever reports whether t is the empty union.
func (t Ty) IsNever() bool { return len(t.alts) == 0 }

// Alternatives returns the basic alternatives of t.
func (t Ty) Alternatives() []Basic { return t.alts }

// Equal reports whether two types are identical.
// Canonicalization makes this a simple structural comparison.
func (t Ty) Equal(u Ty) bool {
	if len(t.alts) != len(u.alts) {
		return false
	}
	for i := range t.alts {
		if t.alts[i].basicSortKey() != u.alts[i].basicSortKey() {
			return false
		}
	}
	return true
}

func (t Ty) String() string {
	if len(t.alts) == 0 {
		return "typing.Never"
	}
	var b strings.Builder
	for i, a := range t.alts {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(a.String())
	}
	return b.String()
}

// displayPrim maps runtime type names to their display form.
func displayPrim(name string) string {
	switch name {
	case "NoneType":
		return "None"
	case "string":
		return "str"
	}
	return name
}

// Basic implementations.

type anyBasic struct{}

func (anyBasic) String() string       { return "typing.Any" }
func (anyBasic) basicSortKey() string { return "\x00any" } // sorts first

type primBasic struct{ name string } // runtime type name

func (b primBasic) String() string       { return displayPrim(b.name) }
func (b primBasic) basicSortKey() string { return "prim:" + b.name }

type listBasic struct{ elem Ty }

func (b listBasic) String() string {
	return "list[" + b.elem.String() + "]"
}
func (b listBasic) basicSortKey() string { return "list:" + b.elem.sortKey() }

type dictBasic struct{ key, val Ty }

func (b dictBasic) String() string {
	return "dict[" + b.key.String() + ", " + b.val.String() + "]"
}
func (b dictBasic) basicSortKey() string {
	return "dict:" + b.key.sortKey() + ":" + b.val.sortKey()
}

type setBasic struct{ elem Ty }

func (b setBasic) String() string       { return "set[" + b.elem.String() + "]" }
func (b setBasic) basicSortKey() string { return "set:" + b.elem.sortKey() }

type tupleBasic struct {
	elems []Ty // fixed arity, if of == nil
	of    *Ty  // unknown arity
}

func (b tupleBasic) String() string {
	if b.of != nil {
		if b.of.IsAny() {
			return "tuple"
		}
		return "tuple[" + b.of.String() + ", ...]"
	}
	var sb strings.Builder
	sb.WriteString("tuple[")
	for i, e := range b.elems {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(e.String())
	}
	sb.WriteString("]")
	return sb.String()
}

func (b tupleBasic) basicSortKey() string {
	if b.of != nil {
		return "tupleof:" + b.of.sortKey()
	}
	var sb strings.Builder
	sb.WriteString("tuple:")
	for _, e := range b.elems {
		sb.WriteString(e.sortKey())
		sb.WriteString(",")
	}
	return sb.String()
}

type iterBasic struct{ item Ty }

func (b iterBasic) String() string {
	if b.item.IsAny() {
		return "typing.Iterable"
	}
	return "typing.Iterable[" + b.item.String() + "]"
}
func (b iterBasic) basicSortKey() string { return "iter:" + b.item.sortKey() }

type callableBasic struct {
	name   string // "" for anonymous callables
	params *ParamSpec
	result Ty
	// specialFn, if non-nil, refines the result type of a call from
	// its argument types (e.g. sorted(xs) yields a list of xs's
	// element type). Returning false falls back to result. It does
	// not participate in identity (basicSortKey): two callables that
	// agree on name, params, and result are interchangeable.
	specialFn specialFunc
}

// A specialFunc computes a builtin call's result type from its
// argument types, or reports false to use the declared result type.
type specialFunc func(o *oracle, pos syntax.Position, args callArgs) (Ty, bool)

func (b callableBasic) String() string {
	if b.name != "" {
		return fmt.Sprintf("def %s", b.name)
	}
	return "typing.Callable"
}

func (b callableBasic) basicSortKey() string {
	return "callable:" + b.name + ":" + b.params.sortKey() + "->" + b.result.sortKey()
}

type typeBasic struct{}

func (typeBasic) String() string       { return "type" }
func (typeBasic) basicSortKey() string { return "type" }

type moduleBasic struct {
	name  string
	attrs map[string]Ty
}

func (b moduleBasic) String() string       { return "module " + b.name }
func (b moduleBasic) basicSortKey() string { return "module:" + b.name }

func (t Ty) sortKey() string {
	var sb strings.Builder
	for i, a := range t.alts {
		if i > 0 {
			sb.WriteString("|")
		}
		sb.WriteString(a.basicSortKey())
	}
	return sb.String()
}
