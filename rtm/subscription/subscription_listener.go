package subscription

import (
	"encoding/json"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
)

type Listener struct {
	OnData              func(json.RawMessage)
	OnSubscribed        func(pdu.SubscribeOk)
	OnUnsubscribed      func(data pdu.UnsubscribeBodyResponse)
	OnPosition          func(string)
	OnSubscriptionInfo  func(pdu.SubscriptionInfo)
	OnSubscribeError    func(pdu.SubscribeError)
	OnUnsubscribeError  func(pdu.UnsubscribeError)
	OnSubscriptionError func(pdu.SubscriptionError)
}

func NewListener() Listener {
	return Listener{}
}
