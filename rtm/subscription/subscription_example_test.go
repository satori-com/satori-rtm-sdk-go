package subscription

import (
	"encoding/json"
	"errors"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
)

// Creates new listener instance and specifies several callbacks
func ExampleNewListener() {
	listener := NewListener()
	listener.OnData = func(message json.RawMessage) {
		// Got message
		logger.Info(string(message))
	}
	listener.OnSubscribeError = func(err pdu.SubscribeError) {
		// Subscribe error
		logger.Error(errors.New(err.Error + "; " + err.Reason))
	}
	listener.OnSubscribed = func(sok pdu.SubscribeOk) {
		logger.Info("Successfully subscribed from position: " + sok.Position)
	}
}