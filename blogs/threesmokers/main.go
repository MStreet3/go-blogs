//go:build ignore && OMIT
// +build ignore,OMIT

package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

type Component int

const (
	TobaccoComponent Component = iota + 1
	PapersComponent
	LighterComponent
)

func (c Component) String() string {
	switch c {
	case TobaccoComponent:
		return "tobacco"
	case PapersComponent:
		return "papers"
	case LighterComponent:
		return "lighter"
	default:
		return "unknown"
	}
}

type Components []Component

func (arr Components) String() string {
	var str []string
	for _, c := range arr {
		str = append(str, c.String())
	}

	front := str[:len(str)-1]

	return fmt.Sprintf("%s and %s", strings.Join(front, ", "), str[len(str)-1])
}

type Tobacco struct{}
type Papers struct{}
type Lighter struct{}

type AvailableComponents struct {
	Tobacco <-chan Tobacco
	Papers  <-chan Papers
	Lighter <-chan Lighter
}

type SmokerWithTobacco struct {
	Papers  <-chan Papers
	Lighter <-chan Lighter
}

type SmokerWithPapers struct {
	Tobacco <-chan Tobacco
	Lighter <-chan Lighter
}

type SmokerWithLighter struct {
	Tobacco <-chan Tobacco
	Papers  <-chan Papers
}

func getComponents() Components {
	var (
		forSWT = Components{
			LighterComponent,
			PapersComponent,
		}

		forSWP = Components{
			LighterComponent,
			TobaccoComponent,
		}

		forSWL = Components{
			TobaccoComponent,
			PapersComponent,
		}

		arr = []Components{
			forSWT, forSWL, forSWP,
		}
	)

	rand.Seed(time.Now().Unix())

	index := rand.Intn(3)

	return arr[index]
}

// STARTAGENT OMIT
func agent(
	stop <-chan struct{}, signal <-chan struct{},
) (<-chan struct{}, AvailableComponents) {

	var (
		done      = make(chan struct{})
		tobaccoCh = make(chan Tobacco)
		papersCh  = make(chan Papers)
		lighterCh = make(chan Lighter)

		ac = AvailableComponents{
			Tobacco: tobaccoCh,
			Papers:  papersCh,
			Lighter: lighterCh,
		} // <1> // HL
	)

	cleanup := func() { // <2> // HL
		defer close(done)

		close(tobaccoCh)
		close(papersCh)
		close(lighterCh)
		log.Println("agent: shutdown complete")
	}

	go func() {
		defer cleanup()

		for {
			select {
			case <-stop:
				return
			case <-signal:
				components := getComponents() // <3> // HL

				for _, c := range components { // <4> // HL
					switch c {
					case TobaccoComponent:
						select {
						case <-stop: // <5> // HL
							return
						case tobaccoCh <- Tobacco{}:
							log.Printf("agent: sent %v\n", TobaccoComponent)
						}
					case PapersComponent:
						select {
						case <-stop:
							return
						case papersCh <- Papers{}:
							log.Printf("agent: sent %v\n", PapersComponent)
						}
					case LighterComponent:
						select {
						case <-stop:
							return
						case lighterCh <- Lighter{}:
							log.Printf("agent: sent %v\n", LighterComponent)
						}
					default:
						log.Fatalf("agent: received unknown value %d", c)
						return
					}
				}
			}
		}
	}()

	return done, ac

}

// STOPAGENT OMIT

// STARTTABLESIG OMIT
func table(
	stop <-chan struct{}, ac AvailableComponents,
) (<-chan struct{}, SmokerWithTobacco, SmokerWithPapers, SmokerWithLighter) {
	// STOPTABLESIG OMIT

	var (
		done = make(chan struct{})

		// STARTTABLESWT OMIT
		swtPapersCh  = make(chan Papers)
		swtLighterCh = make(chan Lighter)
		swt          = SmokerWithTobacco{
			Papers:  swtPapersCh,
			Lighter: swtLighterCh,
		}
		// STOPTABLESWT OMIT

		swpTobaccoCh = make(chan Tobacco)
		swpLighterCh = make(chan Lighter)
		swp          = SmokerWithPapers{
			Tobacco: swpTobaccoCh,
			Lighter: swpLighterCh,
		}

		swlTobaccoCh = make(chan Tobacco)
		swlPapersCh  = make(chan Papers)
		swl          = SmokerWithLighter{
			Tobacco: swlTobaccoCh,
			Papers:  swlPapersCh,
		}
	)

	cleanup := func() {
		defer close(done)

		log.Println("table: shutdown complete")
	}

	go func() {
		defer cleanup()

		for {

			// STARTFETCHCOMPONENTS OMIT
			var (
				hasPapers  bool
				hasLighter bool
				hasTobacco bool
			)

			for components := 0; components < 2; components++ {
				select {
				case <-stop:
					return
				case <-ac.Tobacco:
					hasTobacco = true
				case <-ac.Papers:
					hasPapers = true
				case <-ac.Lighter:
					hasLighter = true
				}
			}

			// STOPFETCHCOMPONENTS OMIT

			// STARTSWTSEND OMIT
			if hasPapers && hasLighter {
				select {
				case <-stop:
					return
				case swtPapersCh <- Papers{}:
				}

				select {
				case <-stop:
					return
				case swtLighterCh <- Lighter{}:
				}

				continue
			}

			// STOPSWTSEND OMIT

			if hasTobacco && hasLighter {
				select {
				case <-stop:
					return
				case swpTobaccoCh <- Tobacco{}:
				}

				select {
				case <-stop:
					return
				case swpLighterCh <- Lighter{}:
				}

				continue
			}

			if hasPapers && hasTobacco {
				select {
				case <-stop:
					return
				case swlPapersCh <- Papers{}:
				}

				select {
				case <-stop:
					return
				case swlTobaccoCh <- Tobacco{}:
				}
			}

		}

	}()

	return done, swt, swp, swl
}

