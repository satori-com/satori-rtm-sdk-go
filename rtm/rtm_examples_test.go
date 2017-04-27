package rtm_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"math"
	"math/rand"
	"time"
)

func ExampleRTM_Publish() {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	client.Publish("<your-channel>", Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})
	logger.Info("Message has been sent")
}

func ExampleRTM_Publish_types() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	var i int = 42
	client.Publish("<your-channel>", i)

	var ui8 uint8 = 1
	client.Publish("<your-channel>", ui8)

	var f32 float32 = 1.2345
	client.Publish("<your-channel>", f32)

	var f64 float64 = math.Pi
	client.Publish("<your-channel>", f64)

	var b bool = true
	client.Publish("<your-channel>", b)

	var str string = "Hello world!"
	client.Publish("<your-channel>", str)

	// Null message
	client.Publish("<your-channel>", nil)
}

func ExampleRTM_PublishAck_simple() {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	response := <-client.PublishAck("<your-channel>", Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})
	logger.Info(response)
}

func ExampleRTM_PublishAck_processErrors() {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}
	client.OnError(func(err rtm.RTMError) {
		logger.Error(err.Reason)
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	c := <-client.PublishAck("<your-channel>", Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})
	if c.Err != nil {
		logger.Error(c.Err)
	} else {
		logger.Info("Got callback:", c.Response)
	}
}

func ExampleRTM_Search() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	// Make some channels to be able to find them
	client.Publish("tetete", "123")
	client.Publish("test", "123")
	<-client.PublishAck("t_1", "123")
	//Wait for the last message callback to be sure that all messages have been sent

	logger.Info("Search 't'")
	search := <-client.Search("t")
	for channel := range search.Channels {
		logger.Info("Found: " + channel)
	}
	logger.Info("Search done")
}

func ExampleRTM_Write_simple() {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	client.Write("<your-channel>", Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})
}

func ExampleRTM_Write_processErrors() {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}
	client.OnError(func(err rtm.RTMError) {
		logger.Error(err.Reason)
	})

	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	w := <-client.Write("<your-channel>", Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})

	if w.Err != nil {
		logger.Error(w.Err)
	} else {
		logger.Info("Got callback: ", w.Response)
	}
}

func ExampleRTM_Read_simple() {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	// Write message and wait for Ack to be sure that the message is there
	<-client.Write("<your-channel>", Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})

	r := <-client.Read("<your-channel>")
	fmt.Printf("Postition: %s; Data: %s\n", string(r.Response.Position), string(r.Response.Message))
}

func ExampleRTM_Read_processErrors() {
	type Message struct {
		Who   string    `json:"who"`
		Where []float32 `json:"where"`
	}

	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}
	client.OnError(func(err rtm.RTMError) {
		logger.Error(err.Reason)
	})
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	// Write message and wait for Ack to be sure that the message is there
	w := <-client.Write("<your-channel>", Message{
		Who:   "zebra",
		Where: []float32{34.134358, -118.321506},
	})
	if w.Err != nil {
		logger.Error(w.Err)
	}

	r := <-client.Read("<your-channel>")
	if r.Err != nil {
		logger.Error(r.Err)
	} else {
		fmt.Printf("Postition: %s; Data: %s\n", string(r.Response.Position), string(r.Response.Message))
	}
}

func ExampleRTM_Subscribe() {
	type Message struct {
		Id int
	}
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	listener := subscription.Listener{
		OnData: func(message json.RawMessage) {
			logger.Info(string(message))
		},
	}
	client.Subscribe(
		"<your-channel>",
		subscription.RELIABLE,
		pdu.SubscribeBodyOpts{
			Filter: "SELECT * FROM `test`",
			History: pdu.SubscribeHistory{
				Count: 1,
				Age:   10,
			},
		},
		listener,
	)

	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	// Send random messages to the channel
	go func() {
		for {
			client.Publish("<your-channel>", Message{
				Id: rand.Intn(10),
			})
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// Exit after 10 seconds
	<-time.After(10 * time.Second)
}

func ExampleRTM_Subscribe_processErrors() {
	type Message struct {
		Id int
	}
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	listener := subscription.Listener{
		OnData: func(message json.RawMessage) {
			// Got message
			logger.Info(string(message))
		},
		OnSubscribed: func(sok pdu.SubscribeOk) {
			// Successfully subscribed
			logger.Info(sok)
		},
		OnSubscriptionInfo: func(info pdu.SubscriptionInfo) {
			// Got "subscription/info" from RTM
			logger.Warn(info)
		},
		OnSubscriptionError: func(err pdu.SubscriptionError) {
			// Got "subscription/error" from RTM
			logger.Error(errors.New(err.Error + "; " + err.Reason))
		},
		OnUnsubscribed: func(unsub pdu.UnsubscribeBodyResponse) {
			// Successfully unsubscribed
			logger.Info(unsub)
		},
	}

	client.Subscribe(
		"<your-channel>",
		subscription.RELIABLE,
		pdu.SubscribeBodyOpts{
			Filter: "SELECT * FROM `test`",
			History: pdu.SubscribeHistory{
				Count: 1,
				Age:   10,
			},
		},
		listener,
	)
	client.Start()

	// Wait for client is connected
	connected := make(chan bool)
	client.OnceConnected(func() {
		connected <- true
	})
	<-connected

	// Send random messages to the channel
	go func() {
		for {
			client.Publish("<your-channel>", Message{
				Id: rand.Intn(10),
			})
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// Exit after 10 seconds
	<-time.After(10 * time.Second)
}

func ExampleRTM() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}

	authEvent := make(chan bool)
	client.OnceAuthenticated(func() {
		logger.Info("Succesfully authenticated")
		authEvent <- true
	})
	client.OnError(func(err rtm.RTMError) {
		logger.Error(err.Reason)
		authEvent <- true
	})

	client.Start()

	select {
	case <-authEvent:
	case <-time.After(5 * time.Second):
		logger.Error(errors.New("Unable to authenticate. Timeout"))
	}
}
