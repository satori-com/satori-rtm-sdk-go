package observer_test

import (
	"errors"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/observer"
	"time"
)

// Embed Observer behavior to a custom struct and call/catch events
func ExampleObserver() {
	type MyObject struct {
		observer.Observer // Implements observer behavior
	}

	o := MyObject{
		Observer: observer.New(),
	}

	o.On("myevent", func(data interface{}) {
		fmt.Println("Got event", data.(int))
	})
	o.Once("myevent", func(data interface{}) {
		fmt.Println("Got once event", data.(int))
	})

	o.Fire("myevent", 1)
	o.Fire("myevent", 2)

	// Wait for events
	<-time.After(10 * time.Millisecond)

	// Output:
	// Got event 1
	// Got once event 1
	// Got event 2
}

// Embed Observer: Fire event and transfer data to event handler
func ExampleObserver_Fire() {
	type MyObject struct {
		observer.Observer // Implements observer behavior
	}

	o := MyObject{
		Observer: observer.New(),
	}

	o.On("error", func(err interface{}) {
		fmt.Println(err.(error).Error())
	})

	o.Fire("error", errors.New("My custom error"))

	// Wait for events
	<-time.After(10 * time.Millisecond)

	// Output: My custom error
}

// Embed Observer: Wait for event using channels
func ExampleObserver_On() {
	type MyObject struct {
		observer.Observer // Implements observer behavior
	}

	o := MyObject{
		Observer: observer.New(),
	}

	go func() {
		<-time.After(10 * time.Millisecond)
		o.Fire("connected", nil)
	}()

	waitConnected := make(chan bool)
	o.On("connected", func(err interface{}) {
		waitConnected <- true
	})

	// Wait for the event
	<-waitConnected
	fmt.Println("Action after event")

	// Output: Action after event
}
