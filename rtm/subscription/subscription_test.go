package subscription

import (
	"encoding/json"
	"errors"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"reflect"
	"testing"
	"time"
)

func TestSubscribePdu(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "test",
		Mode:           RELIABLE,
		Opts: pdu.SubscribeBodyOpts{
			Filter: "SELECT * FROM `test`",
			History: pdu.SubscribeHistory{
				Count: 1,
				Age:   10,
			},
			Position: "123456789",
		},
	})
	subPdu := sub.SubscribePdu()

	if subPdu.Action != "rtm/subscribe" {
		t.Errorf("ID mismatch: %s != %s", subPdu.Action, "rtm/subscribe")
	}

	var subBody pdu.SubscribeBody
	json.Unmarshal(subPdu.Body, &subBody)

	expected := pdu.SubscribeBody{
		Force:          true,
		FastForward:    true,
		SubscriptionId: "test",
		Filter:         "SELECT * FROM `test`",
		History: pdu.SubscribeHistory{
			Count: 1,
			Age:   10,
		},
		Position: "123456789",
	}

	if subBody != expected {
		t.Errorf("Unexpected body:\nActual: %+v\nExpect: %+v", subBody, expected)
	}
}

func TestUnsubscribePdu(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "test",
		Mode:           RELIABLE,
	})
	unsubPdu := sub.UnsubscribePdu()

	if unsubPdu.Action != "rtm/unsubscribe" {
		t.Errorf("ID mismatch: %s != %s", unsubPdu.Action, "rtm/subscribe")
	}

	expected := "{\"subscription_id\":\"test\"}"
	if string(unsubPdu.Body) != expected {
		t.Errorf("Unexpected body:\nActual: %s\nExpected: %s", unsubPdu.Body, expected)
	}
}

func TestModes(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "reliable",
		Mode:           RELIABLE,
	})
	if sub.mode.trackPosition != true || sub.mode.fastForward != true {
		t.Error("RELIABLE mode sets wrong flags")
	}

	sub = New(Config{
		SubscriptionId: "simple",
		Mode:           SIMPLE,
	})
	if sub.mode.trackPosition != false || sub.mode.fastForward != true {
		t.Error("SIMPLE mode sets wrong flags")
	}

	sub = New(Config{
		SubscriptionId: "advanced",
		Mode:           ADVANCED,
	})
	if sub.mode.trackPosition != true || sub.mode.fastForward != false {
		t.Error("ADVANCED mode sets wrong flags")
	}
}

func TestStates(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "reliable",
		Mode:           RELIABLE,
	})
	if sub.GetState() != STATE_UNSUBSCRIBED {
		t.Error("Subscription has SUBSCRIBED state, but should not")
	}

	sub.ProcessSubscribe(pdu.SubscribeOk{
		Position:       "1",
		SubscriptionId: "reliable",
	})

	if sub.GetState() != STATE_SUBSCRIBED {
		t.Error("Subscription has UNSUBSCRIBED state, but should not")
	}

	sub.ProcessDisconnect()
	if sub.GetState() != STATE_UNSUBSCRIBED {
		t.Error("Subscription has SUBSCRIBED state, but should not")
	}
}

func TestDataChannel(t *testing.T) {
	msgC := make(chan string, 3)

	listener := Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				msgC <- string(message)
			}
		},
	}
	config := Config{
		SubscriptionId: "reliable",
		Mode:           RELIABLE,
		Listener:       listener,
	}
	sub := New(config)
	messages := []json.RawMessage{
		json.RawMessage("hello"),
		json.RawMessage("\"{}\""),
		json.RawMessage("!@#123"),
	}
	sub.ProcessData(pdu.SubscriptionData{
		Position:       "123",
		Messages:       messages,
		SubscriptionId: "test",
	})

	expected := []string{"hello", "\"{}\"", "!@#123"}
	for i := 0; i < 3; i++ {
		select {
		case message := <-msgC:
			if message != expected[0] {
				t.Fatal("Messages do not match. Expexted: 'test'. Actual: " + message)
			}
			expected = expected[1:]
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Failed to get messages")
		}
	}
}

func TestOnError(t *testing.T) {
	event := make(chan bool)

	listener := Listener{
		OnSubscribeError: func(err pdu.SubscribeError) {
			event <- true
		},
	}
	sub := New(Config{
		SubscriptionId: "reliable",
		Opts: pdu.SubscribeBodyOpts{
			Position: "123456789",
		},
		Mode:     RELIABLE,
		Listener: listener,
	})
	go sub.ProcessSubscribeError(pdu.SubscribeError{
		Error:  "test_error",
		Reason: "no_reason",
	})

	select {
	case <-event:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Error event did not occur")
	}
}

func TestOnInfo(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "reliable",
		Mode:           RELIABLE,
		Opts: pdu.SubscribeBodyOpts{
			Position: "123456789",
		},
	})
	sub.ProcessInfo(pdu.SubscriptionInfo{
		Info:     "fast_forward",
		Reason:   "slow read",
		Position: "987654321",
	})

	if sub.position != "987654321" {
		t.Fatal("Unable to process subscription info")
	}
}

func TestSubscription_GetSubscriptionId(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "test123",
		Mode:           RELIABLE,
	})
	if sub.GetSubscriptionId() != "test123" {
		t.Fatal("Wrong subscription ID")
	}
}

func TestEvents(t *testing.T) {
	event := make(chan bool)

	listener := Listener{
		OnSubscribed: func(data pdu.SubscribeOk) {
			event <- true
		},
	}
	sub := New(Config{
		SubscriptionId: "test",
		Mode:           RELIABLE,
		Listener:       listener,
	})
	go sub.ProcessSubscribe(pdu.SubscribeOk{
		Position:       "1234567",
		SubscriptionId: "test",
	})

	select {
	case <-event:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Error event did not occur")
	}
}

