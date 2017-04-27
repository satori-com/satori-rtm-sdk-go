package subscription

import (
	"encoding/json"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
)

// You should create listener instance to define application functionality based on subscription state changes
// or subscription events.
// For example, you can define callback for when a channel receives a message, when the application
// subscribes or unsubscribes to a channel, or gets the errors.
//
// You should specify callbacks to subscribe to events.
type Listener struct {
	// Called when the client receives a message from the RTM Service that was published to the subscription.
	OnData func(json.RawMessage)

	// Called after successful subscription
	OnSubscribed func(pdu.SubscribeOk)

	// Called after successful unsubscription
	OnUnsubscribed func(data pdu.UnsubscribeBodyResponse)

	// Called on every received message that has Position param
	OnPosition func(string)

	// Called when the client receives a subscription info from the RTM Service.
	OnSubscriptionInfo func(pdu.SubscriptionInfo)

	// Called when the client receives a subscription error from the RTM Service.
	OnSubscribeError func(pdu.SubscribeError)

	// Called when the client receives error during unsubscribing.
	OnUnsubscribeError func(pdu.UnsubscribeError)

	// Called when the client receives a subscription error from the RTM Service.
	OnSubscriptionError func(pdu.SubscriptionError)
}
