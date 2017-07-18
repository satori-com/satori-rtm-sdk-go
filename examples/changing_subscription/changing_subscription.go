package main

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"os"
	"sync"
)

const (
	ENDPOINT        = "YOUR_ENDPOINT"
	APP_KEY         = "YOUR_APPKEY"
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_SECRET"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	authProvider := auth.New(ROLE, ROLE_SECRET_KEY)
	options := rtm.Options{
		AuthProvider: authProvider,
	}

	client, err := rtm.New(ENDPOINT, APP_KEY, options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client.OnConnected(func() {
		fmt.Println("Connected to Satori RTM!")
		wg.Done()
	})
	client.OnError(func(err rtm.RTMError) {
		fmt.Println("Failed to connect: " + err.Reason.Error())
		os.Exit(2)
	})
	client.Start()

	data_c := make(chan string)
	listener := subscription.Listener{
		OnSubscribed: func(sok pdu.SubscribeOk) {
			fmt.Println("Subscribed to: " + sok.SubscriptionId)
		},
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				data_c <- string(message)
			}
		},
		OnSubscribeError: func(err pdu.SubscribeError) {
			fmt.Printf("Failed to subscribe %s: %s\n", err.Error, err.Reason)
		},
	}

	// Subscribe to the "animals" channel and wait OnSubscribed event
	client.Subscribe("animals", subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)
	wg.Wait()

	// Resubscribe to the channel, but using Filter
	client.Subscribe("animals", subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Filter: "select * from animals where who like 'z%'",
	}, listener)

	// Now we have a subscription that will forward all incoming messages to our go-channel.
	for message := range data_c {
		fmt.Println("Got message:", message)
	}
}
