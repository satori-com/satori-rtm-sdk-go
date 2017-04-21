package subscription

import (
	"encoding/json"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"testing"
	"time"
)

func TestSubscribePdu(t *testing.T) {
	sub := New("test", RELIABLE, pdu.SubscribeBodyOpts{
		Filter: "SELECT * FROM `test`",
		History: pdu.SubscribeHistory{
			Count: 1,
			Age:   10,
		},
		Position: "123456789",
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
	sub := New("test", RELIABLE, pdu.SubscribeBodyOpts{})
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
	sub := New("reliable", RELIABLE, pdu.SubscribeBodyOpts{})
	if sub.mode.trackPosition != true || sub.mode.fastForward != true {
		t.Error("RELIABLE mode sets wrong flags")
	}

	sub = New("simple", SIMPLE, pdu.SubscribeBodyOpts{})
	if sub.mode.trackPosition != false || sub.mode.fastForward != true {
		t.Error("SIMPLE mode sets wrong flags")
	}

	sub = New("advanced", ADVANCED, pdu.SubscribeBodyOpts{})
	if sub.mode.trackPosition != true || sub.mode.fastForward != false {
		t.Error("ADVANCED mode sets wrong flags")
	}
}

func TestStates(t *testing.T) {
	sub := New("reliable", RELIABLE, pdu.SubscribeBodyOpts{})
	if sub.GetState() != STATE_UNSUBSCRIBED {
		t.Error("Subscription has SUBSCRIBED state, but should not")
	}

	sub.OnSubscribe(pdu.SubscribeOk{
		Position:       "1",
		SubscriptionId: "reliable",
	})

	if sub.GetState() != STATE_SUBSCRIBED {
		t.Error("Subscription has UNSUBSCRIBED state, but should not")
	}

	sub.OnDisconnect()
	if sub.GetState() != STATE_UNSUBSCRIBED {
		t.Error("Subscription has SUBSCRIBED state, but should not")
	}
}

func TestDataChannel(t *testing.T) {
	sub := New("reliable", RELIABLE, pdu.SubscribeBodyOpts{})
	messages := []json.RawMessage{
		json.RawMessage("hello"),
		json.RawMessage("\"{}\""),
		json.RawMessage("!@#123"),
	}
	sub.OnData(pdu.SubscriptionData{
		Position:       "123",
		Messages:       messages,
		SubscriptionId: "test",
	})

	expected := []string{"hello", "\"{}\"", "!@#123"}
	for i := 0; i < 3; i++ {
		select {
		case message := <-sub.Data():
			if string(message) != expected[0] {
				t.Fatal("Messages do not match. Expexted: 'test'. Actual: " + string(message))
			}
			expected = expected[1:]
		default:
			t.Fatal("Failed to get message")
		}
	}
}

func TestDataChannelOverflow(t *testing.T) {
	sub := New("reliable", RELIABLE, pdu.SubscribeBodyOpts{})
	messages := []json.RawMessage{
		json.RawMessage("hello"),
	}

	processed := make(chan bool)
	var counter int
overflow:
	for counter = 0; counter <= MAX_QUEUE*2; counter++ {
		go func() {
			sub.OnData(pdu.SubscriptionData{
				Position:       "123",
				Messages:       messages,
				SubscriptionId: "test",
			})
			processed <- true
		}()
		select {
		case <-processed:
		case <-time.After(100 * time.Millisecond):
			break overflow
		}
	}
	if len(sub.Data()) != MAX_QUEUE {
		t.Fatal("MAX_QUEUE limit exceeded")
	}
}

func TestOnError(t *testing.T) {
	sub := New("reliable", RELIABLE, pdu.SubscribeBodyOpts{
		Position: "123456789",
	})

	event := make(chan bool)
	sub.On("subscribeError", func(data interface{}) {
		event <- true
	})

	sub.OnSubscribeError(pdu.SubscribeError{
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
	sub := New("reliable", RELIABLE, pdu.SubscribeBodyOpts{
		Position: "123456789",
	})

	sub.OnInfo(pdu.SubscriptionInfo{
		Info:     "fast_forward",
		Reason:   "slow read",
		Position: "987654321",
	})

	if sub.position != "987654321" {
		t.Fatal("Unable to process subscription info")
	}
}

func TestSubscription_GetSubscriptionId(t *testing.T) {
	sub := New("test123", RELIABLE, pdu.SubscribeBodyOpts{})
	if sub.GetSubscriptionId() != "test123" {
		t.Fatal("Wrong subscription ID")
	}
}

func TestEvents(t *testing.T) {
	sub := New("test", RELIABLE, pdu.SubscribeBodyOpts{})

	event := make(chan bool)
	sub.On("subscribed", func(data interface{}) {
		event <- true
	})

	sub.OnSubscribe(pdu.SubscribeOk{
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
	sub := New("test123", SIMPLE, pdu.SubscribeBodyOpts{
		Position: "123",
	})

	var subPdu pdu.SubscribeBody
	json.Unmarshal(sub.SubscribePdu().Body, &subPdu)

	if subPdu.Position != "123" {
		t.Fatal("Position for SIMPLE mode is wrong")
	}

	sub.OnSubscribe(pdu.SubscribeOk{
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
	sub := New("test123", RELIABLE, pdu.SubscribeBodyOpts{
		Position: "123",
	})

	var subPdu pdu.SubscribeBody
	json.Unmarshal(sub.SubscribePdu().Body, &subPdu)

	if subPdu.Position != "123" {
		t.Fatal("Position for SIMPLE mode is wrong")
	}

	sub.OnSubscribe(pdu.SubscribeOk{
		Position:       "321",
		SubscriptionId: "test123",
	})

	json.Unmarshal(sub.SubscribePdu().Body, &subPdu)
	if subPdu.Position != "321" {
		t.Fatal("Position is wrong after connected")
	}
}