func TestSimpleSubscription(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "test123",
		Mode:           SIMPLE,
		Opts: pdu.SubscribeBodyOpts{
			Position: "123",
		},
	})

	var subPdu pdu.SubscribeBody
	json.Unmarshal(sub.SubscribePdu().Body, &subPdu)

	if subPdu.Position != "123" {
		t.Fatal("Position for SIMPLE mode is wrong")
	}

	sub.ProcessSubscribe(pdu.SubscribeOk{
		Position:       "321",
		SubscriptionId: "test123",
	})

	subPdu = pdu.SubscribeBody{}
	json.Unmarshal(sub.SubscribePdu().Body, &subPdu)
	if subPdu.Position != "" {
		t.Fatal("Position still exists, but should not")
	}
}

func TestReliableSubscription(t *testing.T) {
	sub := New(Config{
		SubscriptionId: "test123",
		Mode:           RELIABLE,
		Opts: pdu.SubscribeBodyOpts{
			Position: "123",
		},
	})

	var subPdu pdu.SubscribeBody
	json.Unmarshal(sub.SubscribePdu().Body, &subPdu)

	if subPdu.Position != "123" {
		t.Fatal("Position for SIMPLE mode is wrong")
	}

	sub.ProcessSubscribe(pdu.SubscribeOk{
		Position:       "321",
		SubscriptionId: "test123",
	})

	json.Unmarshal(sub.SubscribePdu().Body, &subPdu)
	if subPdu.Position != "321" {
		t.Fatal("Position is wrong after connected")
	}
}

func TestStructTransfer(t *testing.T) {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	occurred := make(chan bool)

	listener := Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				var msg Message
				json.Unmarshal(message, &msg)

				if msg.Who != "zebra" || reflect.DeepEqual(&msg.Where, []float32{34.134358, -118.321506}) {
					t.Fatal("Wrong decoded message")
				}
				occurred <- true
			}
		},
	}
	sub := New(Config{
		SubscriptionId: "test123",
		Mode:           RELIABLE,
		Listener:       listener,
	})

	message, _ := json.Marshal(&Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})
	messages := []json.RawMessage{message}
	go sub.ProcessData(pdu.SubscriptionData{
		Position:       "123",
		Messages:       messages,
		SubscriptionId: "test123",
	})

	select {
	case <-occurred:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Unable to transfer Struct")
	}
}

func TestSubscriptionEvents(t *testing.T) {
	var subId string = "test123"
	event := make(chan bool)

	listener := Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				if string(message) != "hello" {
					t.Fatal("Wrong OnData message")
				}
				event <- true
			}
		},
		OnSubscribed: func(sok pdu.SubscribeOk) {
			if sok.Position != "123456789" {
				t.Fatal("Wrong position in OnSubscribed")
			}
			event <- true
		},
		OnSubscriptionInfo: func(info pdu.SubscriptionInfo) {
			if info.Info != "sub_info" || info.Reason != "sub_reason" {
				t.Fatal("Wrong reason or info in OnInfo")
			}
			event <- true
		},
		OnSubscriptionError: func(err pdu.SubscriptionError) {
			if err.Reason != "sub_reason" || err.Error != "sub_error" {
				t.Fatal("Wrong reason or info in OnSubscriptionError")
			}
			event <- true
		},
	}

	sub := New(Config{
		SubscriptionId: subId,
		Mode:           RELIABLE,
		Listener:       listener,
	})

	go sub.ProcessData(pdu.SubscriptionData{
		Position: "123456789",
		Messages: []json.RawMessage{
			json.RawMessage("hello"),
		},
		SubscriptionId: subId,
	})
	if err := waitEvent(event); err != nil {
		t.Fatal(err)
	}

	go sub.ProcessSubscribe(pdu.SubscribeOk{
		Position:       "123456789",
		SubscriptionId: subId,
	})
	if err := waitEvent(event); err != nil {
		t.Fatal(err)
	}

	go sub.ProcessInfo(pdu.SubscriptionInfo{
		Reason:   "sub_reason",
		Info:     "sub_info",
		Position: "123",
	})
	if err := waitEvent(event); err != nil {
		t.Fatal(err)
	}

	go sub.ProcessSubscriptionError(pdu.SubscriptionError{
		Error:    "sub_error",
		Reason:   "sub_reason",
		Position: "123",
	})
	if err := waitEvent(event); err != nil {
		t.Fatal(err)
	}
}

func waitEvent(event <-chan bool) error {
	select {
	case <-event:
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New("Timeout: Event did not occurred")
	}
}

// Tests memory leaks. Shows results only when running
// "go test" together with "-gcflags=-m -run TestSubscriptionMemoryLeak"
func TestSubscriptionMemoryLeak(t *testing.T) {
	event := make(chan bool)

	listener := Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				if string(message) != "hello" {
					t.Fatal("Wrong OnData message")
				}
				event <- true
			}
		},
	}
	sub := New(Config{
		SubscriptionId: "test",
		Mode:           RELIABLE,
		Listener:       listener,
	})

	for i := 0; i <= 100000; i++ {
		go sub.ProcessData(pdu.SubscriptionData{
			Position: "123456789",
			Messages: []json.RawMessage{
				json.RawMessage("hello"),
			},
			SubscriptionId: "test",
		})
		if err := waitEvent(event); err != nil {
			t.Fatal(err)
		}
	}
}
