// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package starlark

// This file defines the runtime representation of type annotations,
// closely following the type system of starlark-rust
// (https://github.com/facebook/starlark-rust/blob/main/docs/types.md).
//
// A type annotation is an ordinary Starlark expression that is
// evaluated (in the enclosing scope, when the def statement executes)
// to a value, which is then converted by TypeOf to a *Type, a compiled
// matcher. Type constructors such as list support indexing
// (list[int]) and the | operator (int | None) so that type
// expressions compose as first-class values.

import (
	"fmt"
	"strings"

	"go.starlark.net/syntax"
)

// A TypeMatcher is implemented by values that can act as Starlark type
// annotations with custom matching logic. Embedder-defined types
// (e.g. record and enum types) implement this interface to participate
// in runtime type checking.
type TypeMatcher interface {
	Value
	// Matches reports whether v conforms to this type.
	Matches(v Value) bool
}

// A TypeName is implemented by values that can act as Starlark type
// annotations matching purely by type name: a value v matches if
// v.Type() == TypeName().
type TypeName interface {
	Value
	// TypeName returns the name of the type this value denotes.
	TypeName() string
}

// EllipsisType is the type of Ellipsis, the value of the token `...`,
// which may appear only within type expressions (e.g. tuple[int, ...]).
type EllipsisType byte

// Ellipsis is the value of the token `...`.
const Ellipsis = EllipsisType(0)

func (EllipsisType) String() string        { return "..." }
func (EllipsisType) Type() string          { return "ellipsis" }
func (EllipsisType) Freeze()               {} // immutable
func (EllipsisType) Truth() Bool           { return True }
func (EllipsisType) Hash() (uint32, error) { return 1500009243, nil } // arbitrary

// A Type is a Starlark value representing a type, such as int,
// list[int], or int | None. It is the result of evaluating a type
// annotation, and the operand of isinstance.
type Type struct {
	repr typeRepr
}

var (
	_ Value       = (*Type)(nil)
	_ Comparable  = (*Type)(nil)
	_ TypeMatcher = (*Type)(nil)
)

func (t *Type) String() string        { return t.repr.str() }
func (t *Type) Type() string          { return "type" }
func (t *Type) Freeze()               {} // immutable
func (t *Type) Truth() Bool           { return True }
func (t *Type) Hash() (uint32, error) { return hashString(t.repr.str()), nil }

func (t *Type) CompareSameType(op syntax.Token, y_ Value, depth int) (bool, error) {
	y := y_.(*Type)
	switch op {
	case syntax.EQL:
		return t.String() == y.String(), nil
	case syntax.NEQ:
		return t.String() != y.String(), nil
	}
	return false, fmt.Errorf("%s %s %s not implemented", t.Type(), op, y.Type())
}

// Matches reports whether v conforms to type t.
func (t *Type) Matches(v Value) bool { return t.repr.matches(v) }

// Predefined types, exposed in Starlark as members of the typing
// module (see lib/typing).
var (
	TypingAny      = &Type{anyType{}}      // typing.Any: matches any value
	TypingNever    = &Type{neverType{}}    // typing.Never: matches no value
	TypingCallable = &Type{callableType{}} // typing.Callable: matches callable values
	TypingIterable = &Type{iterableType{}} // typing.Iterable: matches iterable values
	typeOfTypes    = &Type{typeType{}}     // the `type` constructor used as a type
	noneTypeValue  = &Type{noneTypeRepr{}} // None used as a type
)

// typeRepr is the internal representation of a Type: one alternative
// of the type grammar.
type typeRepr interface {
	str() string
	matches(v Value) bool
}

// typing.Any
type anyType struct{}

func (anyType) str() string          { return "typing.Any" }
func (anyType) matches(v Value) bool { return true }

// typing.Never
type neverType struct{}

func (neverType) str() string          { return "typing.Never" }
func (neverType) matches(v Value) bool { return false }

// None
type noneTypeRepr struct{}

func (noneTypeRepr) str() string          { return "None" }
func (noneTypeRepr) matches(v Value) bool { return v == None }

// type (matches values that are themselves types)
type typeType struct{}

func (typeType) str() string { return "type" }
func (typeType) matches(v Value) bool {
	_, err := TypeOf(v)
	return err == nil
}

// a type matched by name: int, bool, str, bytes, range, ...
// display is the name used in type expressions (e.g. "str");
// runtime is the corresponding Value.Type() string (e.g. "string").
type nameType struct {
	display string
	runtime string
}

func (t nameType) str() string          { return t.display }
func (t nameType) matches(v Value) bool { return v.Type() == t.runtime }

// float: like starlark-rust, a float annotation also accepts ints
// (numeric coercion).
type floatType struct{}

