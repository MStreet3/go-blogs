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

var r = rand.New(rand.NewSource(time.Now().Unix()))

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
func (r *eventuallyFatal) Read() (*Message, error) {
	if r.err != nil { // HL
		return nil, r.err // HL
	}

	if rand.Intn(4) == 3 {
		r.err = ErrFatalSocketError
		return nil, r.err
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

// STARTMONITOR OMIT
// monitor is a routine with the single responsibility of closing its returned // HL
// channel if it gets a true value from the function isUnhealthy. // HL
func monitor(
	stop <-chan struct{}, errs <-chan error, isUnhealthy func(error) bool,
) <-chan struct{} {

	done := make(chan struct{})

	go func() {
		defer close(done)
		defer log.Println("monitor: shutting down")

		for {
			select {
			case <-stop:
				return
			case e, ok := <-errs: // <1> // HL
				if !ok {
					return
				}

				if isUnhealthy(e) { // <2> // HL
					log.Printf("monitor: ward is unhealthy; received error %v\n", e)
					return
				}
			}
		}
	}()

	return done
}

// STOPMONITOR OMIT

// STARTSTEWARD OMIT
func connectionSteward(
	stop <-chan struct{}, network ConnectCloser, pulseInterval time.Duration,
) (<-chan struct{}, <-chan error) {

	// Define channels that other clients may consume. <1> // HL
	done := make(chan struct{})
	errs := make(chan error, 1)

	// Define a cleanup function that closes the owned channels. // HL
	cleanup := func() {
		close(errs)
		close(done)
	}

	// Define a function to send errors in a non-blocking fashion. // HL
	sendErr := func(e error) {
		select {
		case <-stop:
			return
		case errs <- e:
		default:
			log.Println("steward: no error listeners")
		}
	}

	// isUnhealthy contains the business logic for when to trigger a ward // HL
	// restart. // HL
	isUnhealthy := func(err error) bool {
		if err != nil {
			if errors.Is(err, ErrFatalSocketError) { // <2> // HL
				return true
			}
			return false
		}
		return false
	}

	go func() {
		defer cleanup()

		for {
			select {
			case <-stop:
				return
			default:
				// Attempt to connect to the network.
				conn, err := network.Connect() // <3> // HL
				if err != nil {
					log.Printf("steward: got error %v while connecting", err)
					sendErr(err)

					// Wait for pulseInterval duration of time
					// to pass before retrying to connect.
					<-time.After(pulseInterval)
					continue
				}

				// Start a new ward to read from the connection.
				log.Println("steward: starting ward")
				stopWard := make(chan struct{})
				reading, readerErrs := readerWard(stopWard, conn, // <4> // HL
					pulseInterval/2)

				// Monitor the ward's health.
				log.Println("steward: monitoring ward")
				restart := monitor(stopWard, readerErrs, isUnhealthy) // <5> // HL

				// Wait for the signal to restart or to stop
				// completely.
				select { // <6> // HL
				case <-stop:
					log.Println("steward: received shutdown signal; stopping ward")
				case <-restart:
					log.Println("steward: stopping unhealthy ward")
				}

				// Cleanup the ward and connection. // <7> // HL
				close(stopWard)
				<-reading
				network.Close()
			}
		}
	}()

	return done, errs
}

// STOPSTEWARD OMIT

// STARTMAIN OMIT
func main() {
	stop := make(chan struct{})
	network := &eventuallyFatalConnection{}
	pulseInterval := 300 * time.Millisecond

	done, _ := connectionSteward(stop, network, pulseInterval) // HL

	time.AfterFunc(5*time.Second, func() {
		close(stop)
	})

	<-done
}

// STOPMAIN OMIT
