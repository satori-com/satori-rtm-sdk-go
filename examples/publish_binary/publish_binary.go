package main

import (
	"encoding/json"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"os"
	"time"
)

const (
	ENDPOINT        = "YOUR_ENDPOINT"
	APP_KEY         = "YOUR_APPKEY"
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_SECRET"

	// Any channel name
	CHANNEL = "videos"
)

type Frame struct {
	Id      int     `json:"frame_id"`
	Payload []uint8 `json:"payload"`
}

func main() {
	options := rtm.Options{}
	if ROLE_SECRET_KEY != "YOUR_SECRET" {
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

	connected := make(chan bool)
	client.OnConnected(func() {
		fmt.Println("Connected to Satori RTM!")
		connected <- true
	})
	client.OnError(func(err rtm.RTMError) {
		fmt.Println(err.Reason)
	})

	listener := subscription.Listener{
		OnData: func(data pdu.SubscriptionData) {
			for _, message := range data.Messages {
				var frame Frame

				// Try to unmarshal message to Frame struct
				err := json.Unmarshal(message, &frame)
				if err == nil {
					fmt.Printf("Got frame: %+v\n", frame)
				} else {
					fmt.Println("Failed to parse the incoming message:", string(message))
				}
			}
		},
		OnSubscribed: func(sok pdu.SubscribeOk) {
			fmt.Println("Subscribed to the channel:", sok.SubscriptionId)
		},
		OnSubscribeError: func(err pdu.SubscribeError) {
			fmt.Println("Failed to subscribe:", err.Error, err.Reason)
		},
	}

	client.Subscribe(CHANNEL, subscription.SIMPLE, pdu.SubscribeBodyOpts{}, listener)
	client.Start()

	<-connected
	frame_id := 0
	for {
		frame := Frame{
			Id:      frame_id,
			Payload: []uint8{12, 42, 0, 1, 255, 100},
		}
		response := <-client.PublishAck(CHANNEL, frame)
		if response.Err == nil {
			fmt.Printf("Frame is published: %+v\n", frame)
		} else {
			fmt.Println("Publish request failed: " + response.Err.Error())
		}
		frame_id++
		time.Sleep(2 * time.Second)
	}
}
