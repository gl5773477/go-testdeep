// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"reflect"

	"github.com/maxatome/go-testdeep/internal/ctxerr"
	"github.com/maxatome/go-testdeep/internal/types"
)

type tdIsa struct {
	tdExpectedType
	checkImplement bool
}

var _ TestDeep = &tdIsa{}

// Isa operator checks the data type or whether data implements an
// interface or not.
//
// Typically type checks:
//   Isa(time.Time{})
//   Isa(&time.Time{})
//   Isa(map[string]time.Time{})
//
// For interfaces it is a bit more complicated, as:
//   fmt.Stringer(nil)
// is not an interface, but just nil... To bypass this golang
// limitation, Isa accepts pointers on interfaces. So checking that
// data implements fmt.Stringer interface should be written as:
//   Isa((*fmt.Stringer)(nil))
//
// Of course, in the latter case, if data type is *fmt.Stringer, Isa
// will match too (in fact before checking whether it implements
// fmt.Stringer or not.)
//
// TypeBehind method returns the reflect.Type of "model".
func Isa(model interface{}) TestDeep {
	modelType := reflect.ValueOf(model).Type()

	return &tdIsa{
		tdExpectedType: tdExpectedType{
			Base:         NewBase(3),
			expectedType: modelType,
		},
		checkImplement: modelType.Kind() == reflect.Ptr &&
			modelType.Elem().Kind() == reflect.Interface,
	}
}

func (i *tdIsa) Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error {
	gotType := got.Type()

	if gotType == i.expectedType {
		return nil
	}

	if i.checkImplement {
		if gotType.Implements(i.expectedType.Elem()) {
			return nil
		}
	}

	if ctx.BooleanError {
		return ctxerr.BooleanError
	}
	return ctx.CollectError(i.errorTypeMismatch(types.RawString(gotType.String())))
}

func (i *tdIsa) String() string {
	return i.expectedType.String()
}
