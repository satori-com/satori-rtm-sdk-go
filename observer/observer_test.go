package observer

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type A struct {
	Observer
}

func TestEvent(t *testing.T) {
	a := A{
		Observer: New(),
	}

	event := make(chan bool)
	go func() {
		time.Sleep(100 * time.Millisecond)
		a.Fire("myevent", nil)
	}()

	a.On("myevent", func(data interface{}) {
		event <- true
	})
	select {
	case <-event:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Event timeout")
	}
}

func TestMultipleListeners(t *testing.T) {
	a := A{
		Observer: New(),
	}

	eventA := make(chan bool)
	eventB := make(chan bool)

	a.On("event", func(data interface{}) {
		eventA <- true
	})
	a.On("event", func(data interface{}) {
		eventB <- true
	})

	go a.Fire("event", nil)

	for i := 0; i < 2; i++ {
		select {
		case <-eventA:
		case <-eventB:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Event A or/and B did not occur")
		}
	}
}

func TestDataTransfer(t *testing.T) {
	a := A{
		Observer: New(),
	}
	var data int
	event := make(chan bool)

	a.On("myevent", func(d interface{}) {
		data = d.(int)
		event <- true
	})

	a.Fire("myevent", 123)

	select {
	case <-event:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Data was not passed")
	}

	if data != 123 {
		t.Fatal("Data was not passed correctly")
	}
}

func TestObserverQueue(t *testing.T) {
	var wg sync.WaitGroup
	a := A{
		Observer: New(),
	}

	result := make([]string, 0)

	wg.Add(2)
	a.On("test", func(data interface{}) {
		result = append(result, data.(string))
		wg.Done()
	})
	a.On("test", func(data interface{}) {
		result = append(result, data.(string))
		wg.Done()
	})
	a.Fire("test", "hello")
	a.On("test", func(data interface{}) {
		result = append(result, data.(string))
	})

	wg.Wait()
	if !reflect.DeepEqual(result, []string{"hello", "hello"}) {
		t.Fatal("Wrong events order")
	}
}

func TestOff(t *testing.T) {
	a := A{
		Observer: New(),
	}

	results := make([]string, 0)
	id := a.On("myevent", func(data interface{}) {
		results = append(results, data.(string))
	})
	a.Fire("myevent", "1")
	a.Off("myevent", id)
	event := make(chan bool)
	a.On("myevent", func(data interface{}) {
		event <- true
	})
	a.Fire("myevent", "2")

	select {
	case <-event:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Second event did not occur")
	}

	if !reflect.DeepEqual(results, []string{"1"}) {
		t.Fatal("Unregister method does not work")
	}
}
