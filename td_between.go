// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package testdeep

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/maxatome/go-testdeep/internal/ctxerr"
	"github.com/maxatome/go-testdeep/internal/types"
	"github.com/maxatome/go-testdeep/internal/util"
)

type boundCmp uint8

const (
	boundNone boundCmp = iota
	boundIn
	boundOut
)

type tdBetween struct {
	Base
	expectedMin reflect.Value
	expectedMax reflect.Value

	minBound boundCmp
	maxBound boundCmp
}

var _ TestDeep = &tdBetween{}

// BoundsKind type qualifies the "Between" bounds.
type BoundsKind uint8

const (
	// BoundsInIn allows to match between "from" and "to" both included
	BoundsInIn BoundsKind = iota
	// BoundsInOut allows to match between "from" included and "to" excluded
	BoundsInOut
	// BoundsOutIn allows to match between "from" excluded and "to" included
	BoundsOutIn
	// BoundsOutOut allows to match between "from" and "to" both excluded
	BoundsOutOut
)

type tdBetweenTime struct {
	tdBetween
	expectedType reflect.Type
	mustConvert  bool
}

var _ TestDeep = &tdBetweenTime{}

// Between operator checks that data is between "from" and
// "to". "from" and "to" can be any numeric, string or time.Time (or
// assignable) value. "from" and "to" must be the same kind as the
// compared value if numeric, and the same type if string or time.Time (or
// assignable). "bounds" allows to specify whether bounds are included
// or not. See Bounds* constants for details. If "bounds" is missing,
// it defaults to BoundsInIn.
//
// TypeBehind method returns the reflect.Type of "from" (same as the "to" one.)
func Between(from interface{}, to interface{}, bounds ...BoundsKind) TestDeep {
	b := tdBetween{
		expectedMin: reflect.ValueOf(from),
		expectedMax: reflect.ValueOf(to),
	}

	const usage = "usage: Between(NUM|STRING|TIME, NUM|STRING|TIME[, BOUNDS_KIND])"

	if len(bounds) > 0 {
		if len(bounds) > 1 {
			panic(usage)
		}

		if bounds[0] == BoundsInIn || bounds[0] == BoundsInOut {
			b.minBound = boundIn
		} else {
			b.minBound = boundOut
		}

		if bounds[0] == BoundsInIn || bounds[0] == BoundsOutIn {
			b.maxBound = boundIn
		} else {
			b.maxBound = boundOut
		}
	} else {
		b.minBound = boundIn
		b.maxBound = boundIn
	}

	if b.expectedMax.Type() != b.expectedMin.Type() {
		panic("from and to params must have the same type")
	}

	return b.initBetween(usage)
}

func (b *tdBetween) initBetween(usage string) TestDeep {
	b.Base = NewBase(4)

	if !b.expectedMax.IsValid() {
		b.expectedMax = b.expectedMin
	}

	switch b.expectedMin.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if b.expectedMin.Int() > b.expectedMax.Int() {
			b.expectedMin, b.expectedMax = b.expectedMax, b.expectedMin
		}
		return b

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if b.expectedMin.Uint() > b.expectedMax.Uint() {
			b.expectedMin, b.expectedMax = b.expectedMax, b.expectedMin
		}
		return b

	case reflect.Float32, reflect.Float64:
		if b.expectedMin.Float() > b.expectedMax.Float() {
			b.expectedMin, b.expectedMax = b.expectedMax, b.expectedMin
		}
		return b

	case reflect.String:
		if b.expectedMin.String() > b.expectedMax.String() {
			b.expectedMin, b.expectedMax = b.expectedMax, b.expectedMin
		}
		return b

	case reflect.Struct:
		var bt tdBetweenTime
		if b.expectedMin.Type() == timeType {
			bt = tdBetweenTime{
				tdBetween:    *b,
				expectedType: timeType,
			}
		} else if b.expectedMin.Type().ConvertibleTo(timeType) {
			bt = tdBetweenTime{
				tdBetween:    *b,
				expectedType: b.expectedMin.Type(),
				mustConvert:  true,
			}
			bt.expectedMin = b.expectedMin.Convert(timeType)
			bt.expectedMax = b.expectedMax.Convert(timeType)
		}

		if bt.expectedMin.Interface().(time.Time).After(
			bt.expectedMax.Interface().(time.Time)) {
			bt.expectedMin, bt.expectedMax = bt.expectedMax, bt.expectedMin
		}

		return &bt
	}
	panic(usage)
}

func (b *tdBetween) nInt(tolerance reflect.Value) {
	if diff := tolerance.Int(); diff != 0 {
		base := b.expectedMin.Int()

		max := base + diff
		if max < base {
			max = math.MaxInt64
		}

		min := base - diff
		if min > base {
			min = math.MinInt64
		}

		b.expectedMin = reflect.New(tolerance.Type()).Elem()
		b.expectedMin.SetInt(min)

		b.expectedMax = reflect.New(tolerance.Type()).Elem()
		b.expectedMax.SetInt(max)
	}
}

