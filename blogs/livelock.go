//go:build ignore && OMIT
// +build ignore,OMIT

package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

var ErrFatalSocketError = errors.New("fatal socket error")

type ConnectCloser interface {
	Connect() (Reader, error)
	Close() error
}

// STARTREADER OMIT
type Reader interface {
	Read() (*Message, error)
}

// STOPREADER OMIT

type Message struct {
	Content string
}

// STARTEVENTUALLYFATALREADER OMIT
type eventuallyFatal struct {
	err error
}

// eventuallyFatal will eventally be stuck in a state of only returning // HL
// ErrFatalSocketError. // HL
func (eveFatal *eventuallyFatal) Read() (*Message, error) {
	if eveFatal.err != nil { // HL
		return nil, eveFatal.err // HL
	}

	if rand.Intn(4) == 3 {
		eveFatal.err = ErrFatalSocketError
		return nil, eveFatal.err
	}

	return &Message{
		Content: fmt.Sprintf("%d", rand.Int()),
	}, nil
}

// STOPEVENTUALLYFATALREADER OMIT

type eventuallyFatalConnection struct{}

func (conn *eventuallyFatalConnection) Connect() (Reader, error) {
	log.Println("conn: connected successfully")
	return &eventuallyFatal{}, nil
}

func (conn *eventuallyFatalConnection) Close() error {
	log.Println("conn: disconnected successfully")
	return nil
}

// STARTREADERWARD OMIT
func readerWard(
	stop <-chan struct{}, conn Reader, pulseInterval time.Duration,
) (<-chan struct{}, <-chan error) {

	done := make(chan struct{}) // <1> // HL
	errs := make(chan error, 1) // <2> // HL
	ticker := time.NewTicker(pulseInterval)

	cleanup := func() {
		ticker.Stop()
		close(errs)
		close(done)
	}

	sendErr := func(e error) { // <5> // HL
		select {
		case <-stop:
			return
		case errs <- e:
		default:
			log.Println("ward: no error listeners")
		}
	}

	go func() {
		defer cleanup() // <3> // HL

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				msg, err := conn.Read() // <4> // HL

				if err != nil {
					sendErr(err)
					continue
				}

				log.Printf("ward: read message %s", msg.Content)
			}
		}
	}()

	return done, errs // <6> // HL
}

// STOPREADERWARD OMIT

// STARTMAINLIVELOCK OMIT
func main() {
	stop := make(chan struct{})
	reader, _ := new(eventuallyFatalConnection).Connect()
	pulseInterval := 300 * time.Millisecond

	done, errs := readerWard(stop, reader, pulseInterval) // <1> // HL

	go func() {
		for e := range errs {
			log.Printf("main: read error %v\n", e) // <2> // HL
		}
	}()

	time.AfterFunc(3*time.Second, func() {
		close(stop)
	})

	<-done // <3> // HL
}

// STOPMAINLIVELOCK OMIT
