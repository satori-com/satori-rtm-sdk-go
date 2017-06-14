// RTM client.
//
// Use the RTM to create a client instance from which you can
// publish messages and subscribe to channels. Create separate
// Subscription objects for each channel to which you want to subscribe.
//
//  // Create a client
//  client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{})
//
// A state machine for the client defines the status of the client instance.
// A client instance can be in one of the following states:
// 	STATE_STOPPED
// 	STATE_CONNECTING
// 	STATE_CONNECTED
// 	STATE_AWAITING
//
// EVENTS
//
// RTM Client allows to subscribe to the state changing events. An event occurs when the client
// enters or leaves a state.
//
// Use client.On<Event> function to continuously processing events or On<Event>Once for one-time processing. Example:
//
//   client.OnConnected(func() {
//     logger.Info("Connected")
//   })
//
// You can use the following event consts to subscribe on. RTM client has:
//
//  // STATE_STOPPED state when creating a new instance or when calling client.Stop()
//  OnStopped, OnStoppedOnce, OnLeaveStopped, OnLeaveStoppedOnce
//
//  // STATE_CONNECTING state when starting connecting to endpoint
//  OnConnecting, OnConnectingOnce, OnLeaveConnecting, OnLeaveConnectingOnce
//
//  // STATE_CONNECTED state when client established connection and ready to publish/read/write/etc
//  OnConnected, OnConnectedOnce, OnLeaveConnected, OnLeaveConnectedOnce
//
//  // STATE_AWAITING state when network connection is broken or unexpectedly closed
//  OnAwaiting, OnAwaitingOnce, OnLeaveAwaiting, OnLeaveAwaitingOnce
//
// RTM Client allows to use Event-Based model for other Events. Example:
//
//   client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{})
//   if err != nil {
//     logger.Fatal(err)
//   }
//   client.OnError(func(err RTMError) {
//     if err.Code == ERROR_CODE_TRANSPORT {
//       // Socket connection is broken
//       logger.Err(err.Reason)
//     }
//   })
//
//   client.OnAuthenticatedOnce(func() {
//     logger.Info("Successfully authenticated")
//   })
//
// Or use multiple event handlers for the same event
//
//   client.OnError(func(err RTMError) {
//     if err.Code == ERROR_CODE_TRANSPORT {
//       logger.Err(err.Reason)
//     }
//   })
//
//   client.OnError(func(err RTMError) {
//     if err.Code == ERROR_CODE_AUTHENTICATION {
//       // Make some actions if we failed to authenticate
//       logger.Warn(err.Reason.Error())
//     }
//   })
//
// List of available events:
//
//   OnStart, OnStartOnce, OnStop, OnStopOnce, OnOpen, OnOpenOnce, OnError, OnErrorOnce,
//   OnDataError, OnDataErrorOnce, OnAuthenticated, OnAuthenticatedOnce
//
// ERRORS
//
// When subscribing to the OnError event, callback function will always get RTMError type.
// RTMError contains two fields:
//
//   Code
//   Reason
//
// Error "Code" can be compared with the following types to determine type of Error
//
//    ERROR_CODE_APPLICATION    - Application layer errors. Occur when creating new client with wrong params, getting
//                                error response from RTM, etc
//    ERROR_CODE_TRANSPORT      - Transport layer errors. Occur if RTM client failed to send/read message
//                                using connection
//    ERROR_CODE_PDU            - Occur when receiving Error PDU response from RTM
//    ERROR_CODE_INVALID_JSON   - Occur if you try to send wrong json PDU. E.g. when you try to send invalid json.RawMessage
//    ERROR_CODE_AUTHENTICATION - All authentication-related errors
//
// Error Code example:
//   client.OnError(func(err RTMError) {
//     if err.Code == ERROR_CODE_AUTHENTICATION {
//       // Auth issue
//       logger.Error(err.Reason)
//     }
//   })
//
//
// SUBSCRIPTIONS
//
// RTM client allows to subscribe to channels.
//
//   err := client.Subscribe("<your-channel>", subscription.RELIABLE, pdu.SubscribeBodyOpts{}, subscription.Listener{})
//
// Each subscription has 3 available subscription modes:
//
//   RELIABLE
//   SIMPLE
//   ADVANCED
//
// Check the rtm/subscription sub-package to get more information about the modes.
//
// A subscription has an ability to specify listeners on the following Events:
//
//   OnData, OnSubscribed, OnUnsubscribed, OnPosition, OnSubscriptionInfo,
//   OnSubscribeError, OnUnsubscribeError, OnSubscriptionError
//
// You should specify listeners when creating a new subscription. Example:
//
//   listener := subscription.Listener{
//     OnSubscribed: func(sok pdu.SubscribeOk) {
//       // Successfully subscribed
//       logger.Info(sok)
//     },
//     OnSubscriptionError: func(err pdu.SubscriptionError) {
//       // Got "subscription error" from RTM
//       logger.Error(errors.New(err.Error + "; " + err.Reason))
//     },
//   }
//   err := client.Subscribe("<your-channel>", subscription.RELIABLE, pdu.SubscribeBodyOpts{}, listener)
//
//
// Set OnData callback to get subscription messages
//
//   // Example: Get messages and cast them to Message type
//   type Point struct {
//     Who   string    `json:"who"`
//     Where []float32 `json:"where"`
//   }
//
//   listener := subscription.Listener{
//     OnData: func(data pdu.SubscriptionData) {
//       for _, message := range data.Messages {
//         var point Point
//         json.Unmarshal(message, &point)
//         logger.Info(point.Who, point.Where)
//       }
//     },
//   }
//   sub, err := client.Subscribe("<your-channel>", subscription.RELIABLE, pdu.SubscribeBodyOpts{}, listener)
//
// AUTH
//
// You can specify role to get role-based permissions (E.g. get an access to Subscribe/Publish to some channels)
// Follow the link to get more information: https://www.satori.com/docs/using-satori/authentication
//
// Use Auth sub-package to authenticate using Role/SecretKey
//
//   authProvider := auth.New("<your-role>", "<your-rolekey>")
//   client, err := rtm.New("<your-endpoint>", "<your-appkey>", rtm.Options{
//     AuthProvider: authProvider,
//   })
//
package rtm

