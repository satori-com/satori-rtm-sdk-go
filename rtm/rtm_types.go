package rtm

import (
	"github.com/satori-com/satori-rtm-sdk-go/rtm/connection"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"sync"
)

type Auth interface {
	Authenticate(conn *connection.Connection) error
}

type Options struct {
	AuthProvider Auth
	HttpsProxy   Proxy
}

type Proxy struct {
	Host string
	Port int
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
