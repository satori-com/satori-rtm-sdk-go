// RTM Connection.
//
// Access the RTM Service on the connection level to connect to the RTM Service, send and receive PDUs, and wait
// for responses from the RTM Service.
package connection

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

const (
	MAX_ID                     = math.MaxInt32
	MAX_ACKS_QUEUE_LENGTH      = 10000
	MAX_UNPROCESSED_ACKS_QUEUE = 100
)

type Connection struct {
	wsConn *websocket.Conn
	lastID int
	acks   acksType
	mutex  sync.Mutex

	// http://godoc.org/github.com/gorilla/websocket#hdr-Concurrency
	// Gorilla websocket package is not thread-safe. So we need to handle it by ourselves
	rSockMutex sync.Mutex
	wSockMutex sync.Mutex
}

type acksType struct {
	ch        chan pdu.RTMQuery
	listeners map[string]chan pdu.RTMQuery
	mutex     sync.Mutex
}

type Options struct {
	Proxy *url.URL
}

// Creates a new instance for a specific RTM Service endpoint.
// Establishes Websocket connection to the Service.
func New(endpoint string, opts Options) (*Connection, error) {
	var err error
	dialer := websocket.Dialer{
		Proxy: http.ProxyURL(opts.Proxy),
	}

	conn := &Connection{}
	conn.lastID = 0
	conn.wsConn, _, err = dialer.Dial(endpoint, nil)
	if err != nil {
		return nil, err
	}

	conn.initAcks()

	return conn, nil
}

// Closes a specific connection. The close event is propagated to all listeners.
func (c *Connection) Close() {
	defer func() {
		// Channel can be already closed. Call recover to avoid panic when closing closed channel
		recover()
	}()
	if c.wsConn != nil {
		c.wsConn.Close()
	}

	// Close Ack listeners channel
	for _, ch := range c.acks.listeners {
		close(ch)
	}

	close(c.acks.ch)
}

// Sends a Protocol Data Unit (PDU) to the RTM Service. The typed response from
// the RTM Service is passed to the go-channel.
//
// This method combines the specified operation with the PDU body into a PDU and
// sends it to the RTM Service. The PDU body must be able to be serialized into a JSON object.
//
// This method should be used when RTM sends multiple PDUs response. All incoming PDUs from the
// RTM Service will be passed to go-channel.
func (c *Connection) SendAck(action string, body json.RawMessage) (<-chan pdu.RTMQuery, error) {
	query := pdu.RTMQuery{
		Action: action,
		Body:   body,
		Id:     c.nextID(),
	}

	ch := make(chan pdu.RTMQuery, 1)
	c.addListener(query.Id, ch)

	return ch, c.socketSend(query)
}

// Sends a Protocol Data Unit (PDU) to the RTM Service.
//
// This method combines the specified operation with the PDU body into a PDU and
// sends it to the RTM Service. The PDU body must be able to be serialized into a JSON object.
func (c *Connection) Send(action string, body json.RawMessage) error {
	query := pdu.RTMQuery{
		Action: action,
		Body:   body,
	}

	return c.socketSend(query)
}

func (c *Connection) socketSend(query pdu.RTMQuery) error {
	message, err := json.Marshal(&query)
	if err != nil {
		return err
	}

	logger.Debug("send>", string(message))
	c.wSockMutex.Lock()
	defer c.wSockMutex.Unlock()
	err = c.wsConn.WriteMessage(websocket.TextMessage, message)

	if err != nil {
		c.Close()
		return err
	}

	return nil
}

// Reads a new message from the Websocket connection and convert the message to pdu.RTMQuery
func (c *Connection) Read() (pdu.RTMQuery, error) {
	var response pdu.RTMQuery

	c.rSockMutex.Lock()
	defer c.rSockMutex.Unlock()
	_, data, err := c.wsConn.ReadMessage()
	if err != nil {
		c.Close()
		return pdu.RTMQuery{}, err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		c.Close()
		return pdu.RTMQuery{}, err
	}

	logger.Debug("recv<", response.String())

	if len(response.Id) != 0 {
		c.acks.ch <- response
	}

	return response, nil
}

func (c *Connection) nextID() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lastID == MAX_ID {
		c.lastID = 0
	}
	c.lastID++
	return strconv.Itoa(c.lastID)
}

func (c *Connection) initAcks() {
	c.acks.ch = make(chan pdu.RTMQuery, MAX_UNPROCESSED_ACKS_QUEUE)
	c.acks.listeners = make(map[string]chan pdu.RTMQuery, MAX_ACKS_QUEUE_LENGTH)

	go func(c *Connection) {
		for response := range c.acks.ch {
			c.acks.mutex.Lock()
			ch := c.acks.listeners[response.Id]
			c.acks.mutex.Unlock()

			// Exception for the "search" API: Do not delete listener channel until the last message
			if response.Action != "rtm/search/data" {
				c.deleteListener(response.Id)
			}

			if ch != nil {
				if pdu.GetResponseCode(response) != pdu.CODE_DATA_REQUEST {
					defer close(ch)
				}
				ch <- response
			}
		}
	}(c)
}

func (c *Connection) addListener(id string, channel chan pdu.RTMQuery) {
	c.acks.mutex.Lock()
	defer c.acks.mutex.Unlock()
	c.acks.listeners[id] = channel
}

func (c *Connection) deleteListener(id string) {
	c.acks.mutex.Lock()
	defer c.acks.mutex.Unlock()
	delete(c.acks.listeners, id)
}