import (
	"encoding/json"
	"errors"
	"github.com/satori-com/satori-rtm-sdk-go/fsm"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/observer"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/connection"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"regexp"
)

const (
	API_VER                = "v2"
	MAX_RECONNECT_TIME_SEC = 120

	STATE_STOPPED    = "stopped"
	STATE_CONNECTING = "connecting"
	STATE_CONNECTED  = "connected"
	STATE_AWAITING   = "awaiting"

	ACK   = true
	NOACK = false
)

var (
	ERROR_SUBSCRIPTION_NOT_FOUND = errors.New("Subscription not found")
	ERROR_UNSUPPORTED_TYPE       = errors.New("Unable to send data. Unsupported type")
	ERROR_EMPTY_ENDPOINT         = errors.New("Endpoint is empty")
	ERROR_EMPTY_APP_KEY          = errors.New("App key is empty")
	ERROR_NOT_CONNECTED          = errors.New("Not connected")
)

type RTM struct {
	endpoint string
	appKey   string
	opts     Options

	conn           *connection.Connection
	reconnectCount int
	subscriptions  subscriptionsType

	fsm *fsm.FSM

	// Implements Observer behavior
	observer.Observer
}

// Creates a new RTM client instance.
// "endpoint" and "appkey" are mandatory fields and cannot be empty.
//
// You should run Start() after creating a new instance to establish connection to RTM
func New(endpoint, appkey string, opts Options) (*RTM, error) {
	logger.Info("Creating new RTM object")

	if len(endpoint) == 0 {
		return nil, RTMError{
			Code:   ERROR_CODE_APPLICATION,
			Reason: ERROR_EMPTY_ENDPOINT,
		}
	}

	if len(appkey) == 0 {
		return nil, RTMError{
			Code:   ERROR_CODE_APPLICATION,
			Reason: ERROR_EMPTY_APP_KEY,
		}
	}

	rtm := &RTM{
		Observer: observer.New(),

		appKey:   appkey,
		endpoint: appendVersion(endpoint),
		opts:     opts,

		subscriptions: subscriptionsType{
			list: make(map[string]*subscription.Subscription),
		},
	}
	rtm.initFSM()

	return rtm, nil
}

// Returns a subsciption. Subscription struct for the associated subscription id.
// The Subscription object must exist. Otherwise function returns ERROR_SUBSCRIPTION_NOT_FOUND error
func (rtm *RTM) GetSubscription(subscriptionId string) (*subscription.Subscription, error) {
	rtm.subscriptions.mutex.Lock()
	defer rtm.subscriptions.mutex.Unlock()
	if sub, ok := rtm.subscriptions.list[subscriptionId]; ok {
		return sub, nil
	}
	return nil, ERROR_SUBSCRIPTION_NOT_FOUND
}

