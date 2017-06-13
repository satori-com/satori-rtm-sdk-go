package main

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"os"
)

const (
	// Replace these values with your project's credentials
	// from Dev Portal (https://developer.satori.com/#/projects).
	//
	// Or use public one from https://www.satori.com/
	ENDPOINT = "YOUR_ENDPOINT"
	APP_KEY  = "YOUR_APPKEY"

	// Role and Secret are optional. Leaving these variables as is means no authentication.
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_SECRET"

	// Any channel name
	CHANNEL = "YOUR_CHANNEL_NAME"
)

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

	client, err := rtm.New(ENDPOINT, APP_KEY, options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// We just created a new client. The client is not connected yet and we will connect it later.
	// Before start the client you can specify additional callbacks to be able to react on events.
	// The full events list is specified here: https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-EVENTS
	//
	// Let's use only three of them: OnConnected, OnError and OnStop
	client.OnConnectedOnce(func() {
		fmt.Println("Connected to RTM!")
	})
	client.OnError(func(err rtm.RTMError) {
		fmt.Println(err.Reason)
	})

	// For synchronisation reason we will use typed channel (type string) to be able to collect all incoming messages
	data_c := make(chan string)

	// We create a subscription listener in order to receive callbacks
	// for incoming data, state changes and errors.
	//
	// The full list of available subscription events is here:
	// https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-SUBSCRIPTIONS
	listener := subscription.Listener{
		// In this callback we will process all incoming messages
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				data_c <- string(message)
			}
		},

		// Called when the subscription is established. Once we subscribed we create 2 demo animals that
		// will randomly move.
		// Be aware: All callbacks MUST NOT block the main thread. You should use go-routines in cases if you need
		// to wait for some data/events/etc.
		OnSubscribed: func(pdu.SubscribeOk) {
			fmt.Println("Successfully subscribed. Waiting for new messages in the '" + CHANNEL + "' channel")
			fmt.Println("Press CTRL + C to exit")
		},

		// Called when failed to subscribe
		OnSubscribeError: func(err pdu.SubscribeError) {
			fmt.Println("Failed to subscribe:", err.Error, err.Reason)
			os.Exit(2)
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

	// Now we have a subscription that will forward all incoming messages to our go-channel.
	for message := range data_c {
		fmt.Printf("Got message: %s\n", message)
	}
}
