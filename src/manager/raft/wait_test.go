package raft

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWait(t *testing.T) {
	testWait := newWait()

	ch1 := testWait.register(uint64(1), func() { fmt.Println("test1") }, func() { fmt.Println("cancel1") })
	ch2 := testWait.register(uint64(2), func() { fmt.Println("test2") }, func() { fmt.Println("cancel2") })
	ch3 := testWait.register(uint64(3), func() { fmt.Println("test3") }, func() { fmt.Println("cancel3") })
	ch4 := testWait.register(uint64(4), func() { fmt.Println("test4") }, func() { fmt.Println("cancel4") })
	assert.NotNil(t, ch1)
	assert.NotNil(t, ch2)
	assert.NotNil(t, ch3)
	assert.NotNil(t, ch4)

	trigger1 := testWait.trigger(uint64(1), "trigger1")
	assert.Equal(t, trigger1, true)
	trigger5 := testWait.trigger(uint64(5), "triggr5")
	assert.Equal(t, trigger5, false)

	ch1Value := (<-ch1).(string)
	assert.Equal(t, ch1Value, "trigger1")

	testWait.cancel(uint64(2))

	testWait.cancelAll()
}

func TestDuplicateRegister(t *testing.T) {
	testWait := newWait()

	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()

	ch1 := testWait.register(uint64(1), func() { fmt.Println("test1") }, func() { fmt.Println("cancel1") })
	ch2 := testWait.register(uint64(1), func() { fmt.Println("test1") }, func() { fmt.Println("cancel1") })
	assert.NotNil(t, ch1)
	assert.NotNil(t, ch2)
}