// Publishes a message to a channel.
func (rtm *RTM) Publish(channel string, message interface{}) error {
	_, err := rtm.socketSend("rtm/publish", &pdu.PublishBody{
		Channel: channel,
		Message: rtm.ConvertToRawJson(message),
	}, NOACK)
	return err
}

// Publishes a message to a channel with Acknowledge. The RTM client must be connected.
// Returns the channel that will receive the message when RTM confirms message delivery or error occurred
func (rtm *RTM) PublishAck(channel string, message interface{}) <-chan PublishResponse {
	var err error
	retCh := make(chan PublishResponse, 1)

	c, err := rtm.socketSend("rtm/publish", &pdu.PublishBody{
		Channel: channel,
		Message: rtm.ConvertToRawJson(message),
	}, ACK)
	if err != nil {
		retCh <- PublishResponse{
			Err: err,
		}
		close(retCh)
		return retCh
	}

	go func() {
		defer close(retCh)
		message := <-c

		responseCode := pdu.GetResponseCode(message)
		if responseCode == pdu.CODE_OK_REQUEST {
			var response pdu.PublishBodyResponse
			json.Unmarshal(message.Body, &response)
			retCh <- PublishResponse{
				Response: response,
			}
		} else {
			err := pdu.GetResponseError(message)
			retCh <- PublishResponse{
				Err: RTMError{
					Code:   ERROR_CODE_APPLICATION,
					Reason: err,
				},
			}
		}
	}()

	return retCh
}

// Writes a value to the specified channel. The RTM client must be connected.
// Returns the channel that will receive the message when RTM confirms message delivery or error occurred
func (rtm *RTM) Write(channel string, message interface{}) <-chan WriteResponse {
	var err error
	retCh := make(chan WriteResponse, 1)

	c, err := rtm.socketSend("rtm/write", &pdu.WriteBody{
		Channel: channel,
		Message: rtm.ConvertToRawJson(message),
	}, ACK)

	if err != nil {
		retCh <- WriteResponse{
			Err: err,
		}
		close(retCh)
		return retCh
	}

	go func() {
		defer close(retCh)
		message := <-c

		responseCode := pdu.GetResponseCode(message)
		if responseCode == pdu.CODE_OK_REQUEST {
			var response pdu.WriteBodyResponse
			json.Unmarshal(message.Body, &response)
			retCh <- WriteResponse{
				Response: response,
			}

		} else {
			err := pdu.GetResponseError(message)
			retCh <- WriteResponse{
				Err: RTMError{
					Code:   ERROR_CODE_APPLICATION,
					Reason: err,
				},
			}
		}
	}()

	return retCh
}

// Deletes the value for the associated channel. The RTM client must be connected.
// Returns the channel that will receive the message when RTM confirms message delivery or error occurred
func (rtm *RTM) Delete(channel string) <-chan DeleteResponse {
	var err error
	retCh := make(chan DeleteResponse, 1)

	c, err := rtm.socketSend("rtm/delete", &pdu.DeleteBody{
		Channel: channel,
	}, ACK)

	if err != nil {
		retCh <- DeleteResponse{
			Err: err,
		}
		close(retCh)
		return retCh
	}

	go func() {
		defer close(retCh)
		message := <-c

		responseCode := pdu.GetResponseCode(message)
		if responseCode == pdu.CODE_OK_REQUEST {
			var response pdu.DeleteBodyResponse
			json.Unmarshal(message.Body, &response)
			retCh <- DeleteResponse{
				Response: response,
			}
		} else {
			err := pdu.GetResponseError(message)
			retCh <- DeleteResponse{
				Err: RTMError{
					Code:   ERROR_CODE_APPLICATION,
					Reason: err,
				},
			}
		}
	}()

	return retCh
}

// Reads the latest message written to a specific channel. The RTM client must be connected.
// Returns the channel that will receive the message when RTM responds or error occurred
func (rtm *RTM) Read(channel string) <-chan ReadResponse {
	return rtm.ReadPos(channel, "")
}

