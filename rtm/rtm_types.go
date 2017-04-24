package rtm

import (
	"github.com/satori-com/satori-rtm-sdk-go/rtm/connection"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"sync"
)

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
	EVENT_CLOSED           = "closed"
	EVENT_OPEN             = "open"
	EVENT_ERROR            = "error"
	EVENT_DATA_ERROR       = "dataError"
	EVENT_AUTHENTICATED    = "authenticated"
)

type Auth interface {
	Authenticate(conn *connection.Connection) error
}

type Options struct {
	AuthProvider Auth
}

type subscriptionsType struct {
	list  map[string]*subscription.Subscription
	mutex sync.Mutex
}

type PublishResponse struct {
	Response pdu.PublishBodyResponse
	Err      error
}

type WriteResponse struct {
	Response pdu.WriteBodyResponse
	Err      error
}

type ReadResponse struct {
	Response pdu.ReadBodyResponse
	Err      error
}

type DeleteResponse struct {
	Response pdu.DeleteBodyResponse
	Err      error
}

type SearchResponse struct {
	Channels <-chan string
	Err      error
}

type UnsunscribeResponse struct {
	Response pdu.UnsubscribeBodyResponse
	Err      error
}
