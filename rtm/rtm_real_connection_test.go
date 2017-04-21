package rtm

import (
	"encoding/json"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"strings"
	"testing"
	"time"
)

func TestWrongEndpoint(t *testing.T) {
	client, _ := New("ws://wrong-host-name.www", "123", Options{})
	event := make(chan bool)
	client.Once("error", func(err interface{}) {
		if !strings.Contains(err.(error).Error(), "no such host") {
			t.Fatal("Wrong error returned")
		}
		event <- true
	})

	defer client.Stop()
	go client.Start()

	select {
	case <-event:
	case <-time.After(5 * time.Second):
		t.Fatal("Cannot get endpoint error")
	}
}

func TestWrongAuth(t *testing.T) {
	credentials, err := getCredentials()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}

	authProvider := auth.New("non-existing-role-name", "wrong-secret-key")
	client, _ := New(credentials.Endpoint, credentials.AppKey, Options{
		AuthProvider: authProvider,
	})
	event := make(chan bool)
	client.Once("error", func(err interface{}) {
		// Try to convert to AuthError
		var conv pdu.Error
		err = json.Unmarshal([]byte(err.(error).Error()), &conv)
		if err == nil {
			if conv.Error != "authentication_failed" {
				t.Fatal("Wrong error type returned")
			}
		} else {
			t.Fatal("Unable to cast to PDU error")
		}

		event <- true
	})

	defer client.Stop()
	go client.Start()

	select {
	case <-event:
	case <-time.After(5 * time.Second):
		t.Fatal("Cannot get endpoint error")
	}
}

func TestClientDisconnect(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	event := make(chan bool)
	// Multiple event handler
	client.On("enterConnected", func(interface{}) {
		event <- true
	})
	defer client.Stop()
	go client.Start()

	select {
	case <-event:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to connect")
	}

	// Drop connection
	client.conn.SetDeadline(time.Now())

	select {
	case <-event:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to connect after drop connection")
	}
}
