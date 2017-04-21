package fsm

import (
	"testing"
)

func TestInitiation(t *testing.T) {
	f, _ := New("stopped", States{"stopped": Events{}})
	if f.CurrentState() != "stopped" {
		t.Fatalf("Expected state '%s' did not match the actual '%s'", f.CurrentState(), "stopped")
	}
}

func TestWrongInitialState(t *testing.T) {
	_, err := New("stopped", States{"started": Events{}})
	if err == nil {
		t.Fatalf("Unable to get '%s' error", ERROR_WRONG_INITIAL_STATE.Error())
	}
}

func TestTransition(t *testing.T) {
	f, _ := New("closed", States{
		"closed": Events{
			"open": func(f *FSM) {
				f.Transition("opened")
			},
		},
		"opened": Events{
			"close": func(f *FSM) {
				f.Transition("closed")
			},
		},
	})

	if f.CurrentState() != "closed" {
		t.Fatalf("Expected state '%s' did not match the actual '%s'", f.CurrentState(), "closed")
	}

	f.Event("open")
	if f.CurrentState() != "opened" {
		t.Fatalf("Expected state '%s' did not match the actual '%s'", f.CurrentState(), "opened")
	}

	f.Event(EventName("close"))
	if f.CurrentState() != "closed" {
		t.Fatalf("Expected state '%s' did not match the actual '%s'", f.CurrentState(), "closed")
	}
}

func TestStateEvents(t *testing.T) {
	enterClosed := false
	leaveClosed := false

	f, _ := New("closed", States{
		"closed": Events{
			"enterClosed": func(f *FSM) {
				enterClosed = true
			},
			"open": func(f *FSM) {
				f.Transition("opened")
			},
			"leaveClosed": func(f *FSM) {
				leaveClosed = true
			},
		},
		"opened": Events{
			"close": func(f *FSM) {
				f.Transition("closed")
			},
		},
	})

	f.Event("open")
	if !leaveClosed {
		t.Fatal("leaveClosed event did not occured")
	}
	if enterClosed {
		t.Fatal("enterClosed occured but should not")
	}
	f.Event(EventName("close"))
	if !enterClosed {
		t.Fatal("enterClosed event did not occured")
	}
}

func TestWrongTransitionState(t *testing.T) {
	f, _ := New("closed", States{
		"closed": Events{},
	})

	err := f.Transition("opened")
	if err != ERROR_NO_STATE_EXISTS {
		t.Fatal("ERROR_NO_STATE_EXISTS error has not been returned")
	}
}
