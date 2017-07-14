package connection

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEmptyEndpoint(t *testing.T) {
	_, err := New("", Options{})
	if !strings.Contains(err.Error(), "malformed") {
		t.Fatal("Unprocessed empty endpoint error")
	}
}

func TestBadSSLSelfSigned(t *testing.T) {
	_, err := New("wss://self-signed.badssl.com/", Options{})
	if !strings.Contains(err.Error(), "certificate signed by unknown authority") {
		t.Fatal("Connected to host with self-signed certificate")
	}
}

func TestBadSSLExpired(t *testing.T) {
	_, err := New("wss://expired.badssl.com/", Options{})
	if !strings.Contains(err.Error(), "certificate has expired") {
		t.Fatal("Connected to host with expired certificate")
	}
}

func TestBasicConnection(t *testing.T) {
	conn, err := New("ws://echo.websocket.org/", Options{})
	if err != nil {
		t.Fatal("Unable to connect to ws://echo.websocket.org/")
	}

	conn.Close()
}

func TestListener(t *testing.T) {
	cred, err := getCredentials()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	conn, err := New(cred.Endpoint+"v2?appkey="+cred.AppKey, Options{})
	if err != nil {
		t.Fatal("Unable to connect to " + cred.Endpoint)
	}

	resp, err := conn.SendAck("test", json.RawMessage("{}"))
	go conn.Read()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-resp:
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to get response")
	}
}

func TestSocketSend(t *testing.T) {
	cred, err := getCredentials()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	conn, err := New(cred.Endpoint+"v2?appkey="+cred.AppKey, Options{})
	if err != nil {
		t.Fatal("Unable to connect to " + cred.Endpoint)
	}

	err = conn.Send("test", json.RawMessage("{}"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestBrokenConnection(t *testing.T) {
	cred, err := getCredentials()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	conn, err := New(cred.Endpoint+"v2?appkey="+cred.AppKey, Options{})
	if err != nil {
		t.Fatal("Unable to connect to " + cred.Endpoint)
	}

	// Brake connection
	conn.Close()

	err = conn.Send("test", json.RawMessage("{}"))
	if err == nil {
		t.Fatal("We were able to send message to broken connection")
	}
}

func TestCloseWithListeners(t *testing.T) {
	cred, err := getCredentials()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	conn, err := New(cred.Endpoint+"v2?appkey="+cred.AppKey, Options{})
	if err != nil {
		t.Fatal("Unable to connect to " + cred.Endpoint)
	}

	resp, err := conn.SendAck("test", json.RawMessage("{}"))
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
	select {
	case _, ok := <-resp:
		if ok {
			t.Fatal("ok")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Unable to wait for response")
	}

}

func TestMaxNextID(t *testing.T) {
	cred, err := getCredentials()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	conn, err := New(cred.Endpoint+"v2?appkey="+cred.AppKey, Options{})
	if err != nil {
		t.Fatal("Unable to connect to " + cred.Endpoint)
	}

	conn.lastID = MAX_ID
	if conn.nextID() != "1" {
		t.Fatal("Unable to reset lastID. Int overflow")
	}
}
