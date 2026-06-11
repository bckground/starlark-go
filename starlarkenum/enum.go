// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package starlarkenum defines the Starlark 'enum' builtin, a port of
// the enumeration types of starlark-rust
// (https://github.com/facebook/starlark-rust/blob/main/docs/types.md).
//
// An enum type is a callable value that restricts values to a fixed
// set of strings:
//
//	Color = enum("red", "green", "blue")
//	c = Color("red")
//	c.value      # "red"
//	c.index      # 0
//	Color[1]     # Color("green")
//	Color.values()  # ["red", "green", "blue"]
//	len(Color)   # 3
//
// Enum types participate in type annotations: a value matches the
// annotation Color if and only if it is an element of that exact
// enum type.
//
// Like starlark-rust's LibraryExtension::EnumType, the builtin is
// opt-in; add it to the predeclared environment:
//
//	predeclared := starlark.StringDict{"enum": starlarkenum.Predeclared["enum"]}
package starlarkenum // import "go.starlark.net/starlarkenum"

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Predeclared contains the values of this package's builtins.
var Predeclared = starlark.StringDict{
	"enum": starlark.NewBuiltin("enum", MakeEnumType),
}

// An EnumType is the result of enum(...): a callable that returns the
// interned EnumValue for a given element string.
type EnumType struct {
	name     string // exported name; "" until assigned to a global
	elements []*EnumValue
	index    map[string]int
}

var (
	_ starlark.Callable    = (*EnumType)(nil)
	_ starlark.Indexable   = (*EnumType)(nil)
	_ starlark.Iterable    = (*EnumType)(nil)
	_ starlark.HasAttrs    = (*EnumType)(nil)
	_ starlark.TypeMatcher = (*EnumType)(nil)
	_ starlark.TypeName    = (*EnumType)(nil)
	_ starlark.Exportable  = (*EnumType)(nil)
)

// MakeEnumType implements the enum(*values) builtin.
// All values must be distinct strings.
func MakeEnumType(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(kwargs) > 0 {
		return nil, fmt.Errorf("enum: unexpected keyword arguments")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("enum: at least one value is required")
	}
	et := &EnumType{
		elements: make([]*EnumValue, len(args)),
		index:    make(map[string]int, len(args)),
	}
	for i, arg := range args {
		s, ok := starlark.AsString(arg)
		if !ok {
			return nil, fmt.Errorf("enum: got %s, want string", arg.Type())
		}
		if _, dup := et.index[s]; dup {
			return nil, fmt.Errorf("enum: duplicate value %q", s)
		}
		et.index[s] = i
		et.elements[i] = &EnumValue{typ: et, value: s, index: i}
	}
	return et, nil
}

func (et *EnumType) String() string {
	var b strings.Builder
	b.WriteString("enum(")
	for i, e := range et.elements {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(starlark.String(e.value).String())
	}
	b.WriteString(")")
	return b.String()
}

func (et *EnumType) Type() string         { return "enum_type" }
func (et *EnumType) Truth() starlark.Bool { return starlark.True }
func (et *EnumType) Freeze()              {} // immutable
func (et *EnumType) Name() string         { return et.displayName() }

func (et *EnumType) Hash() (uint32, error) {
	return starlark.String(et.String()).Hash()
}

// ExportAs records the name of the global variable to which the type
// is first assigned, like starlark-rust's export_as.
func (et *EnumType) ExportAs(name string) {
	if et.name == "" {
		et.name = name
	}
}

// TypeName returns the name used when the enum type appears in a type
// annotation.
func (et *EnumType) TypeName() string { return et.displayName() }

func (et *EnumType) displayName() string {
	if et.name != "" {
		return et.name
	}
	return "enum"
}

// Matches reports whether v is an element of this exact enum type.
func (et *EnumType) Matches(v starlark.Value) bool {
	ev, ok := v.(*EnumValue)
	return ok && ev.typ == et
}