// Reads the message with the specified position written to a specific channel. The RTM client must be connected.
// Returns the channel that will receive the message when RTM responds or error occurred
func (rtm *RTM) ReadPos(channel string, position string) <-chan ReadResponse {
	var err error
	retCh := make(chan ReadResponse, 1)

	c, err := rtm.socketSend("rtm/read", &pdu.ReadBody{
		Channel:  channel,
		Position: position,
	}, ACK)

	if err != nil {
		retCh <- ReadResponse{
			Err: err,
		}
		close(retCh)
		return retCh
	}

	go func() {
		defer close(retCh)
		message := <-c

		responseCode := pdu.GetResponseCode(message)
		if responseCode == pdu.CODE_OK_REQUEST {
			var response pdu.ReadBodyResponse
			json.Unmarshal(message.Body, &response)

			retCh <- ReadResponse{
				Response: response,
			}
		} else {
			err := pdu.GetResponseError(message)
			retCh <- ReadResponse{
				Err: RTMError{
					Code:   ERROR_CODE_APPLICATION,
					Reason: err,
				},
			}
		}
	}()

	return retCh
}

// Performs a channel search for a given user-defined prefix. This method passes
// replies to the go-channel.
//
// Go channel contains channel names returned by RTM. Channel will be closed after reading the last message.
// Returns the channel that will receive the messages with channel names when RTM responds or error occurred
func (rtm *RTM) Search(prefix string) <-chan SearchResponse {
	var err error
	retCh := make(chan SearchResponse)

	c, err := rtm.socketSend("rtm/search", &pdu.SearchBody{
		Prefix: prefix,
	}, ACK)

	if err != nil {
		retCh <- SearchResponse{
			Err: err,
		}
		close(retCh)
		return retCh
	}

	go func() {
		defer close(retCh)
		message := <-c

		responseCode := pdu.GetResponseCode(message)
		if responseCode == pdu.CODE_OK_REQUEST {
			channels := make(chan string)
			retCh <- SearchResponse{
				Channels: channels,
			}
			var messages []pdu.RTMQuery
			messages = append(messages, message)
			for message = range c {
				messages = append(messages, message)
			}

			for _, message = range messages {
				var response pdu.SearchBodyResponse
				json.Unmarshal(message.Body, &response)

				for _, channel := range response.Channels {
					channels <- channel
				}
			}
		} else {
			err := pdu.GetResponseError(message)
			retCh <- SearchResponse{
				Err: RTMError{
					Code:   ERROR_CODE_APPLICATION,
					Reason: err,
				},
			}
		}
	}()

	return retCh
}

// Checks if the client is connected
func (rtm *RTM) IsConnected() bool {
	if rtm.fsm.CurrentState() == STATE_CONNECTED {
		return true
	}

	return false
}

// Creates a subscription to the specified channel.
//
// When you create a channel subscription, you can specify additional properties,
// for example, add a filter to the subscription and specify the
// behavior of the SDK when resubscribing after a reconnection.
//
//   subscriptionId string
//
// String that identifies the channel. If you do not
// use the filter parameter, it is the channel name. Otherwise,
// it is a unique identifier for the channel (subscription id).
//
//   mode subscription.Mode
//
// Subscription mode. This mode determines the behaviour of the Golang
// SDK and RTM when resubscribing after a reconnection.
//
// For more information about the options for a subscription,
// see sub-module subscription/Mode (RELIABLE, SIMPLE, ADVANCED) in the online docs.
//
//   opts pdu.SubscribeBodyOpts
//
// Additional subscription options for a channel subscription. These options
// are sent to RTM in the body element of the
// Protocol Data Unit (PDU) that represents the subscribe request.
//
// For more information about the body element of a PDU,
// see sub-module pdu/SubscribeBodyOpts in the online docs.
// and rtm/subscription sub-package
//
//   listener subscription.Listener
//
// Listener instance to define application functionality based on subscription state changes or subscription events.
// For example, you can define callback for when a channel receives a message, when the application
// subscribes or unsubscribes to a channel, or gets the errors.
func (rtm *RTM) Subscribe(subscriptionId string, mode subscription.Mode, opts pdu.SubscribeBodyOpts, listener subscription.Listener) error {
	sub := subscription.New(subscription.Config{
		SubscriptionId: subscriptionId,
		Mode:           mode,
		Opts:           opts,
		Listener:       listener,
	})
	if rtm.fsm.CurrentState() == STATE_CONNECTED {
		err := rtm.processSubscription(sub)
		return err
	} else {
		rtm.subscriptions.mutex.Lock()
		defer rtm.subscriptions.mutex.Unlock()
		rtm.subscriptions.list[subscriptionId] = sub
	}

	return nil
}

