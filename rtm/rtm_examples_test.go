package rtm_test

import (
	"errors"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"math/rand"
	"strconv"
	"time"
)

func ExampleRTM_Publish() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	for i := 0; i < 3; i++ {
		err := client.Publish("<your-channel>", time.Now().String()+" Message"+strconv.Itoa(i))
		if err != nil {
			logger.Error(err)
		}
	}
	logger.Info("Sent 3 messages")
}

func ExampleRTM_PublishAck_simple() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Publish 3 messages with Ack one after one
	for i := 0; i < 3; i++ {
		<-client.PublishAck("<your-channel>", time.Now().String()+" Ack message"+strconv.Itoa(i))
	}
	logger.Info("Sent 3 messages with Ack")
}

func ExampleRTM_PublishAck_processErrors() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}
	client.Start()

	// Publish 3 messages with Ack one after one
	for i := 0; i < 3; i++ {
		c := <-client.PublishAck("<your-channel>", time.Now().String()+" Ack message"+strconv.Itoa(i))
		if c.Err != nil {
			logger.Error(c.Err)
		} else {
			logger.Info("Got callback:", c.Response)
		}
	}
	logger.Info("Sent 3 messages with Ack")
}

func ExampleRTM_Search() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	// Wait for connected state
	connected := make(chan bool)
	client.On("enterConnected", func(interface{}) {
		connected <- true
	})
	<-connected

	// Make some channels to be able to find them
	client.Publish("tetete", "123")
	client.Publish("test", "123")
	<-client.PublishAck("t_1", "123")
	//Wait for the last message callback to be sure that all messages have been sent

	logger.Info("Search 't'")
	search := <-client.Search("t")
	for channel := range search.Channels {
		logger.Info("Found: " + channel)
	}
	logger.Info("Search done")
}

func ExampleRTM_Write_simple() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()
	client.Write("<your-channel>", "data111")
}

func ExampleRTM_Write_processErrors() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}

	client.Start()
	w := <-client.Write("<your-channel>", "data111")

	if w.Err != nil {
		logger.Error(w.Err)
	} else {
		logger.Info("Got callback: ", w.Response)
	}
}

func ExampleRTM_Read_simple() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})
	client.Start()

	<-client.Write("<your-channel>", "data111")
	// We can wait for ack callback to be sure the data is there

	r := <-client.Read("<your-channel>")
	fmt.Printf("Postition: %s; Data: %s\n", string(r.Response.Position), string(r.Response.Message))
}

func ExampleRTM_Read_processErrors() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}
	client.Start()

	w := <-client.Write("<your-channel>", "data111")
	// We can wait for ack callback to be sure the data is there
	if w.Err != nil {
		logger.Error(w.Err)
	}

	r := <-client.Read("<your-channel>")
	if r.Err != nil {
		logger.Error(r.Err)
	}
	fmt.Printf("Postition: %s; Data: %s\n", string(r.Response.Position), string(r.Response.Message))
}

func ExampleRTM_Subscribe() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, _ := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	sub, _ := client.Subscribe("<your-channel>", subscription.RELIABLE, pdu.SubscribeBodyOpts{
		Filter: "SELECT * FROM `test`",
		History: pdu.SubscribeHistory{
			Count: 1,
			Age:   10,
		},
	})

	client.Start()

	// Send random messages to the channel
	go func() {
		for {
			client.Publish("<your-channel>", strconv.Itoa(rand.Intn(10)))
			time.Sleep(200 * time.Millisecond)
		}
	}()

	for {
		message := <-sub.Data()
		fmt.Println("Got message: " + string(message))
	}
}

func ExampleRTM() {
	authProvider := auth.New("<your-role>", "<your-rolekey>")
	client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
		AuthProvider: authProvider,
	})

	if err != nil {
		logger.Fatal(err)
	}

	authEvent := make(chan bool)
	client.Once("authenticated", func(data interface{}) {
		logger.Info("Succesfully authenticated")
		authEvent <- true
	})
	client.On("error", func(err interface{}) {
		logger.Error(err.(error))
		authEvent <- true
	})

	client.Start()

	select {
	case <-authEvent:
	case <-time.After(5 * time.Second):
		logger.Error(errors.New("Unable to authenticate. Timeout"))
	}
}