// CallInternal returns the interned element for the given string, or
// the value itself if it is already an element of this enum.
func (et *EnumType) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var v starlark.Value
	if err := starlark.UnpackPositionalArgs(et.displayName(), args, kwargs, 1, &v); err != nil {
		return nil, err
	}
	if ev, ok := v.(*EnumValue); ok && ev.typ == et {
		return ev, nil
	}
	if s, ok := starlark.AsString(v); ok {
		if i, ok := et.index[s]; ok {
			return et.elements[i], nil
		}
		return nil, fmt.Errorf("%s: unknown enum element %q", et.displayName(), s)
	}
	return nil, fmt.Errorf("%s: got %s, want string", et.displayName(), v.Type())
}

func (et *EnumType) Len() int                   { return len(et.elements) }
func (et *EnumType) Index(i int) starlark.Value { return et.elements[i] }

func (et *EnumType) Iterate() starlark.Iterator { return &enumIterator{et, 0} }

type enumIterator struct {
	et *EnumType
	i  int
}

func (it *enumIterator) Next(p *starlark.Value) bool {
	if it.i < len(it.et.elements) {
		*p = it.et.elements[it.i]
		it.i++
		return true
	}
	return false
}

func (it *enumIterator) Done() {}

func (et *EnumType) Attr(name string) (starlark.Value, error) {
	if name == "values" {
		return starlark.NewBuiltin("values", enumValues).BindReceiver(et), nil
	}
	// element access by name: Color.red
	if i, ok := et.index[name]; ok {
		return et.elements[i], nil
	}
	return nil, nil
}

func (et *EnumType) AttrNames() []string {
	names := make([]string, 0, len(et.elements)+1)
	names = append(names, "values")
	for _, e := range et.elements {
		names = append(names, e.value)
	}
	return names
}

func enumValues(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs("values", args, kwargs, 0); err != nil {
		return nil, err
	}
	et := b.Receiver().(*EnumType)
	values := make([]starlark.Value, len(et.elements))
	for i, e := range et.elements {
		values[i] = starlark.String(e.value)
	}
	return starlark.NewList(values), nil
}

// An EnumValue is an element of an EnumType. All values are created
// when the type is constructed and interned: calling the type returns
// the existing element.
type EnumValue struct {
	typ   *EnumType
	value string
	index int
}

var (
	_ starlark.HasAttrs   = (*EnumValue)(nil)
	_ starlark.Comparable = (*EnumValue)(nil)
)

// EnumType returns the type of which ev is an element.
func (ev *EnumValue) EnumType() *EnumType { return ev.typ }

// Value returns the element's string value.
func (ev *EnumValue) Value() string { return ev.value }

// Index returns the element's position within its type.
func (ev *EnumValue) Index() int { return ev.index }

func (ev *EnumValue) String() string {
	return fmt.Sprintf("%s(%s)", ev.typ.instancePrefix(), starlark.String(ev.value).String())
}

func (et *EnumType) instancePrefix() string {
	if et.name != "" {
		return et.name
	}
	return "enum()"
}

func (ev *EnumValue) Type() string         { return "enum" }
func (ev *EnumValue) Truth() starlark.Bool { return starlark.True }
func (ev *EnumValue) Freeze()              {} // immutable

// Hash delegates to the hash of the element string, so enum values
// and their strings may be used interchangeably as dict keys.
func (ev *EnumValue) Hash() (uint32, error) {
	return starlark.String(ev.value).Hash()
}

func (ev *EnumValue) Attr(name string) (starlark.Value, error) {
	switch name {
	case "value":
		return starlark.String(ev.value), nil
	case "index":
		return starlark.MakeInt(ev.index), nil
	}
	return nil, nil
}

func (ev *EnumValue) AttrNames() []string { return []string{"index", "value"} }

func (x *EnumValue) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	y := y_.(*EnumValue)
	eq := x.typ == y.typ && x.index == y.index
	switch op {
	case syntax.EQL:
		return eq, nil
	case syntax.NEQ:
		return !eq, nil
	default:
		return false, fmt.Errorf("%s %s %s not implemented", x.Type(), op, y.Type())
	}
}
