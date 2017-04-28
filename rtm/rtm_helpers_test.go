package rtm

import (
	"encoding/json"
	"errors"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

type credentialsT struct {
	Endpoint          string `json:"endpoint"`
	AppKey            string `json:"appkey"`
	RoleName          string `json:"auth_role_name"`
	RoleSecretKey     string `json:"auth_role_secret_key"`
	RestrictedChannel string `json:"auth_restricted_channel"`
}

/*
 *	Helper functions
 */
func getRTM() (*RTM, error) {
	credentials, err := getCredentials()
	if err != nil {
		return &RTM{}, err
	}

	authProvider := auth.New(credentials.RoleName, credentials.RoleSecretKey)
	client, _ := New(credentials.Endpoint, credentials.AppKey, Options{
		AuthProvider: authProvider,
	})
	return client, nil
}

func waitForConnected(rtm *RTM) error {
	connected := make(chan bool)
	rtm.On(EVENT_CONNECTED, func(interface{}) {
		connected <- true
	})
	select {
	case <-connected:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("Timeout")
	}
}

func getCredentials() (credentialsT, error) {
	var credentials credentialsT
	var content []byte
	var err error
	// Check credentials ENV path
	path := os.Getenv("CREDENTIALS")

	content, err = ioutil.ReadFile(path)
	if err != nil {
		content, err = ioutil.ReadFile("../credentials.json")
	}

	if err == nil {
		err = json.Unmarshal(content, &credentials)

		if err != nil {
			return credentials, err
		}
		return credentials, nil
	}

	return credentials, err
}

func getChannel() string {
	return randStringRunes("channel-", 6)
}

func randStringRunes(prefix string, n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return prefix + string(b)
}
