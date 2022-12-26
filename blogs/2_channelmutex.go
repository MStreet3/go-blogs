//go:build ignore && OMIT
// +build ignore,OMIT

package main

import (
	"fmt"
	"sync"
)

// STARTMUTEX OMIT
type mutex[T any] struct {
	c chan T
}

func (m mutex[T]) Lock() T {
	return <-m.c
}

func (m mutex[T]) Unlock(val T) {
	m.c <- val
}

// STOPMUTEX OMIT

// STARTCONSTRUCTOR OMIT
// NewMutex instantiates a buffered channel with capacity one, places a value into // HL
// the buffer and uses this channel to construct a new mutex type. // HL
func NewMutex[T any](val T) mutex[T] {
	c := make(chan T, 1)
	c <- val
	return mutex[T]{c}
}

// STOPCONSTRUCTOR OMIT

func bad() {
	var wg sync.WaitGroup
	var count int // <1> // HL

	mu := NewMutex(count) // <2> // HL

	for i := 0; i < 10; i++ {
		wg.Add(1)
		// STARTBAD OMIT
		go func() {
			defer wg.Done()

			count = mu.Lock()
			count++
			mu.Unlock(count) // HL

			// Modify count after unlock is a data race, however,
			// Go compiler has no issues compiling this code!
			count++ // HL
		}()
		// STOPBAD OMIT
	}

	wg.Wait()

	count = mu.Lock() // <6> // HL
	fmt.Printf("Got: %d\n", count)
}

// STARTMAIN1 OMIT
func main() {
	var wg sync.WaitGroup
	var count int // <1> // HL

	mu := NewMutex(count) // <2> // HL

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			count = mu.Lock() // <3> // HL
			count++           // <4> // HL
			mu.Unlock(count)  // <5> // HL
		}()
	}

	wg.Wait()

	count = mu.Lock() // <6> // HL
	fmt.Printf("Got: %d\n", count)
}

// STOPMAIN1 OMIT
