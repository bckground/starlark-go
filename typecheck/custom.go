// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// Custom types: the static analogue of starlark-rust's TyUser
// (starlark/src/typing/custom.rs and user.rs). A CustomTy describes a
// type the checker itself does not know, such as a record or enum
// type minted by a TypeFactory; optional capability interfaces add
// callability, indexing, and iteration, mirroring TyUserParams.
//
// CustomTy implementations live with the runtime types they describe
// (starlarkrecord, starlarkenum) or, for the universal error_tags, in
// this package. The static type and the runtime TypeMatcher must
// accept the same values; see TestRecordEnumAgreement.

import (
	"fmt"
	"sync/atomic"
)

// A CustomTy is an externally defined static type. Identity is
// nominal: each call of Custom mints a distinct type, and two Custom
// types intersect only if they originate from the same call. This
// mirrors the runtime, where a record type matches instances by
// pointer identity: two structurally identical record types do not
// intersect.
type CustomTy interface {
	// TyName returns the display name of the type, typically the
	// name of the global to which the minting call was assigned.
	TyName() string
	// Attr returns the type of the named attribute. ok=false means
	// the type certainly has no such attribute and the checker
	// reports an error; a lenient implementation returns
	// (Any(), true) for unknown names.
	Attr(name string) (Ty, bool)
}

// CustomCallable is implemented by CustomTy values that may be
// called, such as enum types.
type CustomCallable interface {
	CustomTy
	// CallSignature returns the parameters and result type of a call.
	CallSignature() (*ParamSpec, Ty)
}

// CustomIndexable is implemented by CustomTy values that support the
// [] operator.
type CustomIndexable interface {
	CustomTy
	// IndexResult returns the type of t[i] for an index of type
	// index, or false if the type cannot be indexed by it.
	IndexResult(index Ty) (Ty, bool)
}

// CustomIterable is implemented by CustomTy values that may be
// iterated.
type CustomIterable interface {
	CustomTy
	// IterItem returns the type of the elements yielded by iteration.
	IterItem() Ty
}

// Intersects reports whether a value could simultaneously inhabit
// both types. Compatibility checking uses intersection, not
// subtyping: a check fails only when the types certainly do not
// overlap. It is exported for CustomTy implementations (e.g. an enum
// type's IndexResult accepts any index that intersects int).
func Intersects(a, b Ty) bool { return intersects(a, b) }

var customSerial atomic.Uint64

// Custom returns a Ty whose sole alternative is the custom type c.
func Custom(c CustomTy) Ty {
	return Ty{alts: []Basic{customBasic{c: c, serial: customSerial.Add(1)}}}
}

type customBasic struct {
	c      CustomTy
	serial uint64 // nominal identity
}

func (b customBasic) String() string { return b.c.TyName() }
func (b customBasic) basicSortKey() string {
	return fmt.Sprintf("custom:%s:%d", b.c.TyName(), b.serial)
}