func (floatType) str() string { return "float" }
func (floatType) matches(v Value) bool {
	switch v.(type) {
	case Float, Int:
		return true
	}
	return false
}

// list or list[T]
type listType struct {
	elem *Type // nil: any list
}

func (t listType) str() string {
	if t.elem == nil {
		return "list"
	}
	return "list[" + t.elem.String() + "]"
}

func (t listType) matches(v Value) bool {
	list, ok := v.(*List)
	if !ok {
		return false
	}
	if t.elem != nil {
		for i := 0; i < list.Len(); i++ {
			if !t.elem.Matches(list.Index(i)) {
				return false
			}
		}
	}
	return true
}

// dict or dict[K, V]
type dictType struct {
	key, val *Type // both nil, or both non-nil
}

func (t dictType) str() string {
	if t.key == nil {
		return "dict"
	}
	return "dict[" + t.key.String() + ", " + t.val.String() + "]"
}

func (t dictType) matches(v Value) bool {
	dict, ok := v.(*Dict)
	if !ok {
		return false
	}
	if t.key != nil {
		for _, item := range dict.Items() {
			if !t.key.Matches(item[0]) || !t.val.Matches(item[1]) {
				return false
			}
		}
	}
	return true
}

// set or set[T]
type setType struct {
	elem *Type // nil: any set
}

func (t setType) str() string {
	if t.elem == nil {
		return "set"
	}
	return "set[" + t.elem.String() + "]"
}

func (t setType) matches(v Value) bool {
	set, ok := v.(*Set)
	if !ok {
		return false
	}
	if t.elem != nil {
		iter := set.Iterate()
		defer iter.Done()
		var e Value
		for iter.Next(&e) {
			if !t.elem.Matches(e) {
				return false
			}
		}
	}
	return true
}

// tuple, tuple[T1, T2], or tuple[T, ...]
type tupleType struct {
	elems []*Type // nil: any tuple
	open  bool    // tuple[T, ...]: every element matches elems[0]
}

func (t tupleType) str() string {
	if t.elems == nil {
		return "tuple"
	}
	var b strings.Builder
	b.WriteString("tuple[")
	for i, elem := range t.elems {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(elem.String())
	}
	if t.open {
		b.WriteString(", ...")
	}
	b.WriteString("]")
	return b.String()
}

func (t tupleType) matches(v Value) bool {
	tuple, ok := v.(Tuple)
	if !ok {
		return false
	}
	if t.elems == nil {
		return true
	}
	if t.open {
		for _, e := range tuple {
			if !t.elems[0].Matches(e) {
				return false
			}
		}
		return true
	}
	if len(tuple) != len(t.elems) {
		return false
	}
	for i, e := range tuple {
		if !t.elems[i].Matches(e) {
			return false
		}
	}
	return true
}

// T1 | T2
type unionType struct {
	alts []*Type
}

func (t unionType) str() string {
	var b strings.Builder
	for i, alt := range t.alts {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(alt.String())
	}
	return b.String()
}

func (t unionType) matches(v Value) bool {
	for _, alt := range t.alts {
		if alt.Matches(v) {
			return true
		}
	}
	return false
}

// typing.Callable or typing.Callable[[T1, T2], R]
type callableType struct {
	params []*Type // nil if unparameterized or `...`
	result *Type   // nil if unparameterized
}

func (t callableType) str() string {
	if t.result == nil {
		return "typing.Callable"
	}
	var b strings.Builder
	b.WriteString("typing.Callable[")
	if t.params == nil {
		b.WriteString("...")
	} else {
		b.WriteString("[")
		for i, p := range t.params {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(p.String())
		}
		b.WriteString("]")
	}
	b.WriteString(", ")
	b.WriteString(t.result.String())
	b.WriteString("]")
	return b.String()
}

func (t callableType) matches(v Value) bool {
	// At runtime only callability is checked; parameter and result
	// types are not (they would require calling the function).
	_, ok := v.(Callable)
	return ok
}

// typing.Iterable or typing.Iterable[T]
type iterableType struct {
	elem *Type // informational; ignored by the runtime match
}

func (t iterableType) str() string {
	if t.elem == nil {
		return "typing.Iterable"
	}
	return "typing.Iterable[" + t.elem.String() + "]"
}

func (t iterableType) matches(v Value) bool {
	// Only iterability is checked at runtime: checking element types
	// would require (observably) iterating the value.
	_, ok := v.(Iterable)
	return ok
}

// an embedder-defined matcher (e.g. record or enum types)
type matcherType struct {
	m TypeMatcher
}

