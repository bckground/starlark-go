// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package starlarkrecord defines the Starlark 'record' and 'field'
// builtins, a port of the record types of starlark-rust
// (https://github.com/facebook/starlark-rust/blob/main/docs/types.md).
//
// A record type is a callable value that constructs instances with a
// fixed set of typed fields:
//
//	MyRecord = record(host=str, port=field(int, 80))
//	v = MyRecord(host="localhost")
//	v.host                          # "localhost"
//	v.port                          # 80
//
// Record types participate in type annotations: a value matches the
// annotation MyRecord if and only if it is an instance of that exact
// record type.
//
// Like starlark-rust's LibraryExtension::RecordType, the builtins are
// opt-in; add them to the predeclared environment:
//
//	predeclared := starlark.StringDict{}
//	for name, v := range starlarkrecord.Predeclared {
//		predeclared[name] = v
//	}
package starlarkrecord // import "go.starlark.net/starlarkrecord"

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Predeclared contains the values of this package's builtins.
var Predeclared = starlark.StringDict{
	"record": starlark.NewBuiltin("record", MakeRecordType),
	"field":  starlark.NewBuiltin("field", MakeField),
}

// A RecordType is the result of record(...): a callable that
// constructs Record instances with a fixed set of typed fields.
type RecordType struct {
	name   string // exported name; "" until assigned to a global
	fields []fieldSpec
}

type fieldSpec struct {
	name string
	typ  *starlark.Type
	def  starlark.Value // nil => required
}

var (
	_ starlark.Callable    = (*RecordType)(nil)
	_ starlark.TypeMatcher = (*RecordType)(nil)
	_ starlark.TypeName    = (*RecordType)(nil)
	_ starlark.Exportable  = (*RecordType)(nil)
)

// MakeRecordType implements the record(**kwargs) builtin.
// Each keyword argument is a type, or a Field created by field().
func MakeRecordType(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("record: unexpected positional arguments")
	}
	rt := &RecordType{fields: make([]fieldSpec, 0, len(kwargs))}
	for _, kwarg := range kwargs {
		name := string(kwarg[0].(starlark.String))
		switch v := kwarg[1].(type) {
		case *Field:
			rt.fields = append(rt.fields, fieldSpec{name, v.typ, v.def})
		default:
			typ, err := starlark.TypeOf(v)
			if err != nil {
				return nil, fmt.Errorf("record: invalid type for field %q: %w", name, err)
			}
			rt.fields = append(rt.fields, fieldSpec{name, typ, nil})
		}
	}
	return rt, nil
}

func (rt *RecordType) String() string {
	var b strings.Builder
	b.WriteString("record(")
	for i, f := range rt.fields {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(f.name)
		b.WriteString("=")
		if f.def != nil {
			fmt.Fprintf(&b, "field(%s, %s)", f.typ.String(), f.def.String())
		} else {
			b.WriteString(f.typ.String())
		}
	}
	b.WriteString(")")
	return b.String()
}

func (rt *RecordType) Type() string          { return "record_type" }
func (rt *RecordType) Truth() starlark.Bool  { return starlark.True }
func (rt *RecordType) Name() string          { return rt.displayName() }
func (rt *RecordType) Hash() (uint32, error) { return starlark.String(rt.String()).Hash() }

func (rt *RecordType) Freeze() {
	for _, f := range rt.fields {
		if f.def != nil {
			f.def.Freeze()
		}
	}
}

// ExportAs records the name of the global variable to which the type
// is first assigned, like starlark-rust's export_as.
func (rt *RecordType) ExportAs(name string) {
	if rt.name == "" {
		rt.name = name
	}
}

// TypeName returns the name used when the record type appears in a
// type annotation.
func (rt *RecordType) TypeName() string { return rt.displayName() }

func (rt *RecordType) displayName() string {
	if rt.name != "" {
		return rt.name
	}
	return "record"
}

func (rt *RecordType) instanceName() string {
	if rt.name != "" {
		return rt.name
	}
	return "anon"
}

// Matches reports whether v is an instance of this exact record type.
func (rt *RecordType) Matches(v starlark.Value) bool {
	r, ok := v.(*Record)
	return ok && r.rtype == rt
}

