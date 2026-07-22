// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package typing defines the Starlark 'typing' module, which provides
// the special types of the type-annotation system, mirroring the
// typing module of starlark-rust.
package typing // import "go.starlark.net/lib/typing"

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module typing is a Starlark module of special type values.
//
//	typing = module(
//	   Any,       # matches any value
//	   Never,     # matches no value
//	   Callable,  # matches callable values; Callable[[T1, T2], R]
//	   Iterable,  # matches iterable values; Iterable[T]
//	)
//
// To use it in Starlark programs, add it to the predeclared
// environment:
//
//	predeclared := starlark.StringDict{"typing": typing.Module}
var Module = &starlarkstruct.Module{
	Name: "typing",
	Members: starlark.StringDict{
		"Any":      starlark.TypingAny,
		"Never":    starlark.TypingNever,
		"Callable": starlark.TypingCallable,
		"Iterable": starlark.TypingIterable,
	},
}
