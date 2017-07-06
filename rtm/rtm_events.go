package rtm

const (
	EVENT_STOPPED          = "enterStopped"
	EVENT_LEAVE_STOPPED    = "leaveStopped"
	EVENT_CONNECTING       = "enterConnecting"
	EVENT_LEAVE_CONNECTING = "leaveConnecting"
	EVENT_CONNECTED        = "enterConnected"
	EVENT_LEAVE_CONNECTED  = "leaveConnected"
	EVENT_AWAITING         = "enterAwaiting"
	EVENT_LEAVE_AWAITING   = "leaveAwaiting"
	EVENT_START            = "start"
	EVENT_STOP             = "stop"
	EVENT_OPEN             = "open"
	EVENT_CLOSE            = "close"
	EVENT_ERROR            = "error"
	EVENT_AUTHENTICATED    = "authenticated"
)

// EVENT_STOPPED

func (rtm *RTMClient) OnStopped(callback func()) interface{} {
	return rtm.On(EVENT_STOPPED, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnStoppedOnce(callback func()) {
	rtm.Once(EVENT_STOPPED, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveStopped(callback func()) interface{} {
	return rtm.On(EVENT_LEAVE_STOPPED, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveStoppedOnce(callback func()) {
	rtm.Once(EVENT_LEAVE_STOPPED, func(data interface{}) {
		callback()
	})
}

// EVENT_CONNECTING

func (rtm *RTMClient) OnConnecting(callback func()) interface{} {
	return rtm.On(EVENT_CONNECTING, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnConnectingOnce(callback func()) {
	rtm.Once(EVENT_CONNECTING, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveConnecting(callback func()) interface{} {
	return rtm.On(EVENT_LEAVE_CONNECTING, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveConnectingOnce(callback func()) {
	rtm.Once(EVENT_LEAVE_CONNECTING, func(data interface{}) {
		callback()
	})
}

// EVENT_CONNECTED

func (rtm *RTMClient) OnConnected(callback func()) interface{} {
	return rtm.On(EVENT_CONNECTED, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnConnectedOnce(callback func()) {
	rtm.Once(EVENT_CONNECTED, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveConnected(callback func()) interface{} {
	return rtm.On(EVENT_LEAVE_CONNECTED, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveConnectedOnce(callback func()) {
	rtm.Once(EVENT_LEAVE_CONNECTED, func(data interface{}) {
		callback()
	})
}

// EVENT_AWAITING

func (rtm *RTMClient) OnAwaiting(callback func()) interface{} {
	return rtm.On(EVENT_AWAITING, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnAwaitingOnce(callback func()) {
	rtm.Once(EVENT_AWAITING, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveAwaiting(callback func()) interface{} {
	return rtm.On(EVENT_LEAVE_AWAITING, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnLeaveAwaitingOnce(callback func()) {
	rtm.Once(EVENT_LEAVE_AWAITING, func(data interface{}) {
		callback()
	})
}

// Other events

func (rtm *RTMClient) OnStart(callback func()) interface{} {
	return rtm.On(EVENT_START, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnStartOnce(callback func()) {
	rtm.Once(EVENT_START, func(data interface{}) {
		callback()
	})
}

func (rtm *RTMClient) OnStop(callback func()) interface{} {
	return rtm.On(EVENT_STOP, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnStopOnce(callback func()) {
	rtm.Once(EVENT_STOP, func(data interface{}) {
		callback()
	})
}

func (rtm *RTMClient) OnOpen(callback func()) interface{} {
	return rtm.On(EVENT_OPEN, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnOpenOnce(callback func()) {
	rtm.Once(EVENT_OPEN, func(data interface{}) {
		callback()
	})
}

func (rtm *RTMClient) OnError(callback func(err RTMError)) interface{} {
	return rtm.On(EVENT_ERROR, func(data interface{}) {
		err := data.(RTMError)
		callback(err)
	})
}
func (rtm *RTMClient) OnErrorOnce(callback func(err RTMError)) {
	rtm.Once(EVENT_ERROR, func(data interface{}) {
		err := data.(RTMError)
		callback(err)
	})
}

func (rtm *RTMClient) OnAuthenticated(callback func()) interface{} {
	return rtm.On(EVENT_AUTHENTICATED, func(data interface{}) {
		callback()
	})
}
func (rtm *RTMClient) OnAuthenticatedOnce(callback func()) {
	rtm.Once(EVENT_AUTHENTICATED, func(data interface{}) {
		callback()
	})
}
