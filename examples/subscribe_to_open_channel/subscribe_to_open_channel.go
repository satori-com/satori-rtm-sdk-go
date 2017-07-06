package main

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"os"
)

const (
	ENDPOINT = "wss://open-data.api.satori.com"
	APP_KEY  = "YOUR_APPKEY"
	CHANNEL  = "OPEN_CHANNEL"
)

func main() {
	client, err := rtm.New(ENDPOINT, APP_KEY, rtm.Options{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client.OnConnected(func() {
		fmt.Println("Connected to Satori RTM!")
	})

	data_c := make(chan string)
	listener := subscription.Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				data_c <- string(message)
			}
		},
	}
	client.Subscribe(CHANNEL, subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)

	client.Start()

	for message := range data_c {
		fmt.Println("Got message:", message)
	}
}
