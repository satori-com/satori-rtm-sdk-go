// The example below shows how to publish a message with an acknowledgement to a channel and handle
// a publish response from the RTM service. In this example, an animal is published to the “animals” channel.
package main

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"os"
	"time"
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

	connected := make(chan bool)
	client.OnConnected(func() {
		fmt.Println("Connected to Satori RTM!")
		connected <- true
	})
	client.OnError(func(err rtm.RTMError) {
		fmt.Println("Failed to connect: " + err.Reason.Error())
		os.Exit(2)
	})

	client.Start()

	<-connected
	for {
		ack := <-client.PublishAck("animals", Animal{
			Who:   "zebra",
			Where: [2]float32{34.134358, -118.321506},
		})

		if ack.Err != nil {
			fmt.Println("Failed to publish: " + ack.Err.Error())
		} else {
			fmt.Println("Publish confirmed")
		}
		time.Sleep(1 * time.Second)
	}
}
