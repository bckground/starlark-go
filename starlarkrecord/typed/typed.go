// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package typed contributes static types for the record and field
// builtins of go.starlark.net/starlarkrecord to a typecheck
// environment, so that the static checker understands module-level
// record definitions:
//
//	Rec = record(host=str, port=field(int, 80))
//	r = Rec(host="localhost")
//
// The checker then reports unknown fields (r.hosst), constructor
// argument errors (Rec(host=8), Rec()), and the annotation `x: Rec`
// matches exactly the instances of Rec, like the runtime matcher.
// Identity is nominal: two structurally identical record types do
// not match each other's instances.
//
// Use it alongside starlarkrecord.Predeclared:
//
//	env := typecheck.UniverseEnv()
//	typed.AddTypes(env)
package typed // import "go.starlark.net/starlarkrecord/typed"

import (
	"go.starlark.net/typecheck"
)

// AddTypes adds the static types of the record and field builtins to
// env. Their runtime counterparts are starlarkrecord.Predeclared.
func AddTypes(env typecheck.Env) {
	anyT := typecheck.Any()
	env["record"] = typecheck.WithFactory(
		typecheck.Function("record",
			&typecheck.ParamSpec{Params: []typecheck.Param{
				{Mode: typecheck.KwargsMode, Ty: anyT},
			}},
			anyT),
		recordFactory)
	env["field"] = typecheck.WithFactory(
		typecheck.Function("field",
			&typecheck.ParamSpec{Params: []typecheck.Param{
				{Name: "typ", Mode: typecheck.PosOrName, Required: true, Ty: anyT},
				{Name: "default", Mode: typecheck.PosOrName, Ty: anyT},
			}},
			anyT),
		fieldFactory)
}

// A recordCtorTy is the static type of the record type value itself:
// a callable constructing instances, with no attributes. Modeling it
// as a custom type (rather than a plain callable) keeps the checker
// lenient about the other things a runtime type value supports, such
// as union expressions (Rec | None).
type recordCtorTy struct {
	name   string
	params *typecheck.ParamSpec
	result typecheck.Ty
}

func (ct *recordCtorTy) TyName() string { return "type[" + ct.name + "]" }

func (ct *recordCtorTy) Attr(name string) (typecheck.Ty, bool) {
	return typecheck.Never(), false
}

func (ct *recordCtorTy) CallSignature() (*typecheck.ParamSpec, typecheck.Ty) {
	return ct.params, ct.result
}

// A recordTy is the static type of the instances of one record type.
type recordTy struct {
	name   string
	fields []field
}

type field struct {
	name string
	ty   typecheck.Ty
}

func (rt *recordTy) TyName() string { return rt.name }

func (rt *recordTy) Attr(name string) (typecheck.Ty, bool) {
	for _, f := range rt.fields {
		if f.name == name {
			return f.ty, true
		}
	}
	return typecheck.Never(), false
}

// recordFactory interprets Rec = record(host=str, port=field(int, 80)):
// the binding's type becomes the constructor (a callable with one
// named-only parameter per field, yielding an instance), and the
// binding denotes the instance type in annotation position.
func recordFactory(call *typecheck.FactoryCall) (typecheck.StaticValue, bool) {
	if len(call.Positional) > 0 {
		return typecheck.StaticValue{}, false
	}
	name := call.Name
	if name == "" {
		name = "record" // like the runtime's displayName fallback
	}
	rt := &recordTy{name: name}
	params := make([]typecheck.Param, 0, len(call.Named))
	for _, arg := range call.Named {
		var fty typecheck.Ty
		required := true
		if data, ok := arg.Value.Data(); ok {
			fs, ok := data.(*fieldSpec)
			if !ok {
				return typecheck.StaticValue{}, false
			}
			fty = fs.ty
			required = !fs.hasDefault
		} else if t, ok := arg.Value.DenotesType(); ok {
			fty = t
		} else {
			return typecheck.StaticValue{}, false
		}
		rt.fields = append(rt.fields, field{arg.Name, fty})
		params = append(params, typecheck.Param{
			Name: arg.Name, Mode: typecheck.NameOnly, Required: required, Ty: fty,
		})
	}
	instance := typecheck.Custom(rt)
	ctor := typecheck.Custom(&recordCtorTy{
		name:   name,
		params: &typecheck.ParamSpec{Params: params},
		result: instance,
	})
	return typecheck.ValueOf(ctor).Denoting(instance), true
}

// A fieldSpec is the static value of a field(typ, default) call.
type fieldSpec struct {
	ty         typecheck.Ty
	hasDefault bool
}

// fieldFactory interprets field(typ) and field(typ, default). The
// default value's type is checked by the runtime when the record
// type is created, so only its presence matters statically.
func fieldFactory(call *typecheck.FactoryCall) (typecheck.StaticValue, bool) {
	var typv *typecheck.StaticValue
	hasDefault := false
	if len(call.Positional) > 2 {
		return typecheck.StaticValue{}, false
	}
	if len(call.Positional) >= 1 {
		typv = &call.Positional[0]
	}
	if len(call.Positional) == 2 {
		hasDefault = true
	}
	for _, na := range call.Named {
		switch na.Name {
		case "typ":
			if typv != nil {
				return typecheck.StaticValue{}, false
			}
			v := na.Value
			typv = &v
		case "default":
			if hasDefault {
				return typecheck.StaticValue{}, false
			}
			hasDefault = true
		default:
			return typecheck.StaticValue{}, false
		}
	}
	if typv == nil {
		return typecheck.StaticValue{}, false
	}
	t, ok := typv.DenotesType()
	if !ok {
		return typecheck.StaticValue{}, false
	}
	return typecheck.DataValue(&fieldSpec{ty: t, hasDefault: hasDefault}), true
}
