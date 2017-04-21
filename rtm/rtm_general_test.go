package rtm

import "testing"

func TestEmptyAppKey(t *testing.T) {
	_, err := New("ws://wrong-host-name.www", "", Options{})

	rtmErr := err.(RTMError)
	if rtmErr.Code != ERROR_CODE_APPLICATION || rtmErr.Reason != ERROR_EMPTY_APP_KEY {
		t.Fatal("AppKey checking does not work")
	}
}

func TestEmptyEndpoint(t *testing.T) {
	_, err := New("", "", Options{})

	rtmErr := err.(RTMError)
	if rtmErr.Code != ERROR_CODE_APPLICATION || rtmErr.Reason != ERROR_EMPTY_ENDPOINT {
		t.Fatal("endpoint checking does not work")
	}
}

func TestNonexistingSubscription(t *testing.T) {
	client, _ := getRTM()
	_, err := client.GetSubscription("non-existing")

	if err != ERROR_SUBSCRIPTION_NOT_FOUND {
		t.Fatal("Got subscription, but should not")
	}
}

func TestVersionedEndpoint(t *testing.T) {
	client, _ := New("wss://some-host-name.www/v123", "123", Options{})

	if client.endpoint != "wss://some-host-name.www/v123" {
		t.Fatal("Client modified versioned endpoint")
	}
}
