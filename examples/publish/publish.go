package main

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"os"
)

const (
	// Replace these values with your project's credentials
	// from Dev Portal (https://developer.satori.com/#/projects).
	ENDPOINT = "YOUR_ENDPOINT"
	APP_KEY  = "YOUR_APPKEY"

	// Role and Secret are optional. Leaving these variables as is means no authentication.
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_SECRET"

	// Any channel name
	CHANNEL = "YOUR_CHANNEL"
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

	client, err := rtm.New(ENDPOINT, APP_KEY, options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	connected := make(chan bool)
	// We just created a new client. The client is not connected yet and we will connect it later.
	// Before start the client you can specify additional callbacks to be able to react on events.
	// The full events list is specified here: https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-EVENTS
	client.OnConnected(func() {
		fmt.Println("Connected to RTM!")
		connected <- true
	})
	client.OnError(func(err rtm.RTMError) {
		fmt.Println("Error: " + err.Reason.Error())
		os.Exit(2)
	})

	client.Start()

	// Wait for connected state
	<-connected

	ack := <-client.PublishAck(CHANNEL, Animal{
		Who: "zebra",
		Where: [2]float32{34.134358, -118.321506},
	})

	if ack.Err != nil {
		fmt.Println("Unable to publish: " + ack.Err.Error())
	} else {
		fmt.Println("Message has been published successfully. Position: " + ack.Response.Position)
	}
}
