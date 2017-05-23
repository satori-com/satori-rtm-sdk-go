// For tutorial purposes, we subscribe to the same channel that we publish a
// message to. So we receive our own published message. This allows end-to-end
// illustration of data flow with just a single client.
//
// ==== Before you start ====
// Make sure that you have `go` installed: https://golang.org/doc/install
// Also please make sure that your Go-Workspace is properly configured: https://golang.org/doc/code.html#Workspaces
// and your GOPATH environment variable is point to your Go-Workspace.
//
// Hint: Run `make run` from the console to make auto-configuration for the Go-Workspace
//
//   $ git clone git@github.com:satori-com/satori-rtm-sdk-go.git
//   $ cd satori-rtm-sdk-go/tutorial/
//   $ make run
//
package main

import (
	"encoding/json"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"time"
)

const (
	// Replace these values with your project's credentials
	// from Dev Portal (https://developer.satori.com/#/projects).
	ENDPOINT = "<ENDPOINT>"
	APP_KEY  = "<APP_KEY>"

	// Role and Secret are optional. Setting these to empty mean no authentication.
	ROLE            = "<ROLE>"
	ROLE_SECRET_KEY = "<ROLE_SECRET>"

	// Any channel name
	CHANNEL = "animal_sightings"
)

// We define message struct that we are going to publish
type Animal struct {
	Who   string    `json:"name"`
	Where []float64 `json:"where"`
}

func main() {
	options := rtm.Options{}
	if ROLE != "" && ROLE_SECRET_KEY != "" {
		// AuthProvider performs additional actions to authenticate the client
		//
		// Satori Docs: Authentication
		// https://www.satori.com/docs/using-satori/authentication
		authProvider := auth.New(ROLE, ROLE_SECRET_KEY)
		options = rtm.Options{
			AuthProvider: authProvider,
		}
	}

	client, err := rtm.New(ENDPOINT, APP_KEY, options)
	if err != nil {
		logger.Fatal(err)
	}

	// We just created a new client. The client is not connected yet and we will connect it later.
	// Before start the client you can specify additional callbacks to be able to react on events.
	// The full events list is specified here: https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-EVENTS
	//
	// Let's use only two of them: OnConnected and OnError.
	client.OnConnectedOnce(func() {
		logger.Info("Connected to RTM!")
	})
	client.OnError(func(err rtm.RTMError) {
		logger.Error(err.Reason)
	})

	// For synchronisation reason we will use bool channel to be able to wait for an incoming message.
	data_c := make(chan bool, 1)

	// We create a subscription listener in order to receive callbacks
	// for incoming data, state changes and errors.
	//
	// The full list of available subscription events is here:
	// https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-SUBSCRIPTIONS
	listener := subscription.Listener{
		// In this callback we will process all incoming messages
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				var animal Animal
				json.Unmarshal(message, &animal)
				logger.Info("Got message: " + fmt.Sprintf("%+v", animal))
			}
			data_c <- true
		},

		// Called when the subscription is established. Once we subscribed we publish one message.
		// Be aware: All callbacks MUST NOT block the main thread. You should use go-routines in cases if you need
		// to wait for some data/events/etc.
		OnSubscribed: func(pdu.SubscribeOk) {
			client.Publish(CHANNEL, Animal{
				Who:   "Zebra",
				Where: []float64{34.134358, -118.321506}},
			)
		},

		// Called when failed to subscribe
		OnSubscriptionError: func(err pdu.SubscriptionError) {
			logger.Warn("Failed to subscribe: ", err.Error, err.Reason)
		},
	}

	// Create Subscription. You should specify SubscriptionID, Subscription Mode, Options and Callbacks listener
	// To get more information about the Subscription modes follow the link:
	// https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm/subscription#pkg-variables
	//
	// Satori Docs: Subscribing
	// https://www.satori.com/docs/using-satori/subscribing
	client.Subscribe(CHANNEL, subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)

	// Now we start the client. After that client will establish connection to the Satori endpoint,
	// pass the authentication and subscribe to the channel.
	// Our callbacks will catch all these events.
	client.Start()

	// Wait for a message in the data_c go channel to put our main thread to sleep
	// (block).  As soon as SDK invokes our OnData callback
	// (from another go-routine), we unblock the main thread.
	// To avoid indefinite hang in case of a
	// failure, the wait is limited to 10 second timeout.
	select {
	case <-data_c:
	case <-time.After(10 * time.Second):
		logger.Warn("Timeout: Unable to get the message")
	}

	// Stop the client with calling STOPPED callbacks and disconnect from RTM
	client.Stop()

	logger.Info("Done. Bye!")
}
