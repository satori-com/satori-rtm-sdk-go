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
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

const (
	// Replace these values with your project's credentials
	// from Dev Portal (https://developer.satori.com/#/projects).
	ENDPOINT = "YOUR_ENDPOINT"
	APP_KEY  = "YOUR_APPKEY"

	// Role and secret are optional. Leaving these variables as is means no authentication.
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_SECRET"

	// Any channel name
	CHANNEL = "animals"
)

// We define message struct that we are going to publish
type Animal struct {
	Who   string     `json:"who"`
	Where [2]float32 `json:"where"`
}

func main() {
	options := rtm.Options{}
	if ROLE_SECRET_KEY != "YOUR_SECRET" {
		// AuthProvider performs additional actions to authenticate the client
		//
		// Satori Docs: Authentication
		// https://www.satori.com/docs/using-satori/authentication
		authProvider := auth.New(ROLE, ROLE_SECRET_KEY)
		options = rtm.Options{
			AuthProvider: authProvider,
		}
	}

	fmt.Println("RTM client config:")
	fmt.Println("	endpoint =", ENDPOINT)
	fmt.Println("	appkey =", APP_KEY)
	fmt.Println("	authenticate? =", options.AuthProvider != nil)
	if options.AuthProvider != nil {
		fmt.Printf("	  (as \"%s\")\n", ROLE)
	}

	client, err := rtm.New(ENDPOINT, APP_KEY, options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// We just created a new client. The client is not connected yet and we will connect it later.
	// Before start the client you can specify additional callbacks to be able to react on events.
	// The full events list is specified here: https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-EVENTS
	//
	// Let's use only 4 of them: OnConnected, OnLeaveConnected, OnError and OnStop
	client.OnConnected(func() {
		fmt.Println("Connected to Satori RTM!")
	})
	client.OnLeaveConnected(func() {
		fmt.Println("Disconnected")
	})
	client.OnError(func(err rtm.RTMError) {
		fmt.Println(err.Reason)
	})
	client.OnStopOnce(func() {
		fmt.Println("Gracefully shutdown a program")
		os.Exit(0)
	})

	// We create a subscription listener in order to receive callbacks
	// for incoming data, state changes and errors.
	//
	// The full list of available subscription events is here:
	// https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-SUBSCRIPTIONS
	listener := subscription.Listener{
		// In this callback we will process all incoming messages
		// Be aware: All callbacks MUST NOT block the main thread. You should use go-routines in cases if you need
		// to wait for some data/events/etc.
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				var animal Animal

				// Try to unmarshal message to Animal struct
				err := json.Unmarshal(message, &animal)
				if err == nil {
					fmt.Printf("Got animal %s: %+v\n", animal.Who, animal.Where)
				} else {
					// We failed to convert the message to the Animal struct.
					fmt.Println("Failed to parse the incoming message:", string(message))
				}
			}
		},

		// Called when the subscription is established.
		OnSubscribed: func(sok pdu.SubscribeOk) {
			fmt.Println("Subscribed to the channel:", sok.SubscriptionId)
		},

		// Called when failed to subscribe
		OnSubscribeError: func(err pdu.SubscribeError) {
			fmt.Println("Failed to subscribe:", err.Error, err.Reason)
		},

		// Called when getting the unsolicited error
		OnSubscriptionError: func(err pdu.SubscriptionError) {
			fmt.Printf("Subscription failed. RTM sent the unsolicited error %s: %s\n", err.Error, err.Reason)
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

	// Catch "Ctrl + C" and gracefully shutdown the program
	sig_c := make(chan os.Signal, 1)
	signal.Notify(sig_c, os.Interrupt)
	go func() {
		<-sig_c
		// We got Ctrl + C. Stop the program
		client.Stop()
	}()
	fmt.Println("Press CTRL-C to exit")
	fmt.Println("====================")

	// At this point, the client may not yet be connected to Satori RTM.
	// If client is not connected then skip publishing.
	for {
		if client.IsConnected() {
			lat := 34.134358 + rand.Float32()/100
			lon := -118.321506 + rand.Float32()/100

			animal := Animal{
				Who:   "zebra",
				Where: [2]float32{lat, lon},
			}
			response := <-client.PublishAck(CHANNEL, animal)
			if response.Err == nil {
				// Publish is confirmed by Satori RTM.
				fmt.Printf("Animal is published: %+v\n", animal)
			} else {
				fmt.Println("Publish request failed: " + response.Err.Error())
			}

		}
		time.Sleep(2 * time.Second)
	}
}
