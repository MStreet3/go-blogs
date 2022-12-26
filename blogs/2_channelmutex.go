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

	mu := NewMutex(0)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		// STARTBAD OMIT
		go func() {
			defer wg.Done()

			val := mu.Lock()
			val++
			mu.Unlock(val) // HL

			// Modify val after unlock is a data race, however,
			// Go compiler has no issues compiling this code!
			val++ // HL
		}()
		// STOPBAD OMIT
	}

	wg.Wait()

	count := mu.Lock()
	fmt.Printf("Got: %d\n", count)
}

// STARTMAIN1 OMIT
func main() {
	var wg sync.WaitGroup

	mu := NewMutex(0) // <1> // HL

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			val := mu.Lock() // <2> // HL
			val++            // <3> // HL
			mu.Unlock(val)   // <4> // HL
		}()
	}

	wg.Wait()

	count := mu.Lock() // <5> // HL
	fmt.Printf("Got: %d\n", count)
}

// STOPMAIN1 OMIT