func hasLighter(
	stop <-chan struct{},
	tobaccoCh <-chan Tobacco,
	papersCh <-chan Papers,
) (<-chan struct{}, <-chan string) {

	var (
		done     = make(chan struct{})
		messages = make(chan string)
	)

	cleanup := func() {
		defer close(done)

		close(messages)
		log.Println("hasLighter: shutdown complete")
	}

	go func() {
		defer cleanup()

		for {
			var hasPapers bool
			var hasTobacco bool

			for !(hasPapers && hasTobacco) {
				select {
				case <-stop:
					return
				case <-papersCh:
					hasPapers = true
				case <-tobaccoCh:
					hasTobacco = true
				}
			}

			msg := fmt.Sprintf("hasLighter: rolled and smoked a cigarette")

			select {
			case <-stop:
				return
			case messages <- msg:
			}
		}
	}()

	return done, messages
}

func hasPapers(
	stop <-chan struct{},
	tobaccoCh <-chan Tobacco,
	lighterCh <-chan Lighter,
) (<-chan struct{}, <-chan string) {

	var (
		done     = make(chan struct{})
		messages = make(chan string)
	)

	cleanup := func() {
		defer close(done)

		close(messages)
		log.Println("hasPapers: shutdown complete")
	}

	go func() {
		defer cleanup()

		for {
			var hasTobacco bool
			var hasLighter bool

			for !(hasTobacco && hasLighter) {
				select {
				case <-stop:
					return
				case <-tobaccoCh:
					hasTobacco = true
				case <-lighterCh:
					hasLighter = true
				}
			}

			msg := fmt.Sprintf("hasPapers: rolled and smoked a cigarette")

			select {
			case <-stop:
				return
			case messages <- msg:
			}
		}
	}()

	return done, messages
}

// STARTHASTOBACCO OMIT
func hasTobacco(
	stop <-chan struct{},
	papersCh <-chan Papers,
	lighterCh <-chan Lighter,
) (<-chan struct{}, <-chan string) {

	var (
		done     = make(chan struct{}) // <1> // HL
		messages = make(chan string)
	)

	cleanup := func() { // <2> // HL
		defer close(done)

		close(messages)
		log.Println("hasTobacco: shutdown complete")
	}

	go func() {
		defer cleanup()

		for {
			var hasPapers bool
			var hasLighter bool

			for !(hasPapers && hasLighter) { // <3> // HL
				select {
				case <-stop:
					return
				case <-papersCh:
					hasPapers = true
				case <-lighterCh:
					hasLighter = true
				}
			}

			msg := fmt.Sprintf("hasTobacco: rolled and smoked a cigarette") // <4> // HL

			select {
			case <-stop:
				return
			case messages <- msg: // <5> // HL
			}
		}
	}()

	return done, messages
}

// STOPHASTOBACCO OMIT

// STARTSIGNALERSIG OMIT
func signaler(stop <-chan struct{}, delay time.Duration) <-chan struct{} {
	// STOPSIGNALERSIG OMIT

	// STARTSIGNALERSETUP OMIT
	var (
		done   = make(chan struct{})
		signal = make(chan struct{})

		// Start the agent goroutine.
		agentDone, forTable = agent(stop, signal)

		// Start gathering supplies on the table and passing them to
		// smokers.
		tableDone, forSWT, forSWP, forSWL = table(stop, forTable)

		// Start each of the smokers.
		tobaccoDone, tobaccoMsgs = hasTobacco(stop, forSWT.Papers,
			forSWT.Lighter)

		papersDone, papersMsgs = hasPapers(stop, forSWP.Tobacco,
			forSWP.Lighter)

		lighterDone, lighterMsgs = hasLighter(stop, forSWL.Tobacco,
			forSWL.Papers)
	)

	// STOPSIGNALERSETUP OMIT

	// STARTWAIT OMIT
	// Wait blocks until all the done channels are closed.
	wait := func() {
		<-agentDone
		<-tableDone
		<-tobaccoDone
		<-papersDone
		<-lighterDone
	}

	// STOPWAIT OMIT

	// STARTSIGNALLOOP OMIT
	// Start signaling to the agent to drop new components after receiving
	// a message from a smoker.
	go func() {
		defer close(done)
		defer log.Println("signaler: shutdown complete")
		defer close(signal)
		defer wait() // <1> // HL
		defer log.Println("signaler: starting shutdown")

		// Initialize the agent by sending a signal.
		signal <- struct{}{} // <2> // HL

		for {
			select {
			case <-stop:
				return
			case msg, ok := <-tobaccoMsgs:
				if !ok {
					return
				}
				log.Println(msg)
			case msg, ok := <-papersMsgs:
				if !ok {
					return
				}
				log.Println(msg)
			case msg, ok := <-lighterMsgs:
				if !ok {
					return
				}
				log.Println(msg)
			}

			time.Sleep(delay) // <3> // HL

			select {
			case <-stop:
				return
			case signal <- struct{}{}: // <4> // HL
			}
		}
	}()
	// STOPSIGNALLOOP OMIT

	return done
}

// STARTMAIN OMIT
func main() {
	var (
		stop        = make(chan struct{})
		timeout     = 3 * time.Second
		delay       = timeout / 9
		isSignaling = signaler(stop, delay)
	)

	time.AfterFunc(timeout, func() {
		log.Println("main: shutting down")
		close(stop)
	})

	defer log.Println("main: shutdown complete")
	<-isSignaling
}

// STOPMAIN OMIT