func (rtm *RTM) processSubscription(sub *subscription.Subscription) error {
	var subscriptionId = sub.GetSubscriptionId()

	subPdu := sub.SubscribePdu()
	c, err := rtm.socketSend(subPdu.Action, &subPdu.Body, ACK)
	if err != nil {
		return err
	}

	go func() {
		data := <-c

		if pdu.GetResponseCode(data) == pdu.CODE_OK_REQUEST {
			var response pdu.SubscribeOk

			rtm.subscriptions.mutex.Lock()
			rtm.subscriptions.list[subscriptionId] = sub
			rtm.subscriptions.mutex.Unlock()

			json.Unmarshal(data.Body, &response)
			sub.ProcessSubscribe(response)
		} else if pdu.GetResponseCode(data) == pdu.CODE_ERROR_REQUEST {
			var response pdu.SubscribeError
			json.Unmarshal(data.Body, &response)

			sub.ProcessSubscribeError(response)
		}
	}()

	return nil
}

func (rtm *RTM) subscribeAll() error {
	if rtm.fsm.CurrentState() == STATE_CONNECTED {
		rtm.subscriptions.mutex.Lock()
		defer rtm.subscriptions.mutex.Unlock()

		for _, sub := range rtm.subscriptions.list {
			rtm.processSubscription(sub)
		}
		return nil
	}

	return ERROR_NOT_CONNECTED
}

func (rtm *RTM) disconnectAll() {
	for _, sub := range rtm.subscriptions.list {
		sub.ProcessDisconnect()
	}
}

// Removes the specified subscription. The RTM client must be connected.
// Returns the channel that will receive the message when RTM confirms unsubscribing or error occurred
func (rtm *RTM) Unsubscribe(subscriptionId string) <-chan UnsunscribeResponse {
	retCh := make(chan UnsunscribeResponse, 1)

	rtm.subscriptions.mutex.Lock()
	if sub, ok := rtm.subscriptions.list[subscriptionId]; ok {
		rtm.subscriptions.mutex.Unlock()
		query := sub.UnsubscribePdu()
		c, err := rtm.socketSend(query.Action, &query.Body, ACK)
		if err != nil {
			retCh <- UnsunscribeResponse{
				Err: err,
			}
			close(retCh)
			return retCh
		}

		go func() {
			defer close(retCh)
			message := <-c

			responseCode := pdu.GetResponseCode(message)
			if responseCode == pdu.CODE_OK_REQUEST {
				var response pdu.UnsubscribeBodyResponse
				json.Unmarshal(message.Body, &response)

				sub.ProcessUnsubscribe(response)
				rtm.subscriptions.mutex.Lock()
				defer rtm.subscriptions.mutex.Unlock()
				delete(rtm.subscriptions.list, response.SubscriptionId)

				retCh <- UnsunscribeResponse{
					Response: response,
				}

			} else {
				var response pdu.UnsubscribeError
				json.Unmarshal(message.Body, &response)
				sub.ProcessUnsubscribeError(response)

				err := pdu.GetResponseError(message)
				retCh <- UnsunscribeResponse{
					Err: RTMError{
						Code:   ERROR_CODE_APPLICATION,
						Reason: err,
					},
				}
			}
		}()

	} else {
		rtm.subscriptions.mutex.Unlock()
		retCh <- UnsunscribeResponse{
			Err: RTMError{
				Code:   ERROR_CODE_APPLICATION,
				Reason: ERROR_SUBSCRIPTION_NOT_FOUND,
			},
		}
		close(retCh)
	}

	return retCh
}

func (rtm *RTM) handleMessage(message pdu.RTMQuery) error {
	act := message.Action
	switch {
	case act == "rtm/subscription/data":
		var response pdu.SubscriptionData
		err := json.Unmarshal(message.Body, &response)
		if err != nil {
			return err
		}
		sub, err := rtm.GetSubscription(response.SubscriptionId)
		if err != nil {
			return err
		}
		sub.ProcessData(response)

	case act == "rtm/subscription/info":
		var response pdu.SubscriptionInfo
		err := json.Unmarshal(message.Body, &response)
		if err != nil {
			return err
		}
		sub, err := rtm.GetSubscription(response.SubscriptionId)
		if err != nil {
			return err
		}
		sub.ProcessInfo(response)
	case act == "rtm/subscription/error":
		var response pdu.SubscriptionError
		err := json.Unmarshal(message.Body, &response)
		if err != nil {
			return err
		}
		sub, err := rtm.GetSubscription(response.SubscriptionId)
		if err != nil {
			return err
		}
		sub.ProcessSubscriptionError(response)
	}

	rtm.Fire(message.Action, message)
	return nil
}

