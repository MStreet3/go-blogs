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

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			sema <- struct{}{}
			count++
			<-sema
		}()
	}

	wg.Wait()

	fmt.Printf("Got: %d\n", count)
}
