// RTM Subscription model.
//
// Model provides subscription modes. Check RELIABLE, SIMPLE and ADVANCED modes. Flags explanation:
//
//  trackPosition
//
// Tracks the stream position received from RTM. RTM includes the position
// parameter in responses to publish and subscribe requests and in subscription data messages.
// The SDK can attempt to resubscribe to the channel data stream from this position.
//
//  fastForward
//
// RTM fast-forwards the subscription when the SDK resubscribes to a channel.
//
package subscription

import (
	"encoding/json"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/observer"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
)

const (
	// Maximal queue of incoming messages. If the queue is full, no new messages will be read from the Socket.
	MAX_QUEUE = 10000

	STATE_UNSUBSCRIBED = 0
	STATE_SUBSCRIBED   = 1
)

var (
	/**
	 * The SDK tracks the position parameter and attempts to use that value when
	 * resubscribing. If the position parameter is expired, RTM fast-forwards
	 * to the earliest possible position value.
	 *
	 * This option may result in missed messages if the application has a slow connection
	 * to RTM and cannot keep up with the channel message data sent from RTM.
	 */
	RELIABLE = Mode{
		trackPosition: true,
		fastForward:   true,
	}

	/**
	 * The SDK does not track the position parameter received from RTM.
	 * Instead, when resubscribing following a reconnection, RTM fast-forwards to
	 * the earliest possible position parameter value.
	 *
	 * This option may result in missed messages during reconnection if the application has
	 * a slow connection to RTM and cannot keep up with the channel message stream sent from RTM.
	 */
	SIMPLE = Mode{
		trackPosition: false,
		fastForward:   true,
	}

	/**
	 * The JavaScript SDK tracks the position parameter and always uses that value when
	 * resubscribing.
	 *
	 * If the stream position is expired when the SDK attempts to resubscribe, RTM
	 * sends an expired_position error and unsubscribes.
	 *
	 * If the application has
	 * a slow connection to RTM and cannot keep up with the channel message data sent from RTM,
	 * RTM sends an out_of_sync error and unsubscribes.
	 */
	ADVANCED = Mode{
		trackPosition: true,
		fastForward:   false,
	}
)

// Subscription mode struct. Check RELIABLE, SIMPLE and ADVANCED vars
type Mode struct {
	trackPosition bool
	fastForward   bool
}

// Subscription instance specification
type Subscription struct {
	state          int
	subscriptionId string
	channel        chan json.RawMessage
	mode           Mode
	position       string
	body           pdu.SubscribeBody

	// Implements Observer behavior
	observer.Observer
}

// Creates new subscription instance
//
// Example:
//  sub, err := New("my-channel", RELIABLE, pdu.SubscribeBodyOpts{
//    Filter: "SELECT * FROM `test`",
//    History: pdu.SubscribeHistory{
//      Count: 1,
//      Age: 10,
//    },
//  })
//
//  sub2, err := New("my-simple-subscription", RELIABLE, pdu.SubscribeBodyOpts{})
//
func New(subscriptionId string, m Mode, opts pdu.SubscribeBodyOpts) *Subscription {
	s := &Subscription{
		Observer: observer.New(),
	}
	s.state = STATE_UNSUBSCRIBED
	s.mode = m
	s.channel = make(chan json.RawMessage, MAX_QUEUE)
	s.subscriptionId = subscriptionId
	s.position = ""

	s.body.Filter = opts.Filter
	s.body.History = opts.History
	s.body.Period = opts.Period
	s.body.Position = opts.Position

	s.body.FastForward = s.mode.fastForward

	if len(s.body.Filter) > 0 {
		s.body.SubscriptionId = s.subscriptionId
	} else {
		s.body.Channel = s.subscriptionId
	}

	return s
}

// Gets data from the subscription
func (s *Subscription) Data() <-chan json.RawMessage {
	return s.channel
}

// Gets PDU to subscribe
func (s *Subscription) SubscribePdu() pdu.RTMQuery {
	query := pdu.RTMQuery{
		Action: "rtm/subscribe",
	}

	if len(s.position) != 0 {
		s.body.Position = s.position
	}

	// Always use force flag to avoid resubscribing errors
	s.body.Force = true

	query.Body, _ = json.Marshal(s.body)

	return query
}

// Gets PDU to unsubscribe
func (s *Subscription) UnsubscribePdu() pdu.RTMQuery {
	query := pdu.RTMQuery{
		Action: "rtm/unsubscribe",
	}
	query.Body, _ = json.Marshal(pdu.UnsubscribeBody{
		SubscriptionId: s.subscriptionId,
	})

	return query
}

func (s *Subscription) OnSubscribe(data pdu.SubscribeOk) {
	s.trackPosition(data.Position)
	s.state = STATE_SUBSCRIBED
	s.body.Position = ""
	s.Fire("subscribed", data)

	logger.Info("Subscription '" + s.subscriptionId + "' is subscribed now")
}

func (s *Subscription) OnDisconnect() {
	s.markUnsubscribe()
}

func (s *Subscription) OnInfo(data pdu.SubscriptionInfo) {
	s.trackPosition(data.Position)
	s.Fire("info", data)

	logger.Warn("Falling behind for '" + s.subscriptionId + "'. Fast forward subscription")
}

func (s *Subscription) OnSubscribeError(data pdu.SubscribeError) {
	s.markUnsubscribe()
	s.Fire("subscribeError", data)

	logger.Warn("Error occured when subscribing to '" + s.subscriptionId + "'")
}

func (s *Subscription) OnSubscriptionError(data pdu.SubscriptionError) {
	s.trackPosition(data.Position)
	s.markUnsubscribe()
	s.Fire("subscriptionError", data)

	logger.Warn("Subscription error for '" + s.subscriptionId + "'")
}

func (s *Subscription) OnUnsubscribeError(data pdu.UnsubscribeError) {
	s.Fire("unsubscribeError", data)
	logger.Warn("Error occured when unsubscribing from '" + s.subscriptionId + "'")
}

func (s *Subscription) OnData(data pdu.SubscriptionData) {
	for _, message := range data.Messages {
		// @Warning
		// Sending messages can be locked if the user reads slower than the messages rps
		// "Channel" is a buffered channel with the MAX_QUEUE size
		s.channel <- message

		// Allow user to use several ways how to get data. Via channel of using Observer event
		s.Fire("data", message)
	}
}

// Marks current subscription as "unsubscribed"
func (s *Subscription) markUnsubscribe() {
	if s.state == STATE_SUBSCRIBED {
		s.state = STATE_UNSUBSCRIBED
		s.Fire("unsubscribed", nil)
	}
}

// Stores current position
func (s *Subscription) trackPosition(position string) {
	if s.mode.trackPosition {
		s.position = position
	}

	s.Fire("position", position)
}

// Gets current subscription state
func (s *Subscription) GetState() int {
	return s.state
}

// Gets current subscription state
func (s *Subscription) GetSubscriptionId() string {
	return s.subscriptionId
}
