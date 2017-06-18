package assert

import (
	"testing"
	"time"
)

type MyStruct struct {
	Sub *MyStruct
}

func TestEqual(t *testing.T) {
	Equal(t, "foo", "foo")
	Equal(t, true, true)

	myStructA := MyStruct{}
	myStructB := MyStruct{}
	Equal(t, myStructA, myStructB)

	// Equal(t, "foo", "bar", "this should blow up")
}

func TestNotEqual(t *testing.T) {
	NotEqual(t, "foo", "bar", "msg!")
	NotEqual(t, nil, false)

	myStructA := MyStruct{}
	myStructB := MyStruct{&myStructA}
	NotEqual(t, myStructA, myStructB)
	NotEqual(t, &myStructA, myStructA)

	// NotEqual(t, "foo", "foo", "this should blow up")
}

func TestTrue(t *testing.T) {
	True(t, true)
}

func TestFalse(t *testing.T) {
	False(t, false)
}

func TestNil(t *testing.T) {
	Nil(t, nil)

	var nilChan chan int
	Nil(t, nilChan)

	var nilFunc func(int) int
	Nil(t, nilFunc)

	var nilInterface interface{}
	Nil(t, nilInterface)

	var nilMap map[string]string
	Nil(t, nilMap)

	var myStruct MyStruct
	Nil(t, myStruct.Sub) // nil pointer

	var nilSlice []string
	Nil(t, nilSlice)

	// Nil(t, "foo", "this should blow up")
}

func TestNotNil(t *testing.T) {
	NotNil(t, "foo")

	myStruct := MyStruct{}
	NotNil(t, myStruct)
	NotNil(t, &myStruct)

	// NotNil(t, nil, "this should blow up")
	// var myNilStruct MyStruct
	// NotNil(t, myNilStruct, "this should blow up")
}

func TestContains(t *testing.T) {
	Contains(t, "foo", "bizmarfooba")
	Contains(t, "", "bizmarfooba")

	// Contains(t, "cool", "", "This should blow up")
}

func TestNotContains(t *testing.T) {
	NotContains(t, "a", "")
	NotContains(t, "Lorem", "lorem")

	// NotContains(t, "c", "abc", "This should blow up")
}

func TestWithinDuration(t *testing.T) {
	now := time.Now()
	WithinDuration(t, time.Millisecond, now, now)
	WithinDuration(t, time.Second, now, now.Add(time.Second))
	WithinDuration(t, time.Second, now, now.Add(-time.Second))
	WithinDuration(t, time.Second, now, now.Add(999*time.Millisecond))
	WithinDuration(t, time.Second, now, now.Add(-999*time.Millisecond))

	// WithinDuration(t, time.Millisecond, now, now.Add(time.Second), "This should blow up")
	// WithinDuration(t, time.Millisecond, now, now.Add(-time.Second), "This should blow up")
}

func TestPanic(t *testing.T) {
	// Simple string
	func() {
		defer Panic(t, "Swerve wildly!")
		panic("Swerve wildly!")
	}()
	
	// Non-string
	func() {
		defer Panic(t, 123)
		panic(123)
	}()
	
	// Everything's cool.
	func() {
		defer Panic(t, nil)
	}()
}
