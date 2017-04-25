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
	"errors"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
)

const (
	STATE_UNSUBSCRIBED = 0
	STATE_SUBSCRIBED   = 1
)

var (
	ERROR_EMPTY_MODE = errors.New("Mode must be specified")
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

type Config struct {
	SubscriptionId string
	Opts           pdu.SubscribeBodyOpts
	Listener       Listener
	Mode           Mode
}

// Subscription mode struct. Check RELIABLE, SIMPLE and ADVANCED vars
type Mode struct {
	trackPosition bool
	fastForward   bool
}

// Subscription instance specification
type Subscription struct {
	state          int
	subscriptionId string
	mode           Mode
	position       string
	body           pdu.SubscribeBody
	listener       Listener
}

func New(config Config) *Subscription {
	s := &Subscription{}
	s.state = STATE_UNSUBSCRIBED
	s.mode = config.Mode
	s.subscriptionId = config.SubscriptionId
	s.position = ""

	opts := config.Opts
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

	s.listener = config.Listener

	return s
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

func (s *Subscription) ProcessSubscribe(data pdu.SubscribeOk) {
	s.trackPosition(data.Position)
	s.state = STATE_SUBSCRIBED
	s.body.Position = ""

	if s.listener.OnSubscribed != nil {
		s.listener.OnSubscribed(data)
	}

	logger.Info("Subscription '" + s.subscriptionId + "' is subscribed now")
}

func (s *Subscription) ProcessDisconnect() {
	s.markUnsubscribe(pdu.UnsubscribeBodyResponse{})
}

func (s *Subscription) ProcessInfo(data pdu.SubscriptionInfo) {
	s.trackPosition(data.Position)

	if s.listener.OnSubscriptionInfo != nil {
		s.listener.OnSubscriptionInfo(data)
	}

	logger.Warn("Falling behind for '" + s.subscriptionId + "'. Fast forward subscription")
}

func (s *Subscription) ProcessSubscribeError(data pdu.SubscribeError) {
	s.markUnsubscribe(pdu.UnsubscribeBodyResponse{})

	if s.listener.OnSubscribeError != nil {
		s.listener.OnSubscribeError(data)
	}

	logger.Warn("Error occured when subscribing to '" + s.subscriptionId + "'")
}

func (s *Subscription) ProcessSubscriptionError(data pdu.SubscriptionError) {
	s.trackPosition(data.Position)
	s.markUnsubscribe(pdu.UnsubscribeBodyResponse{})

	if s.listener.OnSubscriptionError != nil {
		s.listener.OnSubscriptionError(data)
	}
	logger.Warn("Subscription error for '" + s.subscriptionId + "'")
}

func (s *Subscription) ProcessUnsubscribe(data pdu.UnsubscribeBodyResponse) {
	s.markUnsubscribe(data)
}

func (s *Subscription) ProcessUnsubscribeError(data pdu.UnsubscribeError) {
	if s.listener.OnUnsubscribeError != nil {
		s.listener.OnUnsubscribeError(data)
	}
	logger.Warn("Error occured when unsubscribing from '" + s.subscriptionId + "'")
}

func (s *Subscription) ProcessData(data pdu.SubscriptionData) {
	s.trackPosition(data.Position)

	for _, message := range data.Messages {
		if s.listener.OnData != nil {
			s.listener.OnData(message)
		}
	}
}

func (s *Subscription) markUnsubscribe(data pdu.UnsubscribeBodyResponse) {
	if s.state == STATE_SUBSCRIBED {
		s.state = STATE_UNSUBSCRIBED

		if s.listener.OnUnsubscribed != nil {
			s.listener.OnUnsubscribed(data)
		}
	}
}

// Stores current position
func (s *Subscription) trackPosition(position string) {
	if s.mode.trackPosition {
		s.position = position
	}

	if s.listener.OnPosition != nil {
		s.listener.OnPosition(position)
	}
}

// Gets current subscription state
func (s *Subscription) GetState() int {
	return s.state
}

// Gets current subscription state
func (s *Subscription) GetSubscriptionId() string {
	return s.subscriptionId
}
