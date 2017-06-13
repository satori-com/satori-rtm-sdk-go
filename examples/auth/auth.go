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

	// Role and Secret.
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_ROLE_SECRET"
)

func main() {
	// AuthProvider performs additional actions to authenticate the client
	//
	// Satori Docs: Authentication
	// https://www.satori.com/docs/using-satori/authentication
	authProvider := auth.New(ROLE, ROLE_SECRET_KEY)
	client, err := rtm.New(ENDPOINT, APP_KEY, rtm.Options{
		AuthProvider: authProvider,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sync_c := make(chan bool)
	// We just created a new client. The client is not connected yet and we will connect it later.
	// Before start the client you can specify additional callbacks to be able to react on events.
	// The full events list is specified here: https://godoc.org/github.com/satori-com/satori-rtm-sdk-go/rtm#hdr-EVENTS
	client.OnAuthenticated(func() {
		fmt.Println("Successfully authenticated!")
		sync_c <- true
	})
	client.OnConnected(func() {
		fmt.Println("Connected to RTM!")
	})
	client.OnError(func(err rtm.RTMError) {
		if err.Code == rtm.ERROR_CODE_AUTHENTICATION {
			fmt.Println("Authentication error: " + err.Reason.Error())
		} else {
			fmt.Println("Error: " + err.Reason.Error())
		}
		sync_c <- true
	})

	client.Start()

	// Wait for events
	<-sync_c
}
