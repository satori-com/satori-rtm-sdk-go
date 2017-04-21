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
				rtm.Fire("enterStopped", nil)
				rtm.closeConnection()
			},
			"leaveStopped": func(f *fsm.FSM) {
				rtm.Fire("leaveStopped", nil)
			},
			"start": func(f *fsm.FSM) {
				f.Transition(STATE_CONNECTING)
			},
		},
		STATE_CONNECTING: fsm.Events{
			"enterConnecting": func(f *fsm.FSM) {
				logger.Info("Client: Enter Connecting")
				rtm.Fire("enterConnecting", nil)
				err := rtm.connect()
				if err != nil {
					logger.Error(err)
					rtm.Fire("error", err)
				}
			},
			"leaveConnecting": func(f *fsm.FSM) {
				rtm.Fire("leaveConnecting", nil)
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
				rtm.Fire("enterConnected", nil)
				rtm.reconnectCount = 0
				rtm.subscribeAll()

			},
			"leaveConnected": func(f *fsm.FSM) {
				rtm.Fire("leaveConnected", nil)
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
				rtm.Fire("enterAwaiting", nil)
				rtm.closeConnection()

				go func() {
					reconnectTime := rtm.nextReconnectInterval()
					logger.Info("Client: Reconnect after ", reconnectTime, "sec")
					<-time.After(reconnectTime)
					rtm.reconnectCount++
					if f.CurrentState() == STATE_AWAITING {
						f.Transition(STATE_CONNECTING)
					}
				}()
			},
			"leaveAwaiting": func(f *fsm.FSM) {
				rtm.Fire("leaveAwaiting", nil)
			},
			"stop": func(f *fsm.FSM) {
				f.Transition(STATE_STOPPED)
			},
		},
	})

	events := []string{"open", "close", "error", "start", "stop", "reconnect"}
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