// Starts the client.
//
// The client begins to establish the WebSocket connection
// to RTM and then tracks the state of the connection. If the WebSocket
// connection drops for any reason, the Go SDK attempts to reconnect.
//
// You can use Event-Based model to catch application events,
// for example, when the application enters or leaves the
// connecting or connected states.
func (rtm *RTM) Start() {
	rtm.Fire(EVENT_START, nil)
}

func (rtm *RTM) connect() error {
	var err error

	logger.Info("Connecting to", rtm.endpoint)
	rtm.conn, err = connection.New(rtm.endpoint + "?appkey=" + rtm.appKey)

	if err != nil {
		return RTMError{
			Code:   ERROR_CODE_TRANSPORT,
			Reason: err,
		}
	}

	// Subscribe to all messages
	go func(rtm *RTM) {
		for {
			message, err := rtm.socketRead()
			if err != nil {
				// Broken connection.
				return
			}
			err = rtm.handleMessage(message)
			if err != nil {
				logger.Error(err)
				rtm.Fire(EVENT_ERROR, RTMError{
					Code:   ERROR_CODE_PDU,
					Reason: err,
				})
			}
		}
	}(rtm)

	// Auth
	if rtm.opts.AuthProvider != nil {
		err = rtm.opts.AuthProvider.Authenticate(rtm.conn)
		if err != nil {
			// Authentication error
			return RTMError{
				Code:   ERROR_CODE_AUTHENTICATION,
				Reason: err,
			}
		}
		rtm.Fire(EVENT_AUTHENTICATED, nil)
	}

	rtm.Fire(EVENT_OPEN, nil)

	return nil
}

// Stops the client. The SDK begins to close the WebSocket connection and
// does not start it again unless you call Start().
//
// Use this method to explicitly stop all interaction with RTM.
//
// You can use Event-Based model to define application functionality,
// for example, when the application enters or leaves the stopped state.
func (rtm *RTM) Stop() {
	rtm.Fire(EVENT_STOP, nil)
}

func (rtm *RTM) closeConnection() {
	rtm.conn.Close()
}

func (rtm *RTM) socketSend(action string, body interface{}, ack bool) (<-chan pdu.RTMQuery, error) {
	if !rtm.IsConnected() {
		return nil, RTMError{
			Code:   ERROR_CODE_APPLICATION,
			Reason: ERROR_NOT_CONNECTED,
		}
	}

	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, RTMError{
			Code:   ERROR_CODE_INVALID_JSON,
			Reason: err,
		}
	}

	var ch <-chan pdu.RTMQuery

	if ack {
		ch, err = rtm.conn.SendAck(action, rawBody)
	} else {
		err = rtm.conn.Send(action, rawBody)
	}

	if err != nil {
		rtm.Fire(EVENT_ERROR, RTMError{
			Code:   ERROR_CODE_TRANSPORT,
			Reason: err,
		})
		rtm.Fire(EVENT_CLOSE, nil)
		return nil, err
	}

	return ch, nil
}

func (rtm *RTM) socketRead() (pdu.RTMQuery, error) {
	response, err := rtm.conn.Read()
	if err != nil {
		rtm.Fire(EVENT_ERROR, RTMError{
			Code:   ERROR_CODE_TRANSPORT,
			Reason: err,
		})
		rtm.Fire(EVENT_CLOSE, nil)
		return pdu.RTMQuery{}, err
	}

	return response, nil
}

func appendVersion(endpoint string) string {
	re := regexp.MustCompile("/(v\\d+)$")
	ver := re.FindString(endpoint)

	if len(ver) > 0 {
		logger.Warn("Specifying RTM endpoint with protocol version is deprecated.")
		logger.Warn("Please remove version '" + ver + "' from endpoint: '" + endpoint + "'")

		return endpoint
	}

	if endpoint[len(endpoint)-1:] != "/" {
		endpoint += "/"
	}
	return endpoint + API_VER
}
