// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/internal/ctxerr"
	"github.com/maxatome/go-testdeep/internal/location"
	"github.com/maxatome/go-testdeep/internal/types"
)

var (
	testDeeper         = reflect.TypeOf((*TestDeep)(nil)).Elem()
	stringerInterface  = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	errorInterface     = reflect.TypeOf((*error)(nil)).Elem()
	timeType           = reflect.TypeOf(time.Time{})
	intType            = reflect.TypeOf(int(0))
	smuggledGotType    = reflect.TypeOf(SmuggledGot{})
	smuggledGotPtrType = reflect.TypeOf((*SmuggledGot)(nil))
)

// TestingT is the minimal interface used by Cmp to report errors. It
// is commonly implemented by *testing.T and testing.TB.
type TestingT interface {
	Error(args ...interface{})
	Fatal(args ...interface{})
	Helper()
}

// TestingFT (aka. TestingF<ull>T) is the interface used by T to
// delegate common *testing.T functions to it. Of course, *testing.T
// implements it.
type TestingFT interface {
	TestingT
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool
	Run(name string, f func(t *testing.T)) bool
}

// TestDeep is the representation of a testdeep operator. It is not
// intended to be used directly, but through Cmp* functions.
type TestDeep interface {
	types.TestDeepStringer
	Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error
	location.GetLocationer
	setLocation(int)
	HandleInvalid() bool
	TypeBehind() reflect.Type
}

// Base is a base type providing some methods needed by the TestDeep
// interface.
type Base struct {
	types.TestDeepStamp
	location location.Location
}

func pkgFunc(full string) (string, string) {
	// the/package.Foo      → "the/package", "Foo"
	dotPos := strings.LastIndex(full, ".")
	pkg, fn := full[:dotPos], full[dotPos+1:]

	// the/package.(*T).Foo → "the/package", "(*T).Foo"
	if dotPos > 0 && pkg[len(pkg)-1] == ')' {
		dotPos = strings.LastIndex(pkg, ".")
		pkg, fn = full[:dotPos], full[dotPos+1:]
	}
	return pkg, fn
}

func (t *Base) setLocation(callDepth int) {
	var ok bool
	t.location, ok = location.New(callDepth)
	if !ok {
		t.location.File = "???"
		t.location.Line = 0
		return
	}

	// Here package is github.com/maxatome/go-testdeep, or its vendored
	// counterpart
	var pkg string
	pkg, t.location.Func = pkgFunc(t.location.Func)

	// Try to go one level upper, if we are still in go-testdeep package
	cmpLoc, ok := location.New(callDepth + 1)
	if ok {
		cmpPkg, _ := pkgFunc(cmpLoc.Func)
		if cmpPkg == pkg {
			t.location.File = cmpLoc.File
			t.location.Line = cmpLoc.Line
			t.location.BehindCmp = true
		}
	}
}

// GetLocation returns a copy of the location.Location where the TestDeep
// operator has been created.
func (t *Base) GetLocation() location.Location {
	return t.location
}

// HandleInvalid tells testdeep internals that this operator does not
// handle nil values directly.
func (t Base) HandleInvalid() bool {
	return false
}

// TypeBehind returns the type handled by the operator. Only few operators
// knows the type they are handling. If they do not know, nil is
// returned.
func (t Base) TypeBehind() reflect.Type {
	return nil
}

// NewBase returns a new Base struct with location.Location set to the
// "callDepth" depth.
func NewBase(callDepth int) (b Base) {
	b.setLocation(callDepth)
	return
}

// BaseOKNil is a base type providing some methods needed by the TestDeep
// interface, for operators handling nil values.
type BaseOKNil struct {
	Base
}

// HandleInvalid tells testdeep internals that this operator handles
// nil values directly.
func (t BaseOKNil) HandleInvalid() bool {
	return true
}

// NewBaseOKNil returns a new BaseOKNil struct with location.Location set to
// the "callDepth" depth.
func NewBaseOKNil(callDepth int) (b BaseOKNil) {
	b.setLocation(callDepth)
	return
}
