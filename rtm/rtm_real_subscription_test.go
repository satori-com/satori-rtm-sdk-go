package rtm

import (
	"encoding/json"
	"errors"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestWrongPosition(t *testing.T) {
	channel := getChannel()
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}

	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}
	event := make(chan int)

	listener := subscription.Listener{
		OnSubscribeError: func(err pdu.SubscribeError) {
			if err.Error != "invalid_format" {
				t.Fatal("Wrong subscription error")
			}
			event <- 1
		},
		OnSubscribed: func(sok pdu.SubscribeOk) {
			event <- 2
		},
	}
	client.Subscribe(
		channel,
		subscription.SIMPLE,
		pdu.SubscribeBodyOpts{
			Position: "wrong_position",
		},
		listener,
	)

	select {
	case e := <-event:
		if e != 1 {
			t.Fatal("Wrong event order")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Incorrect position error did not occured")
	}

	client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)
	select {
	case e := <-event:
		if e != 2 {
			t.Fatal("Wrong event order")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe with the correct position")
	}
}

func TestMultipleSubscription(t *testing.T) {
	channel := getChannel()
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if waitForConnected(client) != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	var errorOccured = false

	errC := make(chan bool)
	listener := subscription.Listener{
		OnSubscribeError: func(err pdu.SubscribeError) {
			errC <- true
		},
	}
	client.Subscribe(
		channel,
		subscription.SIMPLE,
		pdu.SubscribeBodyOpts{
			Position: "wrong_position",
		},
		listener,
	)

	go func() {
		select {
		case <-errC:
		case <-time.After(5 * time.Second):
			errorOccured = true
		}
		wg.Done()
	}()

	event := make(chan bool)
	listener = subscription.Listener{
		OnSubscribed: func(sok pdu.SubscribeOk) {
			event <- true
		},
	}
	client.Subscribe(
		channel,
		subscription.SIMPLE,
		pdu.SubscribeBodyOpts{
			Filter: "SELECT COUNT(*) FROM `test`",
		},
		listener,
	)

	go func() {
		select {
		case <-event:
		case <-time.After(5 * time.Second):
			errorOccured = true
		}
		wg.Done()
	}()

	listener = subscription.Listener{
		OnSubscribeError: func(err pdu.SubscribeError) {
			errC <- true
		},
	}

	client.Subscribe(
		channel,
		subscription.SIMPLE,
		pdu.SubscribeBodyOpts{
			Position: "wrong_position",
		},
		listener,
	)
	go func() {
		select {
		case <-errC:
		case <-time.After(5 * time.Second):
			errorOccured = true
		}
		wg.Done()
	}()

	wg.Wait()

	// Check the current subscription. Should be the subscription with filter
	sub, _ := client.GetSubscription(channel)
	subPdu := sub.SubscribePdu()

	actualPdu := pdu.SubscribeBodyOpts{}
	json.Unmarshal(subPdu.Body, &actualPdu)

	if actualPdu.Filter != "SELECT COUNT(*) FROM `test`" {
		t.Fatal("Wrong subcription is using")
	}
}

func TestSimpleSubscription(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)

	channel := getChannel()
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}

	defer client.Stop()
	client.Start()

	expected := []int{0, 1, 2}

	listener := subscription.Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				i, _ := strconv.Atoi(string(message))
				if expected[0] != i {
					t.Fatal("Wrong message order or wrong message")
				}
				expected = expected[1:]
				wg.Done()
			}
		},
		OnSubscribed: func(pdu.SubscribeOk) {
			for i := 0; i < 3; i++ {
				client.Publish(channel, i)
			}
		},
	}

	client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)

	wait := make(chan bool)
	go func() {
		wg.Wait()
		wait <- true
	}()

	select {
	case <-wait:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout: Test timeout")
	}
}

func TestSubscriptionFilter(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	wg.Add(2)

	channel := getChannel()
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}

	defer client.Stop()
	client.Start()

	subscribed := make(chan bool)
	expected := []string{"{\"test\":1}", "{\"test\":3}"}

	listener := subscription.Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				if expected[0] != string(message) {
					err = errors.New("Wrong actiual data. Expected: " + expected[0] + " Actual: " + string(message))
				}
				expected = expected[1:]
				wg.Done()
			}
		},
		OnSubscribed: func(pdu.SubscribeOk) {
			subscribed <- true
		},
	}
	client.Subscribe(
		channel,
		subscription.SIMPLE,
		pdu.SubscribeBodyOpts{
			Filter: "select * from `" + channel + "` where test != 2",
		},
		listener,
	)

	select {
	case <-subscribed:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe")
	}

	client.Publish(channel, json.RawMessage("{\"test\":1}"))
	client.Publish(channel, json.RawMessage("{\"test\":2}"))
	client.Publish(channel, json.RawMessage("{\"test\":3}"))

	wg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionAfterDisconnect(t *testing.T) {
	channel := getChannel()
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}

	defer client.Stop()
	client.Start()

	msgC := make(chan bool)
	subscribed := make(chan bool)
	expected := []string{"1", "2"}

	listener := subscription.Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				if string(message) != expected[0] {
					t.Fatal("Wrong subscription message")
				}
				expected = expected[1:]
				msgC <- true
			}
		},
		OnSubscribed: func(pdu.SubscribeOk) {
			subscribed <- true
		},
	}
	client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)

	select {
	case <-subscribed:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe")
	}

	go client.Publish(channel, 1)
	select {
	case <-msgC:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe")
	}

	// Drop connection
	client.conn.SetDeadline(time.Now())
	select {
	case <-subscribed:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe")
	}
	if err != nil {
		t.Fatal(err)
	}

	go client.Publish(channel, 2)
	select {
	case <-msgC:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to resubscribe after dropping connection")
	}
}

func TestRTM_Unsubscribe(t *testing.T) {
	channel := getChannel()
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}

	defer client.Stop()
	client.Start()

	expected := []int{0, 1, 2}
	msgC := make(chan bool)
	subscribed := make(chan bool)

	listener := subscription.Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				if len(expected) == 0 {
					t.Fatal("We got the message, but should not")
				}
				msg, _ := strconv.Atoi(string(message))
				if msg != expected[0] {
					t.Fatal("Wrong message order or wrong message")
				}
				expected = expected[1:]
				msgC <- true
			}
		},
		OnSubscribed: func(pdu.SubscribeOk) {
			subscribed <- true
		},
	}
	client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)

	select {
	case <-subscribed:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe")
	}

	go func() {
		for i := 0; i < 3; i++ {
			client.Publish(channel, i)
		}

	}()

	for i := 0; i < 3; i++ {
		select {
		case <-msgC:
		case <-time.After(5 * time.Second):
			t.Fatal("Unable to get message for subscription")
		}
	}

	c := <-client.Unsubscribe(channel)

	if c.Err != nil {
		t.Fatal("Unable to unsubscribe")
	}

	go func() {
		for i := 0; i < 3; i++ {
			client.Publish(channel, i)
		}

	}()

	select {
	case <-msgC:
		t.Fatal("We are still subscribed")
	case <-time.After(1 * time.Second):
	}
}
