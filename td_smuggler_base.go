// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"reflect"
)

// tdSmugglerBase is the base class of all smuggler TestDeep operators.
type tdSmugglerBase struct {
	Base
	expectedValue reflect.Value
	isTestDeeper  bool
}

func newSmugglerBase(val interface{}, depth ...int) (ret tdSmugglerBase) {
	callDepth := 4
	if len(depth) > 0 {
		callDepth = depth[0]
	}
	ret.Base = NewBase(callDepth)

	// Initializes only if TestDeep operator. Other cases are specific.
	if _, ok := val.(TestDeep); ok {
		ret.expectedValue = reflect.ValueOf(val)
		ret.isTestDeeper = true
	}
	return
}
