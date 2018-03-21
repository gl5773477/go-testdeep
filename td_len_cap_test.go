package testdeep_test

import (
	"testing"

	. "github.com/maxatome/go-testdeep"
)

func TestLen(t *testing.T) {
	checkOK(t, "abcd", Len(4))
	checkOK(t, "abcd", Len(4, 6))
	checkOK(t, "abcd", Len(6, 4))

	checkOK(t, []byte("abcd"), Len(4))
	checkOK(t, []byte("abcd"), Len(4, 6))

	checkOK(t, [5]int{}, Len(5))
	checkOK(t, [5]int{}, Len(4, 6))

	checkOK(t, map[int]bool{1: true, 2: false}, Len(2))
	checkOK(t, map[int]bool{1: true, 2: false}, Len(1, 6))

	checkOK(t, make(chan int, 3), Len(0))

	checkError(t, [5]int{}, Len(4), expectedError{
		Message:  mustBe("bad length"),
		Path:     mustBe("DATA"),
		Got:      mustBe("5"),
		Expected: mustBe("4"),
	})

	checkError(t, 123, Len(4), expectedError{
		Message:  mustBe("bad type"),
		Path:     mustBe("DATA"),
		Got:      mustBe("int"),
		Expected: mustBe("Array, Chan, Map, Slice or string"),
	})

	//
	// Bad usage
	checkPanic(t, func() { Len(1, 2, 3) }, "usage: Len")

	//
	// String
	equalStr(t, Len(3).String(), "len=3")
	equalStr(t, Len(3, 8).String(), "3 ≤ len ≤ 8")
	equalStr(t, Len(8, 3).String(), "3 ≤ len ≤ 8")
}

func TestCap(t *testing.T) {
	checkOK(t, make([]byte, 0, 4), Cap(4))
	checkOK(t, make([]byte, 0, 4), Cap(4, 6))

	checkOK(t, [5]int{}, Cap(5))
	checkOK(t, [5]int{}, Cap(4, 6))

	checkOK(t, make(chan int, 3), Cap(3))

	checkError(t, [5]int{}, Cap(2, 4), expectedError{
		Message:  mustBe("bad capacity"),
		Path:     mustBe("DATA"),
		Got:      mustBe("5"),
		Expected: mustBe("2 ≤ cap ≤ 4"),
	})

	checkError(t, map[int]int{1: 2}, Cap(1), expectedError{
		Message:  mustBe("bad type"),
		Path:     mustBe("DATA"),
		Got:      mustBe("map[int]int"),
		Expected: mustBe("Array, Chan or Slice"),
	})

	//
	// Bad usage
	checkPanic(t, func() { Cap(1, 2, 3) }, "usage: Cap")

	//
	// String
	equalStr(t, Cap(3).String(), "cap=3")
	equalStr(t, Cap(3, 8).String(), "3 ≤ cap ≤ 8")
	equalStr(t, Cap(8, 3).String(), "3 ≤ cap ≤ 8")
}
