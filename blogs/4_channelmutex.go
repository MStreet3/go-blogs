//go:build ignore && OMIT
// +build ignore,OMIT

package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var count int

	sema := make(chan struct{}, 1)

	done := func() {
		defer wg.Done()
		<-sema
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer done()

			sema <- struct{}{}
			count++
		}()
	}

	wg.Wait()

	fmt.Printf("Got: %d\n", count)
}
