// Declares all Protocol Data Unit (PDU) structs.
package pdu

import (
	"encoding/json"
)

const (
	CODE_BAD_REQUEST   = -1
	CODE_OK_REQUEST    = 0
	CODE_ERROR_REQUEST = 1
	CODE_DATA_REQUEST  = 2
)

type RTMQuery struct {
	Action string          `json:"action"`
	Body   json.RawMessage `json:"body,RawMessage"`
	Id     string          `json:"id,omitempty"`
}

type PublishBody struct {
	Channel string          `json:"channel"`
	Message json.RawMessage `json:"message,RawMessage"`
}

type PublishBodyResponse struct {
	Position string `json:"position"`
}

type WriteBody struct {
	Channel string          `json:"channel"`
	Message json.RawMessage `json:"message,RawMessage"`
}

type WriteBodyResponse struct {
	Position string `json:"position"`
}

type ReadBody struct {
	Channel  string `json:"channel"`
	Position string `json:"position,omitempty"`
}

type ReadBodyResponse struct {
	Message  json.RawMessage `json:"message"`
	Position string          `json:"position"`
}

type DeleteBody struct {
	Channel string `json:"channel"`
}

type DeleteBodyResponse struct {
	Position string `json:"position"`
}

type SearchBody struct {
	Prefix string `json:"prefix"`
}

type SearchBodyResponse struct {
	Channels []string `json:"channels"`
}

type SubscribeBody struct {
	Channel        string           `json:"channel,omitempty"`
	Force          bool             `json:"force,omitempty"`
	FastForward    bool             `json:"fast_forward,omitempty"`
	SubscriptionId string           `json:"subscription_id,omitempty"`
	Filter         string           `json:"filter,omitempty"`
	History        SubscribeHistory `json:"history,omitempty"`
	Period         int              `json:"period,omitempty"`
	Position       string           `json:"position,omitempty"`
}

type SubscribeBodyOpts struct {
	Filter   string           `json:"filter,omitempty"`
	History  SubscribeHistory `json:"history,omitempty"`
	Period   int              `json:"period,omitempty"`
	Position string           `json:"position"`
}

type SubscribeHistory struct {
	Count int `json:"count,omitempty"`
	Age   int `json:"age,omitempty"`
}

type SubscribeOk struct {
	Position       string `json:"position"`
	SubscriptionId string `json:"subscription_id"`
}

type SubscribeError struct {
	Error          string `json:"error"`
	Reason         string `json:"reason"`
	SubscriptionId string `json:"subscription_id"`
}

type SubscriptionInfo struct {
	Info           string `json:"info"`
	Reason         string `json:"reason"`
	SubscriptionId string `json:"subscription_id"`
	Position       string `json:"position"`
}

type SubscriptionError struct {
	Error          string `json:"error"`
	Reason         string `json:"reason"`
	Position       string `json:"position"`
	SubscriptionId string `json:"subscription_id"`
}

type SubscriptionData struct {
	Position       string            `json:"position"`
	Messages       []json.RawMessage `json:"messages"`
	SubscriptionId string            `json:"subscription_id"`
}

type UnsubscribeBody struct {
	SubscriptionId string `json:"subscription_id"`
}

type UnsubscribeBodyResponse struct {
	Position       string `json:"position"`
	SubscriptionId string `json:"subscription_id"`
}

type UnsubscribeError struct {
	Error          string `json:"error"`
	Reason         string `json:"reason"`
	SubscriptionId string `json:"subscription_id"`
}

type Error struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
}
