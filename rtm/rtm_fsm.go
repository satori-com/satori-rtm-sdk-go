package rtm

import (
	"github.com/satori-com/satori-rtm-sdk-go/fsm"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"math/rand"
	"time"
)

func (rtm *RTM) initFSM() {
	rtm.fsm, _ = fsm.New("stopped", fsm.States{
		STATE_STOPPED: fsm.Events{
			"enterStopped": func(f *fsm.FSM) {
				logger.Info("Client: Enter Stopped")
				rtm.Fire(EVENT_STOPPED, nil)
				rtm.closeConnection()
			},
			"leaveStopped": func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_STOPPED, nil)
			},
			"start": func(f *fsm.FSM) {
				f.Transition(STATE_CONNECTING)
			},
		},
		STATE_CONNECTING: fsm.Events{
			"enterConnecting": func(f *fsm.FSM) {
				logger.Info("Client: Enter Connecting")
				rtm.Fire(EVENT_CONNECTING, nil)
				err := rtm.connect()
				if err.Reason != nil {
					logger.Error(err.Reason)
					rtm.Fire(EVENT_ERROR, err)
				}
			},
			"leaveConnecting": func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_CONNECTING, nil)
			},
			"open": func(f *fsm.FSM) {
				f.Transition(STATE_CONNECTED)
			},
			"error": func(f *fsm.FSM) {
				f.Transition(STATE_AWAITING)
			},
			"close": func(f *fsm.FSM) {
				f.Transition(STATE_AWAITING)
			},
			"stop": func(f *fsm.FSM) {
				f.Transition(STATE_STOPPED)
			},
		},
		STATE_CONNECTED: fsm.Events{
			"enterConnected": func(f *fsm.FSM) {
				logger.Info("Client: Enter Connected")
				rtm.Fire(EVENT_CONNECTED, nil)
				rtm.reconnectCount = 0
				rtm.subscribeAll()

			},
			"leaveConnected": func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_CONNECTED, nil)
				rtm.disconnectAll()
			},
			"close": func(f *fsm.FSM) {
				f.Transition(STATE_AWAITING)
			},
			"stop": func(f *fsm.FSM) {
				f.Transition(STATE_STOPPED)
			},
			"error": func(f *fsm.FSM) {
				f.Transition(STATE_AWAITING)
			},
		},
		STATE_AWAITING: fsm.Events{
			"enterAwaiting": func(f *fsm.FSM) {
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
			"leaveAwaiting": func(f *fsm.FSM) {
				rtm.Fire(EVENT_LEAVE_AWAITING, nil)
			},
			"stop": func(f *fsm.FSM) {
				f.Transition(STATE_STOPPED)
			},
		},
	})

	events := []string{EVENT_OPEN, EVENT_ERROR, EVENT_START, EVENT_STOP}
	for _, event := range events {
		func(event string) {
			rtm.On(event, func(data interface{}) {
				rtm.fsm.Event(fsm.EventName(event))
			})
		}(event)
	}
}

func (rtm *RTM) nextReconnectInterval() time.Duration {
	reconnect_sec := rtm.reconnectCount * rtm.reconnectCount
	if reconnect_sec > MAX_RECONNECT_TIME_SEC {
		reconnect_sec = MAX_RECONNECT_TIME_SEC
	}
	jitter := rand.Intn(100)
	return time.Duration(reconnect_sec)*time.Second + time.Duration(jitter)*time.Millisecond
}
