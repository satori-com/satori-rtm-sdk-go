package fsm_test

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/fsm"
)

// Create FSM and change its state. Set initial status to "closed"
func ExampleFSM() {
	f, err := fsm.New("closed", fsm.States{
		"closed": fsm.Events{
			"enterClosed": func(f *fsm.FSM) {
				fmt.Println("Enter the 'closed' state")
			},
			"open": func(f *fsm.FSM) {
				fmt.Println("Transition to the 'opened' state")
				f.Transition("opened")
			},
			"leaveClosed": func(f *fsm.FSM) {
				fmt.Println("Leave the 'closed' state")
			},
		},
		"opened": fsm.Events{
			"close": func(f *fsm.FSM) {
				fmt.Println("Transition to the 'closed' state")
				f.Transition("closed")
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	f.Event("open")
	f.Event("close")
	// Output:
	// Transition to the 'opened' state
	// Leave the 'closed' state
	// Transition to the 'closed' state
	// Enter the 'closed' state
}
