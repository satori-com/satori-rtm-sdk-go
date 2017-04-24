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

	sub, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Position: "wrong_position",
	})

	event := make(chan bool)
	sub.Once("subscribeError", func(err interface{}) {
		data := err.(pdu.SubscribeError)
		if data.Error != "invalid_format" {
			t.Fatal("Wrong subscription error")
		}
		event <- true
	})
	select {
	case <-event:
	case <-time.After(5 * time.Second):
		t.Fatal("Incorrect position error did not occured")
	}

	sub, _ = client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{})
	sub.Once("subscribed", func(err interface{}) {
		event <- true
	})
	select {
	case <-event:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe with the correct position")
	}
}

func TestCachedOverride(t *testing.T) {
	channel := getChannel()
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Position: "wrong_position",
	})
	client.Subscribe(channel, subscription.RELIABLE, pdu.SubscribeBodyOpts{
		Filter: "SELECT COUNT(*) from `test`",
	})

	sub, err := client.GetSubscription(channel)
	if err != nil {
		t.Fatal(err)
	}
	eventC := make(chan bool)
	errC := make(chan error)
	sub.Once("subscribed", func(data interface{}) {
		eventC <- true
	})
	sub.Once("error", func(data interface{}) {
		errC <- data.(error)
	})

	defer client.Stop()
	go client.Start()

	select {
	case <-eventC:
	case err := <-errC:
		t.Fatal("Got error instead of 'subscribed'", err)
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe with new params")
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

	sub1, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Position: "wrong_position",
	})
	errC := make(chan bool)
	sub1.Once("error", func(interface{}) {
		errC <- true
	})
	go func() {
		select {
		case <-errC:
		case <-time.After(5 * time.Second):
			errorOccured = true
		}
		wg.Done()
	}()

	sub2, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Filter: "SELECT COUNT(*) FROM `test`",
	})
	event := make(chan bool)
	sub2.Once("subscribed", func(interface{}) {
		event <- true
	})
	go func() {
		select {
		case <-event:
		case <-time.After(5 * time.Second):
			errorOccured = true
		}
		wg.Done()
	}()

	sub3, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Position: "wrong_position",
	})
	sub3.Once("error", func(interface{}) {
		errC <- true
	})

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

	sub, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{})

	expected := []int{0, 1, 2}
	sub.On(subscription.EVENT_DATA, func(data interface{}) {
		actual, err := strconv.Atoi(string(data.(json.RawMessage)))
		if err != nil || expected[0] != actual {
			t.Fatal("Wrong message order or wrong message")
		}
		expected = expected[1:]
		wg.Done()
	})

	err = waitSubscribed(sub)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for i := 0; i < 3; i++ {
			client.Publish(channel, i)
		}
	}()
	wg.Wait()
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

	sub, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Filter: "select * from `" + channel + "` where test != 2",
	})

	expected := []string{"{\"test\":1}", "{\"test\":3}"}
	sub.On(subscription.EVENT_DATA, func(data interface{}) {
		msg := string(data.(json.RawMessage))
		if expected[0] != msg {
			err = errors.New("Wrong actiual data. Expected: " + expected[0] + " Actual: " + msg)
		}

		expected = expected[1:]
		wg.Done()
	})
	err = waitSubscribed(sub)
	if err != nil {
		t.Fatal(err)
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

	sub, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{})
	message := make(chan bool)
	expected := []string{"1", "2"}
	sub.On(subscription.EVENT_DATA, func(data interface{}) {
		if string(data.(json.RawMessage)) != expected[0] {
			t.Fatal("Wrong subscription message")
		}
		expected = expected[1:]
		message <- true
	})
	err = waitSubscribed(sub)
	if err != nil {
		t.Fatal(err)
	}

	go client.Publish(channel, 1)
	select {
	case <-message:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to subscribe")
	}

	// Drop connection
	client.conn.SetDeadline(time.Now())
	err = waitSubscribed(sub)
	if err != nil {
		t.Fatal(err)
	}

	go client.Publish(channel, 2)
	select {
	case <-message:
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

	sub, _ := client.Subscribe(channel, subscription.SIMPLE, pdu.SubscribeBodyOpts{})
	expected := []int{0, 1, 2}
	message := make(chan bool)
	sub.On(subscription.EVENT_DATA, func(data interface{}) {
		if len(expected) == 0 {
			t.Fatal("We got the message, but should not")
		}
		msg, _ := strconv.Atoi(string(data.(json.RawMessage)))
		if msg != expected[0] {
			t.Fatal("Wrong message order or wrong message")
		}
		expected = expected[1:]
		message <- true
	})
	err = waitSubscribed(sub)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for i := 0; i < 3; i++ {
			client.Publish(channel, i)
		}

	}()

	for i := 0; i < 3; i++ {
		select {
		case <-message:
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
	case <-message:
		t.Fatal("We are still subscribed")
	case <-time.After(1 * time.Second):
	}
}

func waitSubscribed(sub *subscription.Subscription) error {
	subscribedC := make(chan bool)
	errorC := make(chan error)
	sub.Once(subscription.EVENT_SUBSCRIBED, func(data interface{}) {
		subscribedC <- true
	})
	sub.Once(subscription.EVENT_SUBSCRIBE_ERROR, func(data interface{}) {
		errorC <- data.(error)
	})
	select {
	case <-subscribedC:
	case err := <-errorC:
		return err.(error)
	case <-time.After(5 * time.Second):
		return errors.New("Unable to subscribe")
	}

	return nil
}
