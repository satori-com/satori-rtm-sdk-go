// The example below shows how to establish several subscriptions with different views to the same channel.
// Each subscription has a different view on the same channel. You should use the same instance of the RTM client
// for both subscriptions. In this example, the client subscribes with the following views:
// - Leave only zebras in the animals channel.
// - Count messages by animal kind. RTM view aggregates and delivers counts every second (default period).
//
// We use “zebras” as a subscription id for the first subscription and “stats” as a subscription id
// for the second subscription. You can use these subscription ids to distinguish incoming messages,
// state changes and errors in the subscription callbacks if you share the same callbacks between the subscriptions.
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
	ENDPOINT        = "YOUR_ENDPOINT"
	APP_KEY         = "YOUR_APPKEY"
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_ROLE_SECRET"
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
	data_c := make(chan string)

	listener := subscription.Listener{
		OnSubscribed: func(sok pdu.SubscribeOk) {
			fmt.Println("Subscribed to: " + sok.SubscriptionId)
		},
		OnUnsubscribed: func(response pdu.UnsubscribeBodyResponse) {
			fmt.Println("Unsubscribed from: " + response.SubscriptionId)
		},
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				if data.SubscriptionId == "zebras" {
					data_c <- fmt.Sprintf("Got a zebra: %s", message)
				} else {
					data_c <- fmt.Sprintf("Got a count: %s", message)
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

	client.Subscribe("zebras", subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Filter: "SELECT * FROM `animals` WHERE who = 'zebra'",
	}, listener)

	client.Subscribe("stats", subscription.SIMPLE, pdu.SubscribeBodyOpts{
		Filter: "SELECT count(*) as count, who FROM `animals` GROUP BY who",
	}, listener)

	client.Start()

	for data := range data_c {
		fmt.Println(data)
	}
}
