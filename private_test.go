// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"testing"

	"github.com/maxatome/go-testdeep/internal/test"
)

// Edge cases not tested elsewhere...

func TestBase(t *testing.T) {
	td := Base{}

	td.setLocation(200)
	if td.location.File != "???" && td.location.Line != 0 {
		t.Errorf("Location found! => %s", td.location)
	}
}

func TestTdSetResult(t *testing.T) {
	if tdSetResultKind(199).String() != "?" {
		t.Errorf("tdSetResultKind stringification failed => %s",
			tdSetResultKind(199))
	}
}

func TestPkgFunc(t *testing.T) {
	pkg, fn := pkgFunc("the/package.Foo")
	test.EqualStr(t, pkg, "the/package")
	test.EqualStr(t, fn, "Foo")

	pkg, fn = pkgFunc("the/package.(*T).Foo")
	test.EqualStr(t, pkg, "the/package")
	test.EqualStr(t, fn, "(*T).Foo")

	// Theorically not possible, but...
	pkg, fn = pkgFunc(".Foo")
	test.EqualStr(t, pkg, "")
	test.EqualStr(t, fn, "Foo")
}
