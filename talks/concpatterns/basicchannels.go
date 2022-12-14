//go:build ignore && OMIT
// +build ignore,OMIT

package main

func f() {}

func g(a, b int) {}

func main() {
	// STARTMAIN1  OMIT
	go f()
	go g(1, 2)
	// STOPMAIN1 OMIT

	// STARTMAIN2  OMIT
	c := make(chan int)
	go func() { c <- 3 }()
	n := <-c
	// STOPMAIN2 OMIT
	_ = n
}
