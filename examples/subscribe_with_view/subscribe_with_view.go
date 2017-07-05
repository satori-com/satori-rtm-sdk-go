// The example below shows how to subscribe with a view. In this example, the client subscribes with the view
// which leaves only zebras in the animals channel. We use  “zebras” as a subscription id. It can be used later
// to unsubscribe from this view.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"os"
)

const (
	ENDPOINT        = "YOUR_ENDPOINT"
	APP_KEY         = "YOUR_APPKEY"
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_SECRET"
)

type Animal struct {
	Who   string     `json:"who"`
	Where [2]float32 `json:"where"`
}

func main() {
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
	})
	client.OnError(func(err rtm.RTMError) {
		fmt.Println("Failed to connect: " + err.Reason.Error())
		os.Exit(2)
	})

	// For synchronisation reason we will use typed channel (type Animal) to be able to collect all incoming messages
	animals := make(chan Animal)

	listener := subscription.Listener{
		OnSubscribed: func(sok pdu.SubscribeOk) {
			fmt.Println("Subscribed to: " + sok.SubscriptionId)
		},
		OnUnsubscribed: func(response pdu.UnsubscribeBodyResponse) {
			fmt.Println("Unsubscribed from: " + response.SubscriptionId)
		},
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				var animal Animal

				// We assume, that the message in the channel has an Animal struct.
				err := json.Unmarshal(message, &animal)
				if err == nil {
					animals <- animal
				} else {
					fmt.Println("Failed to handle the incoming message:", string(message))
				}
			}
		},
		OnSubscribeError: func(err pdu.SubscribeError) {
			fmt.Printf("Failed to subscribe %s: %s\n", err.Error, err.Reason)
		},
		OnSubscriptionError: func(err pdu.SubscriptionError) {
			fmt.Printf("Subscription failed. RTM sent the unsolicited error %s: %s\n", err.Error, err.Reason)
		},
	}

	subscriptionId := "zebras"
	client.Subscribe(subscriptionId, subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Filter: "SELECT * FROM `animals` WHERE who = 'zebra'",
	}, listener)

	client.Start()

	// Now we have a subscription that will forward all incoming messages to our go-channel.
	for animal := range animals {
		fmt.Printf("Got animal %s: %+v\n", animal.Who, animal.Where)
	}
}