func (t matcherType) str() string {
	// A matcher may implement TypeName to control how it displays in
	// type expressions (e.g. a record type's exported name).
	if n, ok := t.m.(TypeName); ok {
		return n.TypeName()
	}
	return t.m.String()
}
func (t matcherType) matches(v Value) bool { return t.m.Matches(v) }

// typeKind identifies the canonical type-constructor builtins
// (the universal functions int, str, list, dict, ...).
type typeKind int

const (
	kindInvalid typeKind = iota
	kindBool
	kindBytes
	kindDict
	kindFloat
	kindInt
	kindList
	kindRange
	kindSet
	kindStr
	kindTuple
	kindType
)

// typeConstructors maps the canonical type-constructor builtins,
// by identity, to their kind. It is populated by init in library.go
// and never mutated thereafter.
var typeConstructors = make(map[*Builtin]typeKind)

func typeOfKind(kind typeKind) *Type {
	switch kind {
	case kindBool:
		return &Type{nameType{"bool", "bool"}}
	case kindBytes:
		return &Type{nameType{"bytes", "bytes"}}
	case kindDict:
		return &Type{dictType{}}
	case kindFloat:
		return &Type{floatType{}}
	case kindInt:
		return &Type{nameType{"int", "int"}}
	case kindList:
		return &Type{listType{}}
	case kindRange:
		return &Type{nameType{"range", "range"}}
	case kindSet:
		return &Type{setType{}}
	case kindStr:
		return &Type{nameType{"str", "string"}}
	case kindTuple:
		return &Type{tupleType{}}
	case kindType:
		return typeOfTypes
	}
	panic(kind)
}

