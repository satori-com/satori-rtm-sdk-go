package rtm

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestRTM_Write(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	c := <-client.Write(getChannel(), 4)
	if c.Err != nil {
		t.Fatal("Unable to write int. Response: ", c.Err)
	}
}

func TestRTM_Write_Broken(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	raw := json.RawMessage("{{{123]]]")
	c := <-client.Write(getChannel(), raw)

	if c.Err == nil {
		t.Fatal("Did not get the error, but should")
	}
}

func TestRTM_Read_Int(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	var i int = 4
	channel := getChannel()
	<-client.Write(channel, i)

	c := <-client.Read(channel)

	if c.Err != nil {
		t.Fatal(c.Err)
	}

	val, _ := strconv.Atoi(string(c.Response.Message))
	if val != 4 {
		t.Error("Read not int value")
	}
}

func TestRTM_Read_Struct(t *testing.T) {
	type A struct {
		Name string
		Age  int
		Car  bool
	}
	var a, b A
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	a = A{
		Name: "User",
		Age:  19,
		Car:  true,
	}
	channel := getChannel()
	<-client.Write(channel, a)

	c := <-client.Read(channel)

	if c.Err != nil {
		t.Fatal(c.Err)
	}

	err = json.Unmarshal(c.Response.Message, &b)
	if err != nil || a != b {
		t.Error("Structs do not match ", err)
	}
}

func TestRTM_Read_Empty(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	c := <-client.Read(getChannel())

	if c.Err != nil || string(c.Response.Message) != "null" {
		t.Fatal("Not 'null' message returned from non-existing channel")
	}
}

func TestRTM_Read_WrongChannel(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	c := <-client.Read("")

	if c.Err == nil {
		t.Fatal("We have successfully read from channel without name, but should not")
	}
}

func TestRTM_Read_Double(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	channel := getChannel()

	client.Write(channel, "привет1")
	<-client.Write(channel, "привет2")

	c := <-client.Read(channel)

	if string(c.Response.Message) != "\"привет2\"" {
		t.Fatal("Wrong reading order")
	}
}

func TestRTM_Delete_Existing(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	channel := getChannel()

	<-client.Write(channel, 1)
	c := <-client.Read(channel)
	if string(c.Response.Message) != "1" {
		t.Fatal("Wrong reading value")
	}

	d := <-client.Delete(channel)

	if d.Err != nil {
		t.Fatal("Error occured when deleting from channel")
	}

	c = <-client.Read(channel)
	if string(c.Response.Message) != "null" {
		t.Fatal("Wrong reading value after delete")
	}
}

func TestRWPDenied(t *testing.T) {
	credentials, err := getCredentials()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}

	client, _ := New(credentials.Endpoint, credentials.AppKey, Options{})
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	// Check publish
	p := <-client.PublishAck("$system.channel", 123)

	if p.Err == nil {
		t.Fatal("Publish: Response was successfull, but should not")
	}

	rtmErr := p.Err.(RTMError)
	if rtmErr.Code != 0 || rtmErr.Reason.Error() != "{\"error\":\"authorization_denied\",\"reason\":\"Unauthorized\"}" {
		t.Fatal("Publish: Unexpected error")
	}

	// Check write
	w := <-client.Write("$system.channel", 123)

	if w.Err == nil {
		t.Fatal("Write: Response was successfull, but should not")
	}

	rtmErr = w.Err.(RTMError)
	if rtmErr.Code != 0 || rtmErr.Reason.Error() != "{\"error\":\"authorization_denied\",\"reason\":\"Unauthorized\"}" {
		t.Fatal("Write: Unexpected error")
	}

	// Check read
	r := <-client.Read("$system.channel")

	if r.Err == nil {
		t.Fatal("Read: Response was successfull, but should not")
	}

	rtmErr = r.Err.(RTMError)
	if rtmErr.Code != 0 || rtmErr.Reason.Error() != "{\"error\":\"authorization_denied\",\"reason\":\"Unauthorized\"}" {
		t.Fatal("Read: Unexpected error")
	}

	// Check delete
	d := <-client.Delete("$system.channel")

	if d.Err == nil {
		t.Fatal("Delete: Response was successfull, but should not")
	}

	rtmErr = d.Err.(RTMError)
	if rtmErr.Code != 0 || rtmErr.Reason.Error() != "{\"error\":\"authorization_denied\",\"reason\":\"Unauthorized\"}" {
		t.Fatal("Delete: Unexpected error")
	}
}

func TestRTM_Publish(t *testing.T) {
	client, err := getRTM()
	if err != nil {
		t.Skip("Unable to find credentials. Skip test")
	}
	defer client.Stop()
	go client.Start()

	if err = waitForConnected(client); err != nil {
		t.Fatal(err)
	}

	c := <-client.PublishAck(getChannel(), "123")

	if c.Err != nil {
		t.Fatal("Unable to publish data. Got error:", c.Err)
	}
}
