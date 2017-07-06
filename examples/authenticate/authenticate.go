// The example below shows how to authenticate the RTM client and handle authentication errors.
// Replace the RTM credential placeholders with the values obtained from your project in Dev Portal.
// Note, that authentication is required only if you publish, subscribe or perform other operations with a
// restricted channel. Channel permissions are configured in Dev Portal.
package main

import (
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"os"
	"sync"
)

const (
	ENDPOINT        = "YOUR_ENDPOINT"
	APP_KEY         = "YOUR_APPKEY"
	ROLE            = "YOUR_ROLE"
	ROLE_SECRET_KEY = "YOUR_ROLE_SECRET"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	authProvider := auth.New(ROLE, ROLE_SECRET_KEY)
	client, err := rtm.New(ENDPOINT, APP_KEY, rtm.Options{
		AuthProvider: authProvider,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client.OnConnected(func() {
		fmt.Println("Connected to Satori RTM and authenticated as " + ROLE)
		wg.Done()
	})
	client.OnError(func(err rtm.RTMError) {
		if err.Code == rtm.ERROR_CODE_AUTHENTICATION {
			fmt.Println("Authentication error: " + err.Reason.Error())
		} else {
			fmt.Println("Error: " + err.Reason.Error())
		}
		wg.Done()
	})

	client.Start()
	wg.Wait()
}
