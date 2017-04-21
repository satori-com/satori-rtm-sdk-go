// Observer embedded package.
//
// Allows to extend any struct with ability to Fire events and ability
// to listen for any events.
// Check the Examples section to get information how to use the package
//
// All actions, like On, Once, Fire, Off called one by one, so if you call Fire, On, Once or Off multiple times,
// Observer gurantees, that the events will be executed in the same order.
//
// Thread-safe: yes
package observer

import (
	"container/list"
	"sync/atomic"
)

const (
	EVENT_QUEUE_LEN = 500
)

type Observer struct {
	events     map[string]*list.List
	eventQueue chan observerEvent
	id         int32
}

type callbackT struct {
	eventName string
	id        int32
	callback  func(interface{})
	onetime   bool
}

type observerEvent struct {
	Type string
	Data interface{}
}
type fireEvent struct {
	eventName string
	data      interface{}
}
type unregisterEvent struct {
	eventName string
	id        interface{}
}

// Starts observer behavior for object
func New() Observer {
	o := Observer{}
	o.initEvents()

	return o
}

// Init processing for internal Events queue
func (o *Observer) initEvents() {
	o.events = make(map[string]*list.List)
	o.eventQueue = make(chan observerEvent, EVENT_QUEUE_LEN)

	go o.handleQueue()
}

// Adds listener for an event.
// Callback function will be called when fire is Fired.
//
// Callback WILL NOT BE removed after event is occurred and will be called on every Fire
func (o *Observer) On(eventName string, callback func(interface{})) interface{} {
	return o.addCallback(eventName, callback, false)
}

// Adds listener for an event.
// Callback function will be called when fire is Fired.
//
// Callback WILL BE removed after event is occurred. It is one-time callback
func (o *Observer) Once(eventName string, callback func(interface{})) interface{} {
	return o.addCallback(eventName, callback, true)
}

// Unsubscribes from an event. Use the id from the Observer.On() to remove callback function
func (o *Observer) Off(eventName string, id interface{}) {
	o.eventQueue <- observerEvent{
		Type: "unregister",
		Data: unregisterEvent{
			eventName: eventName,
			id:        id,
		},
	}
}

// Fires event. Executes callback functions and passes data to them
func (o *Observer) Fire(eventName string, data interface{}) {
	o.eventQueue <- observerEvent{
		Type: "fire",
		Data: fireEvent{
			eventName: eventName,
			data:      data,
		},
	}
}

func (o *Observer) handleQueue() {
	for event := range o.eventQueue {
		switch event.Type {
		case "register":
			e := event.Data.(callbackT)
			if _, ok := o.events[e.eventName]; !ok {
				o.events[e.eventName] = list.New()
			}

			o.events[e.eventName].PushBack(e)
		case "unregister":
			e := event.Data.(unregisterEvent)
			o.deleteCallback(e.eventName, e.id)
		case "fire":
			e := event.Data.(fireEvent)
			if _, ok := o.events[e.eventName]; ok {
				for item := o.events[e.eventName].Front(); item != nil; item = item.Next() {
					callback := item.Value.(callbackT)
					callback.callback(e.data)
					if callback.onetime {
						o.deleteCallback(callback.eventName, callback.id)
					}
				}
			}
		}
	}
}

func (o *Observer) deleteCallback(eventName string, id interface{}) {
	if _, ok := o.events[eventName]; ok {
		callbackId := id.(int32)
		for item := o.events[eventName].Front(); item != nil; item = item.Next() {
			if item.Value.(callbackT).id == callbackId {
				o.events[eventName].Remove(item)
				break
			}
		}
	}
}

func (o *Observer) addCallback(eventName string, callback func(interface{}), onetime bool) interface{} {
	id := o.nextId()
	o.eventQueue <- observerEvent{
		Type: "register",
		Data: callbackT{
			eventName: eventName,
			callback:  callback,
			onetime:   onetime,
			id:        id,
		},
	}

	return id
}

func (o *Observer) nextId() int32 {
	atomic.AddInt32(&o.id, 1)
	return o.id
}
