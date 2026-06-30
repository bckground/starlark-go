// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// Callable parameter specifications, mirroring starlark-rust's
// ParamSpec (starlark/src/typing/callable_param.rs).

import "strings"

// ParamMode describes how a parameter may be filled.
type ParamMode uint8

const (
	PosOnly    ParamMode = iota // positional-only
	PosOrName                   // positional or named
	NameOnly                    // named-only (after * or *args)
	ArgsMode                    // *args
	KwargsMode                  // **kwargs
)

// A Param is one parameter of a callable type.
type Param struct {
	Name     string // "" for PosOnly and ArgsMode
	Mode     ParamMode
	Required bool
	Ty       Ty // for *args: the element type; for **kwargs: the value type
}

// A ParamSpec describes the parameters of a callable type.
type ParamSpec struct {
	Params []Param
	any    bool // accepts any arguments (signature unknown)
}

// AnyParams returns a ParamSpec that accepts any arguments.
func AnyParams() *ParamSpec { return &ParamSpec{any: true} }

// IsAny reports whether the spec accepts any arguments.
func (ps *ParamSpec) IsAny() bool { return ps == nil || ps.any }

// PositionalOnly returns a ParamSpec of n required positional-only
// parameters with the given types.
func PositionalOnly(types ...Ty) *ParamSpec {
	ps := &ParamSpec{Params: make([]Param, len(types))}
	for i, ty := range types {
		ps.Params[i] = Param{Mode: PosOnly, Required: true, Ty: ty}
	}
	return ps
}

func (ps *ParamSpec) sortKey() string {
	if ps.IsAny() {
		return "*"
	}
	var sb strings.Builder
	for _, p := range ps.Params {
		sb.WriteString(p.Name)
		sb.WriteByte(byte('0' + p.Mode))
		if p.Required {
			sb.WriteString("!")
		}
		sb.WriteString(p.Ty.sortKey())
		sb.WriteString(";")
	}
	return sb.String()
}
