// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"reflect"

	"github.com/maxatome/go-testdeep/internal/ctxerr"
)

type tdIgnore struct {
	BaseOKNil
}

var ignoreSingleton TestDeep = &tdIgnore{}

// Ignore operator is always true, whatever data is. It is useful when
// comparing a slice and wanting to ignore some indexes, for example.
func Ignore() TestDeep {
	// NewBase() useless
	return ignoreSingleton
}

func (i *tdIgnore) Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error {
	return nil
}

func (i *tdIgnore) String() string {
	return "Ignore()"
}
