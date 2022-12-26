//go:build ignore && OMIT
// +build ignore,OMIT

package main

import (
	"fmt"
	"sync"
)

// STARTMAIN1 OMIT
func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup

	var count int // <1> // HL

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			mu.Lock()         // <2> // HL
			defer mu.Unlock() // <3> // HL
			count++           // <4> // HL
		}()
	}

	wg.Wait()
	fmt.Printf("Got: %d\n", count)
}

// STOPMAIN1 OMIT
