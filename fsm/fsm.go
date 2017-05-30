// Go FSM package provides a lightweight finite state machine for Golang.
// FSM allows you to create States with numbers of events for the state.
//
// When creating a new FSM you should describe all possible states and describe possible events.
// Each event can make a transition to another state.
// You can specify events on Enter/Leave the event and
// check if transition has been completed successfully.
//
// Thread-safe: no
package fsm

import (
	"errors"
	"strings"
	"sync"
)

type FSM struct {
	currentState StateName
	states       States
	mutex        sync.RWMutex
}

var (
	ERROR_NO_STATE_EXISTS       = errors.New("Unable to change state: Undefined destination state")
	ERROR_STATE_EVENT_NOT_FOUND = errors.New("Unable to fire event. Current state has no such event")
	ERROR_WRONG_INITIAL_STATE   = errors.New("Wrong initial state")
)

type States map[StateName]Events
type StateName string

type EventName string
type EventHandler func(f *FSM)
type Events map[EventName]EventHandler

// Creates an instance of FSM
func New(initialState StateName, states States) (*FSM, error) {
	if _, ok := states[initialState]; !ok {
		return nil, ERROR_WRONG_INITIAL_STATE
	}

	f := &FSM{
		currentState: initialState,
		states:       states,
	}

	return f, nil
}

// Fires an event of current state
// Returns ERROR_STATE_EVENT_NOT_FOUND error if state not found or type State
func (f *FSM) Event(event EventName) error {
	if e, ok := f.states[f.currentState][event]; ok {
		e(f)
	}
	return ERROR_STATE_EVENT_NOT_FOUND
}

// Makes transition to a new State
// Returns ERROR_NO_STATE_EXISTS if State does not exist
func (f *FSM) Transition(destination StateName) error {
	if _, ok := f.states[destination]; ok {
		f.Event(EventName("leave" + strings.Title(string(f.currentState))))
		f.setState(destination)
		f.Event(EventName("enter" + strings.Title(string(f.currentState))))

		return nil
	}
	return ERROR_NO_STATE_EXISTS
}

// Gets current state of the FSM instance
func (f *FSM) CurrentState() StateName {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.currentState
}

// Sets new state. Should not be called directly. Uses from Transition() only
func (f *FSM) setState(state StateName) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.currentState = state
}
