// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"fmt"
	"testing"

	"github.com/maxatome/go-testdeep/internal/test"
)

func ExampleT_True() {
	t := NewT(&testing.T{})

	got := true
	ok := t.True(got, "check that got is true!")
	fmt.Println(ok)

	got = false
	ok = t.True(got, "check that got is true!")
	fmt.Println(ok)

	// Output:
	// true
	// false
}

func ExampleT_False() {
	t := NewT(&testing.T{})

	got := false
	ok := t.False(got, "check that got is false!")
	fmt.Println(ok)

	got = true
	ok = t.False(got, "check that got is false!")
	fmt.Println(ok)

	// Output:
	// true
	// false
}

func ExampleT_CmpError() {
	t := NewT(&testing.T{})

	got := fmt.Errorf("Error #%d", 42)
	ok := t.CmpError(got, "An error occurred")
	fmt.Println(ok)

	got = nil
	ok = t.CmpError(got, "An error occurred") // fails
	fmt.Println(ok)

	// Output:
	// true
	// false
}

func ExampleT_CmpNoError() {
	t := NewT(&testing.T{})

	got := fmt.Errorf("Error #%d", 42)
	ok := t.CmpNoError(got, "An error occurred") // fails
	fmt.Println(ok)

	got = nil
	ok = t.CmpNoError(got, "An error occurred")
	fmt.Println(ok)

	// Output:
	// false
	// true
}

func ExampleT_CmpPanic() {
	t := NewT(&testing.T{})

	ok := t.CmpPanic(func() { panic("I am panicking!") }, "I am panicking!",
		"Checks for panic")
	fmt.Println("checks exact panic() string:", ok)

	// Can use TestDeep operator too
	ok = t.CmpPanic(func() { panic("I am panicking!") }, Contains("panicking!"),
		"Checks for panic")
	fmt.Println("checks panic() sub-string:", ok)

	// Can detect panic(nil)
	ok = t.CmpPanic(func() { panic(nil) }, nil, "Checks for panic(nil)")
	fmt.Println("checks for panic(nil):", ok)

	// As well as structured data panic
	type PanicStruct struct {
		Error string
		Code  int
	}

	ok = t.CmpPanic(
		func() {
			panic(PanicStruct{Error: "Memory violation", Code: 11})
		},
		PanicStruct{
			Error: "Memory violation",
			Code:  11,
		})
	fmt.Println("checks exact panic() struct:", ok)

	// or combined with TestDeep operators too
	ok = t.CmpPanic(
		func() {
			panic(PanicStruct{Error: "Memory violation", Code: 11})
		},
		Struct(PanicStruct{}, StructFields{
			"Code": Between(10, 20),
		}))
	fmt.Println("checks panic() struct against TestDeep operators:", ok)

	// Of course, do not panic = test failure, even for expected nil
	// panic parameter
	ok = t.CmpPanic(func() {}, nil)
	fmt.Println("checks a panic occurred:", ok)

	// Output:
	// checks exact panic() string: true
	// checks panic() sub-string: true
	// checks for panic(nil): true
	// checks exact panic() struct: true
	// checks panic() struct against TestDeep operators: true
	// checks a panic occurred: false
}

func ExampleT_CmpNotPanic() {
	t := NewT(&testing.T{})

	ok := t.CmpNotPanic(func() {}, nil)
	fmt.Println("checks a panic DID NOT occur:", ok)

	// Classic panic
	ok = t.CmpNotPanic(func() { panic("I am panicking!") },
		"Hope it does not panic!")
	fmt.Println("still no panic?", ok)

	// Can detect panic(nil)
	ok = t.CmpNotPanic(func() { panic(nil) }, "Checks for panic(nil)")
	fmt.Println("last no panic?", ok)

	// Output:
	// checks a panic DID NOT occur: true
	// still no panic? false
	// last no panic? false
}

func TestT(tt *testing.T) {
	t := NewT(tt)
	CmpDeeply(tt, t.Config, DefaultContextConfig)

	t = NewT(tt, ContextConfig{})
	CmpDeeply(tt, t.Config, DefaultContextConfig)

	conf := ContextConfig{
		RootName:  "TEST",
		MaxErrors: 33,
	}
	t = NewT(tt, conf)
	CmpDeeply(tt, t.Config, conf)

	t2 := t.RootName("T2")
	CmpDeeply(tt, t.Config, conf)
	CmpDeeply(tt, t2.Config, ContextConfig{
		RootName:  "T2",
		MaxErrors: 33,
	})

	//
	// Bad usage
	test.CheckPanic(tt,
		func() { NewT(tt, ContextConfig{}, ContextConfig{}) },
		"usage: NewT")
}

func TestRun(tt *testing.T) {
	t := NewT(tt)

	runPassed := false

	ok := t.Run("Test level1",
		func(t *T) {
			ok := t.Run("Test level2",
				func(t *T) {
					runPassed = t.True(true) // test succeeds!
				})

			t.True(ok)
		})

	t.True(ok)
	t.True(runPassed)
}

func TestFailureIsFatal(tt *testing.T) {
	ttt := &TestTestingFT{}

	// All t.True(false) tests of course fail

	// Using default config
	t := NewT(ttt)
	t.True(false) // failure
	CmpNotEmpty(tt, ttt.LastMessage)
	CmpFalse(tt, ttt.IsFatal, "by default it not fatal")

	// Using specific config
	t = NewT(ttt, ContextConfig{FailureIsFatal: true})
	t.True(false) // failure
	CmpNotEmpty(tt, ttt.LastMessage)
	CmpTrue(tt, ttt.IsFatal, "it must be fatal")

	// Using FailureIsFatal()
	t = NewT(ttt).FailureIsFatal()
	t.True(false) // failure
	CmpNotEmpty(tt, ttt.LastMessage)
	CmpTrue(tt, ttt.IsFatal, "it must be fatal")

	// Using FailureIsFatal(true)
	t = NewT(ttt).FailureIsFatal(true)
	t.True(false) // failure
	CmpNotEmpty(tt, ttt.LastMessage)
	CmpTrue(tt, ttt.IsFatal, "it must be fatal")

	// Canceling specific config
	t = NewT(ttt, ContextConfig{FailureIsFatal: false}).FailureIsFatal(false)
	t.True(false) // failure
	CmpNotEmpty(tt, ttt.LastMessage)
	CmpFalse(tt, ttt.IsFatal, "it must be not fatal")
}
