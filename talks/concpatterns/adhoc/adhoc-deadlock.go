//go:build ignore && OMIT
// +build ignore,OMIT

package main

import "fmt"

// forwarder reads values from src channel and writes them on to the sink
// channel.  forwarder closes the sink channel once the src channel is empty.
func forwarder(sink chan<- int, src <-chan int) {
	defer close(sink)
	defer fmt.Println("done forwarding")

	for val := range src {
		sink <- val
	}
}

// consumer reads values from the src channel and writes them to the standard
// output.  consumer reads indefinitely until the src channel is closed.
func consumer(src <-chan int) {
	defer fmt.Println("done printing")

	for num := range src {
		fmt.Println(num)
	}
}

func main() {
	// STARTMAIN1 OMIT
	var (
		data   = []int{1, 2, 3, 4}
		source = make(chan int, len(data)) // HL
		sink   = make(chan int)
	)

	// Load the source channel with data.
	for _, val := range data {
		source <- val
	}

	// Launch a goroutine to forward the data from the source channel and
	// onto the sink channel.  Adhoc convention dictates that only the
	// forwarder function closes the sink channel.
	go forwarder(sink, source) // HL

	// Read all the values from the sink channel.  The reader function
	// requires the adhoc convention that the input channel is eventually
	// closed otherwise it would block forever.
	consumer(sink) // HL

	// STOPMAIN1 OMIT
}
