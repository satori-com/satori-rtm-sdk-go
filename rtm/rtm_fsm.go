package rtm

import (
	"github.com/satori-com/satori-rtm-sdk-go/fsm"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"math/rand"
	"time"
)

func (rtm *RTMClient) initFSM() {
	rtm.fsm, _ = fsm.New(STATE_STOPPED, fsm.States{
		STATE_STOPPED: fsm.Events{
			EVENT_STOPPED: func(f *fsm.FSM) {
				logger.Info("Client: Enter Stopped")
				rtm.closeConnection()
				rtm.Fire(EVENT_STOPPED, nil)
			},
			EVENT_LEAVE_STOPPED: func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_STOPPED, nil)
			},
			EVENT_START: func(f *fsm.FSM) {
				f.Transition(STATE_CONNECTING)
			},
		},
		STATE_CONNECTING: fsm.Events{
			EVENT_CONNECTING: func(f *fsm.FSM) {
				logger.Info("Client: Enter Connecting")
				rtm.Fire(EVENT_CONNECTING, nil)
				err := rtm.connect()
				if err != nil {
					rtmErr := err.(RTMError)
					logger.Error(rtmErr.Reason)

					rtm.Fire(EVENT_ERROR, rtmErr)
					rtm.Fire(EVENT_CLOSE, nil)
				}
			},
			EVENT_LEAVE_CONNECTING: func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_CONNECTING, nil)
			},
			EVENT_OPEN: func(f *fsm.FSM) {
				f.Transition(STATE_CONNECTED)
			},
			EVENT_CLOSE: func(f *fsm.FSM) {
				f.Transition(STATE_AWAITING)
			},
			EVENT_STOP: func(f *fsm.FSM) {
				f.Transition(STATE_STOPPED)
			},
		},
		STATE_CONNECTED: fsm.Events{
			EVENT_CONNECTED: func(f *fsm.FSM) {
				logger.Info("Client: Enter Connected")
				rtm.Fire(EVENT_CONNECTED, nil)
				rtm.reconnectCount = 0
				rtm.subscribeAll()

			},
			EVENT_LEAVE_CONNECTED: func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_CONNECTED, nil)
				rtm.disconnectAll()
			},
			EVENT_CLOSE: func(f *fsm.FSM) {
				f.Transition(STATE_AWAITING)
			},
			EVENT_STOP: func(f *fsm.FSM) {
				f.Transition(STATE_STOPPED)
			},
		},
		STATE_AWAITING: fsm.Events{
			EVENT_AWAITING: func(f *fsm.FSM) {
				logger.Info("Client: Enter Awaiting")
				rtm.Fire(EVENT_AWAITING, nil)
				rtm.closeConnection()

				go func() {
					reconnectTime := rtm.nextReconnectInterval()
					logger.Info("Client: Reconnect after", reconnectTime)
					<-time.After(reconnectTime)
					rtm.reconnectCount++
					if f.CurrentState() == STATE_AWAITING {
						f.Transition(STATE_CONNECTING)
					}
				}()
			},
			EVENT_LEAVE_AWAITING: func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_AWAITING, nil)
			},
			EVENT_STOP: func(f *fsm.FSM) {
				f.Transition(STATE_STOPPED)
			},
		},
	})

	events := []string{EVENT_OPEN, EVENT_START, EVENT_STOP, EVENT_CLOSE}
	for _, event := range events {
		func(event string) {
			rtm.On(event, func(data interface{}) {
				rtm.fsm.Event(fsm.EventName(event))
			})
		}(event)
	}
}

func (rtm *RTMClient) nextReconnectInterval() time.Duration {
	reconnect_sec := rtm.reconnectCount * rtm.reconnectCount
	if reconnect_sec > MAX_RECONNECT_TIME_SEC {
		reconnect_sec = MAX_RECONNECT_TIME_SEC
	}
	jitter := rand.Intn(100)
	return time.Duration(reconnect_sec)*time.Second + time.Duration(jitter)*time.Millisecond
}
