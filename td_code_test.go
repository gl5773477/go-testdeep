// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep"
	"github.com/maxatome/go-testdeep/internal/test"
)

func TestCode(t *testing.T) {
	checkOK(t, 12, testdeep.Code(func(n int) bool { return n >= 10 && n < 20 }))

	checkOK(t, 12, testdeep.Code(func(val interface{}) bool {
		num, ok := val.(int)
		return ok && num == 12
	}))

	checkOK(t, errors.New("foobar"), testdeep.Code(func(val error) bool {
		return val.Error() == "foobar"
	}))

	checkOK(t, json.RawMessage(`[42]`),
		testdeep.Code(func(b json.RawMessage) error {
			var l []int
			err := json.Unmarshal(b, &l)
			if err != nil {
				return err
			}
			if len(l) != 1 || l[0] != 42 {
				return errors.New("42 not found")
			}
			return nil
		}))

	checkError(t, 123, testdeep.Code(func(n float64) bool { return true }),
		expectedError{
			Message:  mustBe("incompatible parameter type"),
			Path:     mustBe("DATA"),
			Got:      mustBe("int"),
			Expected: mustBe("float64"),
		})

	type xInt int
	checkError(t, xInt(12),
		testdeep.Code(func(n int) bool { return n >= 10 && n < 20 }),
		expectedError{
			Message:  mustBe("incompatible parameter type"),
			Path:     mustBe("DATA"),
			Got:      mustBe("testdeep_test.xInt"),
			Expected: mustBe("int"),
		})

	checkError(t, 12,
		testdeep.Code(func(n int) (bool, string) { return false, "custom error" }),
		expectedError{
			Message: mustBe("ran code with %% as argument"),
			Path:    mustBe("DATA"),
			Summary: mustBe("        value: 12\nit failed coz: custom error"),
		})

	checkError(t, 12,
		testdeep.Code(func(n int) bool { return false }),
		expectedError{
			Message: mustBe("ran code with %% as argument"),
			Path:    mustBe("DATA"),
			Summary: mustBe("  value: 12\nit failed but didn't say why"),
		})

	type MyBool bool
	type MyString string
	checkError(t, 12,
		testdeep.Code(func(n int) (MyBool, MyString) { return false, "very custom error" }),
		expectedError{
			Message: mustBe("ran code with %% as argument"),
			Path:    mustBe("DATA"),
			Summary: mustBe("        value: 12\nit failed coz: very custom error"),
		})

	checkError(t, 12,
		testdeep.Code(func(i int) error {
			return errors.New("very custom error")
		}),
		expectedError{
			Message: mustBe("ran code with %% as argument"),
			Path:    mustBe("DATA"),
			Summary: mustBe("        value: 12\nit failed coz: very custom error"),
		})

	//
	// Bad usage
	test.CheckPanic(t, func() { testdeep.Code("test") }, "usage: Code")

	test.CheckPanic(t, func() {
		testdeep.Code(func() bool { return true })
	}, "FUNC must take only one argument")

	test.CheckPanic(t, func() {
		testdeep.Code(func(a int, b string) bool { return true })
	}, "FUNC must take only one argument")

	test.CheckPanic(t, func() {
		testdeep.Code(func(n int) (bool, int) { return true, 0 })
	}, "FUNC must return bool or (bool, string) or error")

	test.CheckPanic(t, func() {
		testdeep.Code(func(n int) (error, string) { return nil, "" }) // nolint: staticcheck
	}, "FUNC must return bool or (bool, string) or error")

	test.CheckPanic(t, func() {
		testdeep.Code(func(n int) (int, string) { return 0, "" })
	}, "FUNC must return bool or (bool, string) or error")

	test.CheckPanic(t, func() {
		testdeep.Code(func(n int) (string, bool) { return "", true })
	}, "FUNC must return bool or (bool, string) or error")

	test.CheckPanic(t, func() {
		testdeep.Code(func(n int) (bool, string, int) { return true, "", 0 })
	}, "FUNC must return bool or (bool, string) or error")

	test.CheckPanic(t, func() {
		testdeep.Code(func(n int) {})
	}, "FUNC must return bool or (bool, string) or error")

	test.CheckPanic(t, func() {
		testdeep.Code(func(n int) int { return 0 })
	}, "FUNC must return bool or (bool, string) or error")

	//
	// String
	test.EqualStr(t,
		testdeep.Code(func(n int) bool { return false }).String(),
		"Code(func(int) bool)")
	test.EqualStr(t,
		testdeep.Code(func(n int) (bool, string) { return false, "" }).String(),
		"Code(func(int) (bool, string))")
	test.EqualStr(t,
		testdeep.Code(func(n int) error { return nil }).String(),
		"Code(func(int) error)")
	test.EqualStr(t,
		testdeep.Code(func(n int) (MyBool, MyString) { return false, "" }).String(),
		"Code(func(int) (testdeep_test.MyBool, testdeep_test.MyString))")
}

func TestCodeTypeBehind(t *testing.T) {
	// Type behind is the code function parameter one

	equalTypes(t, testdeep.Code(func(n int) bool { return n != 0 }), 23)

	type MyTime time.Time

	equalTypes(t,
		testdeep.Code(func(t MyTime) bool { return time.Time(t).IsZero() }),
		MyTime{})
}
