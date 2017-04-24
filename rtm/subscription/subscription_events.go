package subscription

import (
	"encoding/json"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
)

const (
	EVENT_DATA               = "data"
	EVENT_SUBSCRIBED         = "subscribed"
	EVENT_UNSUBSCRIBED       = "unsubscribed"
	EVENT_POSITION           = "position"
	EVENT_INFO               = "info"
	EVENT_SUBSCRIBE_ERROR    = "subscribeError"
	EVENT_UNSUBSCRIBE_ERROR  = "unsubscribeError"
	EVENT_SUBSCRIPTION_ERROR = "subscriptionError"
)

func (s *Subscription) OnData(callback func(json.RawMessage)) interface{} {
	return s.On(EVENT_DATA, func(data interface{}) {
		callback(data.(json.RawMessage))
	})
}

func (s *Subscription) OnSubscribed(callback func(pdu.SubscribeOk)) interface{} {
	return s.On(EVENT_SUBSCRIBED, func(data interface{}) {
		callback(data.(pdu.SubscribeOk))
	})
}

func (s *Subscription) OnceSubscribed(callback func(pdu.SubscribeOk)) {
	s.Once(EVENT_SUBSCRIBED, func(data interface{}) {
		callback(data.(pdu.SubscribeOk))
	})
}

func (s *Subscription) OnUnsubscribed(callback func(data pdu.UnsubscribeBodyResponse)) interface{} {
	return s.On(EVENT_UNSUBSCRIBED, func(data interface{}) {
		callback(data.(pdu.UnsubscribeBodyResponse))
	})
}

func (s *Subscription) OnceUnsubscribed(callback func(data pdu.UnsubscribeBodyResponse)) {
	s.Once(EVENT_UNSUBSCRIBED, func(data interface{}) {
		callback(data.(pdu.UnsubscribeBodyResponse))
	})
}

func (s *Subscription) OnPosition(callback func(string)) interface{} {
	return s.On(EVENT_POSITION, func(data interface{}) {
		callback(data.(string))
	})
}

func (s *Subscription) OncePosition(callback func(string)) {
	s.Once(EVENT_POSITION, func(data interface{}) {
		callback(data.(string))
	})
}

func (s *Subscription) OnInfo(callback func(pdu.SubscriptionInfo)) interface{} {
	return s.On(EVENT_INFO, func(data interface{}) {
		callback(data.(pdu.SubscriptionInfo))
	})
}

func (s *Subscription) OnceInfo(callback func(pdu.SubscriptionInfo)) {
	s.Once(EVENT_INFO, func(data interface{}) {
		callback(data.(pdu.SubscriptionInfo))
	})
}

func (s *Subscription) OnSubscribeError(callback func(pdu.SubscribeError)) interface{} {
	return s.On(EVENT_SUBSCRIBE_ERROR, func(data interface{}) {
		callback(data.(pdu.SubscribeError))
	})
}

func (s *Subscription) OnceSubscribeError(callback func(pdu.SubscribeError)) {
	s.Once(EVENT_SUBSCRIBE_ERROR, func(data interface{}) {
		callback(data.(pdu.SubscribeError))
	})
}

func (s *Subscription) OnUnsubscribeError(callback func(pdu.UnsubscribeError)) interface{} {
	return s.On(EVENT_UNSUBSCRIBE_ERROR, func(data interface{}) {
		callback(data.(pdu.UnsubscribeError))
	})
}

func (s *Subscription) OnceUnsubscribeError(callback func(pdu.UnsubscribeError)) {
	s.Once(EVENT_UNSUBSCRIBE_ERROR, func(data interface{}) {
		callback(data.(pdu.UnsubscribeError))
	})
}

func (s *Subscription) OnSubscriptionError(callback func(pdu.SubscriptionError)) interface{} {
	return s.On(EVENT_SUBSCRIPTION_ERROR, func(data interface{}) {
		callback(data.(pdu.SubscriptionError))
	})
}

func (s *Subscription) OnceSubscriptionError(callback func(pdu.SubscriptionError)) {
	s.Once(EVENT_SUBSCRIPTION_ERROR, func(data interface{}) {
		callback(data.(pdu.SubscriptionError))
	})
}