// TypeOf converts a Starlark value used in type-annotation position to
// a *Type. It is the analogue of eval_type in starlark-rust.
//
// It accepts: a *Type; None; the canonical type constructors (int,
// bool, str, float, bytes, list, dict, tuple, set, range, type);
// tuples of types (denoting a fixed-arity tuple type); and values
// implementing TypeMatcher or TypeName. Any other value is an error.
func TypeOf(v Value) (*Type, error) {
	switch v := v.(type) {
	case *Type:
		return v, nil
	case NoneType:
		return noneTypeValue, nil
	case String:
		return nil, fmt.Errorf("string literal is not allowed in type expression")
	case Tuple:
		elems := make([]*Type, len(v))
		for i, e := range v {
			t, err := TypeOf(e)
			if err != nil {
				return nil, err
			}
			elems[i] = t
		}
		return &Type{tupleType{elems: elems}}, nil
	case *List:
		// legacy union syntax: [T1, T2]
		if v.Len() < 2 {
			return nil, fmt.Errorf("a list in a type expression denotes a union and needs at least 2 elements")
		}
		result, err := TypeOf(v.Index(0))
		if err != nil {
			return nil, err
		}
		for i := 1; i < v.Len(); i++ {
			elem := v.Index(i)
			if !isTypeLike(elem) {
				return nil, fmt.Errorf("value of type %s is not a type", elem.Type())
			}
			result, err = typeUnion(result, elem)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case *Builtin:
		if kind := typeConstructors[v]; kind != kindInvalid {
			return typeOfKind(kind), nil
		}
	}
	// Embedder-defined types.
	if m, ok := v.(TypeMatcher); ok {
		return &Type{matcherType{m}}, nil
	}
	if n, ok := v.(TypeName); ok {
		name := n.TypeName()
		return &Type{nameType{name, name}}, nil
	}
	return nil, fmt.Errorf("value of type %s is not a type", v.Type())
}

// isTypeLike reports whether v can be converted to a type by TypeOf,
// excluding errors. It is used to decide whether `x | y` denotes a
// type union.
func isTypeLike(v Value) bool {
	switch v := v.(type) {
	case *Type, NoneType:
		return true
	case Tuple:
		for _, e := range v {
			if !isTypeLike(e) {
				return false
			}
		}
		return len(v) > 0
	case *Builtin:
		return typeConstructors[v] != kindInvalid
	}
	if _, ok := v.(TypeMatcher); ok {
		return true
	}
	if _, ok := v.(TypeName); ok {
		return true
	}
	return false
}

// typeUnion returns the type denoted by `x | y` where both operands
// are type-like. Nested unions are flattened and duplicates removed.
func typeUnion(x, y Value) (*Type, error) {
	xt, err := TypeOf(x)
	if err != nil {
		return nil, err
	}
	yt, err := TypeOf(y)
	if err != nil {
		return nil, err
	}
	var alts []*Type
	seen := make(map[string]bool)
	add := func(t *Type) {
		if u, ok := t.repr.(unionType); ok {
			for _, alt := range u.alts {
				if !seen[alt.String()] {
					seen[alt.String()] = true
					alts = append(alts, alt)
				}
			}
		} else if !seen[t.String()] {
			seen[t.String()] = true
			alts = append(alts, t)
		}
	}
	add(xt)
	add(yt)
	if len(alts) == 1 {
		return alts[0], nil
	}
	return &Type{unionType{alts}}, nil
}

// typeIndexElems normalizes an index operand into a list of values:
// T[a] yields [a]; T[a, b] yields [a, b].
func typeIndexElems(index Value) []Value {
	if tuple, ok := index.(Tuple); ok {
		return tuple
	}
	return []Value{index}
}

// parameterizeBuiltin implements indexing of a canonical type
// constructor, e.g. list[int], dict[str, int], tuple[int, ...].
func parameterizeBuiltin(kind typeKind, name string, index Value) (Value, error) {
	args := typeIndexElems(index)
	switch kind {
	case kindList, kindSet:
		if len(args) != 1 {
			return nil, fmt.Errorf("%s[...] expects exactly 1 type argument, got %d", name, len(args))
		}
		elem, err := TypeOf(args[0])
		if err != nil {
			return nil, err
		}
		if kind == kindList {
			return &Type{listType{elem}}, nil
		}
		return &Type{setType{elem}}, nil

	case kindDict:
		if len(args) != 2 {
			return nil, fmt.Errorf("dict[...] expects exactly 2 type arguments, got %d", len(args))
		}
		key, err := TypeOf(args[0])
		if err != nil {
			return nil, err
		}
		val, err := TypeOf(args[1])
		if err != nil {
			return nil, err
		}
		return &Type{dictType{key, val}}, nil

	case kindTuple:
		// tuple[T1, T2, ...Tn] or tuple[T, ...]
		if len(args) == 2 {
			if _, ok := args[1].(EllipsisType); ok {
				elem, err := TypeOf(args[0])
				if err != nil {
					return nil, err
				}
				return &Type{tupleType{elems: []*Type{elem}, open: true}}, nil
			}
		}
		elems := make([]*Type, len(args))
		for i, arg := range args {
			if _, ok := arg.(EllipsisType); ok {
				return nil, fmt.Errorf("`...` is only allowed as the last of two arguments of tuple[...]")
			}
			t, err := TypeOf(arg)
			if err != nil {
				return nil, err
			}
			elems[i] = t
		}
		return &Type{tupleType{elems: elems}}, nil
	}
	return nil, fmt.Errorf("%s is not parameterizable", name)
}

// typeIndex implements indexing of a *Type value, which is permitted
// only for typing.Callable and typing.Iterable.
func typeIndex(t *Type, index Value) (Value, error) {
	switch t.repr.(type) {
	case callableType:
		// typing.Callable[[T1, T2], R] or typing.Callable[..., R]
		args := typeIndexElems(index)
		if len(args) != 2 {
			return nil, fmt.Errorf("typing.Callable[...] expects exactly 2 arguments, got %d", len(args))
		}
		result, err := TypeOf(args[1])
		if err != nil {
			return nil, err
		}
		switch params := args[0].(type) {
		case EllipsisType:
			return &Type{callableType{params: nil, result: result}}, nil
		case *List:
			types := make([]*Type, params.Len())
			for i := 0; i < params.Len(); i++ {
				t, err := TypeOf(params.Index(i))
				if err != nil {
					return nil, err
				}
				types[i] = t
			}
			return &Type{callableType{params: types, result: result}}, nil
		}
		return nil, fmt.Errorf("first argument of typing.Callable[...] must be a list of types or `...`")

	case iterableType:
		args := typeIndexElems(index)
		if len(args) != 1 {
			return nil, fmt.Errorf("typing.Iterable[...] expects exactly 1 type argument, got %d", len(args))
		}
		elem, err := TypeOf(args[0])
		if err != nil {
			return nil, err
		}
		return &Type{iterableType{elem}}, nil
	}
	return nil, fmt.Errorf("%s is not parameterizable", t.String())
}

// isinstance(v, t) reports whether value v matches type t.
func isinstance(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var v, t Value
	if err := UnpackPositionalArgs("isinstance", args, kwargs, 2, &v, &t); err != nil {
		return nil, err
	}
	ty, err := TypeOf(t)
	if err != nil {
		return nil, fmt.Errorf("isinstance: %w", err)
	}
	return Bool(ty.Matches(v)), nil
}

// eval_type(t) converts a value in type position to a type value.
func evalType(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	var t Value
	if err := UnpackPositionalArgs("eval_type", args, kwargs, 1, &t); err != nil {
		return nil, err
	}
	ty, err := TypeOf(t)
	if err != nil {
		return nil, fmt.Errorf("eval_type: %w", err)
	}
	return ty, nil
}
