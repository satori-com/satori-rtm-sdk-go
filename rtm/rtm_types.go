package rtm

import (
	"github.com/satori-com/satori-rtm-sdk-go/rtm/connection"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/subscription"
	"net/http"
	"net/url"
	"sync"
)

type Auth interface {
	Authenticate(conn *connection.Connection) error
}

type Options struct {
	AuthProvider Auth

	// Proxy specifies a function to return a proxy for a given
	// Request. If the function returns a non-nil error, the
	// request is aborted with the provided error.
	// If Proxy is nil or returns a nil *URL, no proxy is used.
	//
	// Check ProxyFromEnvironment, as an example: https://golang.org/src/net/http/transport.go?s=9778:9835#L250
	Proxy func(*http.Request) (*url.URL, error)
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
