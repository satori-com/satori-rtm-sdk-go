package observer_test

import (
	"errors"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/observer"
	"time"
	"sync"
)

// Embed Observer behavior to a custom struct and call/catch events
func ExampleObserver() {
	var wg sync.WaitGroup
	type MyObject struct {
		observer.Observer // Implements observer behavior
	}

	o := MyObject{
		Observer: observer.New(),
	}
	wg.Add(3)

	o.On("myevent", func(data interface{}) {
		fmt.Println("Got event", data.(int))
		wg.Done()
	})
	o.Once("myevent", func(data interface{}) {
		fmt.Println("Got once event", data.(int))
		wg.Done()
	})

	o.Fire("myevent", 1)
	o.Fire("myevent", 2)

	// Wait for events
	wg.Wait()

	// Output:
	// Got event 1
	// Got once event 1
	// Got event 2
}

// Embed Observer: Fire event and transfer data to event handler
func ExampleObserver_Fire() {
	var wg sync.WaitGroup
	type MyObject struct {
		observer.Observer // Implements observer behavior
	}

	o := MyObject{
		Observer: observer.New(),
	}
	wg.Add(1)

	o.On("error", func(err interface{}) {
		fmt.Println(err.(error).Error())
		wg.Done()
	})

	o.Fire("error", errors.New("My custom error"))

	// Wait for event
	wg.Wait()

	// Output: My custom error
}

// Embed Observer: Wait for event using channels
func ExampleObserver_On() {
	var wg sync.WaitGroup
	type MyObject struct {
		observer.Observer // Implements observer behavior
	}

	o := MyObject{
		Observer: observer.New(),
	}
	wg.Add(1)

	go func() {
		<-time.After(10 * time.Millisecond)
		o.Fire("connected", nil)
	}()

	o.On("connected", func(err interface{}) {
		wg.Done()
	})

	// Wait for the event
	wg.Wait()
	fmt.Println("Action after event")

	// Output: Action after event
}
