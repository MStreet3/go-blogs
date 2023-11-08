//go:build ignore && OMIT
// +build ignore,OMIT

package main

import "fmt"

type Greeter interface {
	Greet() string
}

type adapter struct {
	u        unadapted
	greeting string
}

func (a adapter) Greet() string {
	return a.u.Greet(a.greeting)
}

type unadapted struct{}

func (u unadapted) Greet(greeting string) string {
	return fmt.Sprintf("Hello %s", greeting)
}