func (b *tdBetween) nUint(tolerance reflect.Value) {
	if diff := tolerance.Uint(); diff != 0 {
		base := b.expectedMin.Uint()

		max := base + diff
		if max < base {
			max = math.MaxUint64
		}

		min := base - diff
		if min > base {
			min = 0
		}

		b.expectedMin = reflect.New(tolerance.Type()).Elem()
		b.expectedMin.SetUint(min)

		b.expectedMax = reflect.New(tolerance.Type()).Elem()
		b.expectedMax.SetUint(max)
	}
}

func (b *tdBetween) nFloat(tolerance reflect.Value) {
	if diff := tolerance.Float(); diff != 0 {
		base := b.expectedMin.Float()

		b.expectedMin = reflect.New(tolerance.Type()).Elem()
		b.expectedMin.SetFloat(base - diff)

		b.expectedMax = reflect.New(tolerance.Type()).Elem()
		b.expectedMax.SetFloat(base + diff)
	}
}

// N operator compares a numeric data against "num" ± "tolerance". If
// "tolerance" is missing, it defaults to 0. "num" and "tolerance"
// must be the same kind as the compared value.
//
// TypeBehind method returns the reflect.Type of "num".
func N(num interface{}, tolerance ...interface{}) TestDeep {
	n := tdBetween{
		Base:        NewBase(3),
		expectedMin: reflect.ValueOf(num),
		minBound:    boundIn,
		maxBound:    boundIn,
	}

	const usage = "usage: N({,U}INT{,8,16,32,64}|FLOAT{32,64}[, TOLERANCE])"

	switch n.expectedMin.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
	default:
		panic(usage)
	}

	n.expectedMax = n.expectedMin

	if len(tolerance) > 0 {
		if len(tolerance) > 1 {
			panic(usage)
		}

		tol := reflect.ValueOf(tolerance[0])
		if tol.Type() != n.expectedMin.Type() {
			panic("tolerance param must have the same type as num one")
		}

		switch tol.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n.nInt(tol)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
			reflect.Uint64:
			n.nUint(tol)

		default: // case reflect.Float32, reflect.Float64:
			n.nFloat(tol)
		}
	}

	return &n
}

// Gt operator checks that data is greater than "val". "val" can be
// any numeric or time.Time (or assignable) value. "val" must be the
// same kind as the compared value if numeric, and the same type if
// time.Time (or assignable).
//
// TypeBehind method returns the reflect.Type of "val".
func Gt(val interface{}) TestDeep {
	b := &tdBetween{
		expectedMin: reflect.ValueOf(val),
		minBound:    boundOut,
	}
	return b.initBetween("usage: Gt(NUM|STRING|TIME)")
}

// Gte operator checks that data is greater or equal than "val". "val"
// can be any numeric or time.Time (or assignable) value. "val" must
// be the same kind as the compared value if numeric, and the same
// type if time.Time (or assignable).
//
// TypeBehind method returns the reflect.Type of "val".
func Gte(val interface{}) TestDeep {
	b := &tdBetween{
		expectedMin: reflect.ValueOf(val),
		minBound:    boundIn,
	}
	return b.initBetween("usage: Gte(NUM|STRING|TIME)")
}

// Lt operator checks that data is lesser than "val". "val" can be
// any numeric or time.Time (or assignable) value. "val" must be the
// same kind as the compared value if numeric, and the same type if
// time.Time (or assignable).
//
// TypeBehind method returns the reflect.Type of "val".
func Lt(val interface{}) TestDeep {
	b := &tdBetween{
		expectedMin: reflect.ValueOf(val),
		maxBound:    boundOut,
	}
	return b.initBetween("usage: Lt(NUM|STRING|TIME)")
}

// Lte operator checks that data is lesser or equal than "val". "val"
// can be any numeric or time.Time (or assignable) value. "val" must
// be the same kind as the compared value if numeric, and the same
// type if time.Time (or assignable).
//
// TypeBehind method returns the reflect.Type of "val".
func Lte(val interface{}) TestDeep {
	b := &tdBetween{
		expectedMin: reflect.ValueOf(val),
		maxBound:    boundIn,
	}
	return b.initBetween("usage: Lte(NUM|STRING|TIME)")
}

func (b *tdBetween) matchInt(got reflect.Value) (ok bool) {
	switch b.minBound {
	case boundIn:
		ok = got.Int() >= b.expectedMin.Int()
	case boundOut:
		ok = got.Int() > b.expectedMin.Int()
	default:
		ok = true
	}
	if ok {
		switch b.maxBound {
		case boundIn:
			ok = got.Int() <= b.expectedMax.Int()
		case boundOut:
			ok = got.Int() < b.expectedMax.Int()
		default:
			ok = true
		}
	}
	return
}

func (b *tdBetween) matchUint(got reflect.Value) (ok bool) {
	switch b.minBound {
	case boundIn:
		ok = got.Uint() >= b.expectedMin.Uint()
	case boundOut:
		ok = got.Uint() > b.expectedMin.Uint()
	default:
		ok = true
	}
	if ok {
		switch b.maxBound {
		case boundIn:
			ok = got.Uint() <= b.expectedMax.Uint()
		case boundOut:
			ok = got.Uint() < b.expectedMax.Uint()
		default:
			ok = true
		}
	}
	return
}