// CallInternal constructs a record instance. Arguments are named-only;
// every value is checked against its field's type.
func (rt *RecordType) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("%s: record instances are created with named arguments only", rt.displayName())
	}
	values := make([]starlark.Value, len(rt.fields))
	for _, kwarg := range kwargs {
		name := string(kwarg[0].(starlark.String))
		i := rt.fieldIndex(name)
		if i < 0 {
			return nil, fmt.Errorf("%s: unexpected field %q", rt.displayName(), name)
		}
		if values[i] != nil {
			return nil, fmt.Errorf("%s: duplicate field %q", rt.displayName(), name)
		}
		v := kwarg[1]
		if !rt.fields[i].typ.Matches(v) {
			return nil, fmt.Errorf("%s: for field %q: Value `%s` of type `%s` does not match the type annotation `%s`",
				rt.displayName(), name, v.String(), v.Type(), rt.fields[i].typ.String())
		}
		values[i] = v
	}
	for i, f := range rt.fields {
		if values[i] == nil {
			if f.def == nil {
				return nil, fmt.Errorf("%s: missing field %q", rt.displayName(), f.name)
			}
			values[i] = f.def
		}
	}
	return &Record{rtype: rt, values: values}, nil
}

func (rt *RecordType) fieldIndex(name string) int {
	for i, f := range rt.fields {
		if f.name == name {
			return i
		}
	}
	return -1
}

// A Field is the result of field(typ, default): a typed record field
// specification with an optional default value. It is only useful as
// an argument to record().
type Field struct {
	typ *starlark.Type
	def starlark.Value // nil if absent
}

// MakeField implements the field(typ, default=...) builtin.
// The default value, if given, is checked against typ immediately.
func MakeField(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var typv, def starlark.Value
	if err := starlark.UnpackArgs("field", args, kwargs, "typ", &typv, "default?", &def); err != nil {
		return nil, err
	}
	typ, err := starlark.TypeOf(typv)
	if err != nil {
		return nil, fmt.Errorf("field: %w", err)
	}
	if def != nil && !typ.Matches(def) {
		return nil, fmt.Errorf("field: default value `%s` of type `%s` does not match the type annotation `%s`",
			def.String(), def.Type(), typ.String())
	}
	return &Field{typ: typ, def: def}, nil
}

func (f *Field) String() string {
	if f.def != nil {
		return fmt.Sprintf("field(%s, %s)", f.typ.String(), f.def.String())
	}
	return fmt.Sprintf("field(%s)", f.typ.String())
}

func (f *Field) Type() string         { return "field" }
func (f *Field) Truth() starlark.Bool { return starlark.True }
func (f *Field) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: field")
}

func (f *Field) Freeze() {
	if f.def != nil {
		f.def.Freeze()
	}
}

// A Record is an instance of a RecordType.
type Record struct {
	rtype  *RecordType
	values []starlark.Value // parallel to rtype.fields
	frozen bool
}

var (
	_ starlark.HasAttrs   = (*Record)(nil)
	_ starlark.Comparable = (*Record)(nil)
)

// RecordType returns the type of which r is an instance.
func (r *Record) RecordType() *RecordType { return r.rtype }

func (r *Record) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "record[%s](", r.rtype.instanceName())
	for i, f := range r.rtype.fields {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(f.name)
		b.WriteString("=")
		b.WriteString(r.values[i].String())
	}
	b.WriteString(")")
	return b.String()
}

func (r *Record) Type() string         { return "record" }
func (r *Record) Truth() starlark.Bool { return starlark.True }

func (r *Record) Freeze() {
	if !r.frozen {
		r.frozen = true
		for _, v := range r.values {
			v.Freeze()
		}
	}
}

func (r *Record) Hash() (uint32, error) {
	// Same algorithm as Tuple.Hash.
	x, m := uint32(8731), uint32(9839)
	for _, v := range r.values {
		y, err := v.Hash()
		if err != nil {
			return 0, err
		}
		x = x ^ y*m
		m += 7349
	}
	return x, nil
}

func (r *Record) Attr(name string) (starlark.Value, error) {
	if i := r.rtype.fieldIndex(name); i >= 0 {
		return r.values[i], nil
	}
	return nil, nil // no such attribute
}

func (r *Record) AttrNames() []string {
	names := make([]string, len(r.rtype.fields))
	for i, f := range r.rtype.fields {
		names[i] = f.name
	}
	return names
}

func (x *Record) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	y := y_.(*Record)
	switch op {
	case syntax.EQL:
		return recordsEqual(x, y, depth)
	case syntax.NEQ:
		eq, err := recordsEqual(x, y, depth)
		return !eq, err
	default:
		return false, fmt.Errorf("%s %s %s not implemented", x.Type(), op, y.Type())
	}
}

func recordsEqual(x, y *Record, depth int) (bool, error) {
	if x.rtype != y.rtype {
		return false, nil
	}
	for i := range x.values {
		eq, err := starlark.EqualDepth(x.values[i], y.values[i], depth-1)
		if err != nil {
			return false, err
		}
		if !eq {
			return false, nil
		}
	}
	return true, nil
}
