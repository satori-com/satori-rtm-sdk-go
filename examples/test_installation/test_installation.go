package main

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"os"
	"sync"
)

const (
	ENDPOINT = "YOUR_ENDPOINT"
	APP_KEY  = "YOUR_APPKEY"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	client, err := rtm.New(ENDPOINT, APP_KEY, rtm.Options{})
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
		wg.Done()
	})

	client.Start()
	wg.Wait()
}