func (b *tdBetween) matchFloat(got reflect.Value) (ok bool) {
	switch b.minBound {
	case boundIn:
		ok = got.Float() >= b.expectedMin.Float()
	case boundOut:
		ok = got.Float() > b.expectedMin.Float()
	default:
		ok = true
	}
	if ok {
		switch b.maxBound {
		case boundIn:
			ok = got.Float() <= b.expectedMax.Float()
		case boundOut:
			ok = got.Float() < b.expectedMax.Float()
		default:
			ok = true
		}
	}
	return
}

func (b *tdBetween) matchString(got reflect.Value) (ok bool) {
	switch b.minBound {
	case boundIn:
		ok = got.String() >= b.expectedMin.String()
	case boundOut:
		ok = got.String() > b.expectedMin.String()
	default:
		ok = true
	}
	if ok {
		switch b.maxBound {
		case boundIn:
			ok = got.String() <= b.expectedMax.String()
		case boundOut:
			ok = got.String() < b.expectedMax.String()
		default:
			ok = true
		}
	}
	return
}

func (b *tdBetween) Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error {
	if got.Type() != b.expectedMin.Type() {
		if ctx.BooleanError {
			return ctxerr.BooleanError
		}
		return ctx.CollectError(&ctxerr.Error{
			Message:  "type mismatch",
			Got:      types.RawString(got.Type().String()),
			Expected: types.RawString(b.expectedMin.Type().String()),
		})
	}

	var ok bool

	switch got.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ok = b.matchInt(got)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ok = b.matchUint(got)

	case reflect.Float32, reflect.Float64:
		ok = b.matchFloat(got)

	case reflect.String:
		ok = b.matchString(got)
	}

	if ok {
		return nil
	}

	if ctx.BooleanError {
		return ctxerr.BooleanError
	}

	var gotStr types.RawString
	if got.Kind() == reflect.String {
		gotStr = types.RawString(util.ToString(got))
	} else {
		gotStr = types.RawString(fmt.Sprintf("%v", got))
	}

	return ctx.CollectError(&ctxerr.Error{
		Message:  "values differ",
		Got:      gotStr,
		Expected: types.RawString(b.String()),
	})
}

func (b *tdBetween) String() string {
	var (
		min, max       interface{}
		minStr, maxStr string
	)

	if b.minBound != boundNone {
		min = b.expectedMin.Interface()

		// We want strings be double-quoted
		if b.expectedMin.Kind() == reflect.String {
			minStr = util.ToString(min)
		} else {
			minStr = fmt.Sprintf("%v", min)
		}
	}
	if b.maxBound != boundNone {
		max = b.expectedMax.Interface()

		// We want strings be double-quoted
		if b.expectedMax.Kind() == reflect.String {
			maxStr = util.ToString(max)
		} else {
			maxStr = fmt.Sprintf("%v", max)
		}
	}

	if min == max {
		return minStr
	}

	if min != nil {
		if max != nil {
			return fmt.Sprintf("%v %c got %c %v",
				minStr,
				util.TernRune(b.minBound == boundIn, '≤', '<'),
				util.TernRune(b.maxBound == boundIn, '≤', '<'),
				maxStr)
		}

		return fmt.Sprintf("%c %v",
			util.TernRune(b.minBound == boundIn, '≥', '>'), minStr)
	}

	return fmt.Sprintf("%c %v",
		util.TernRune(b.maxBound == boundIn, '≤', '<'), maxStr)
}

func (b *tdBetween) TypeBehind() reflect.Type {
	return b.expectedMin.Type()
}

var _ TestDeep = &tdBetweenTime{}

func (b *tdBetweenTime) Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error {
	if got.Type() != b.expectedType {
		if ctx.BooleanError {
			return ctxerr.BooleanError
		}
		return ctx.CollectError(&ctxerr.Error{
			Message:  "type mismatch",
			Got:      types.RawString(got.Type().String()),
			Expected: types.RawString(b.expectedType.String()),
		})
	}

	cmpGot, err := getTime(ctx, got, b.mustConvert)
	if err != nil {
		return ctx.CollectError(err)
	}

	var ok bool
	if b.minBound != boundNone {
		min := b.expectedMin.Interface().(time.Time)

		if b.minBound == boundIn {
			ok = !min.After(cmpGot)
		} else {
			ok = cmpGot.After(min)
		}
	} else {
		ok = true
	}

	if ok && b.maxBound != boundNone {
		max := b.expectedMax.Interface().(time.Time)

		if b.maxBound == boundIn {
			ok = !max.Before(cmpGot)
		} else {
			ok = cmpGot.Before(max)
		}
	}

	if ok {
		return nil
	}

	if ctx.BooleanError {
		return ctxerr.BooleanError
	}
	return ctx.CollectError(&ctxerr.Error{
		Message:  "values differ",
		Got:      types.RawString(fmt.Sprintf("%v", got.Interface())),
		Expected: types.RawString(b.String()),
	})
}

func (b *tdBetweenTime) TypeBehind() reflect.Type {
	return b.expectedType
}
