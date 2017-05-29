package connection

import (
	"os"
	"io/ioutil"
	"encoding/json"
)

type credentialsT struct {
	Endpoint          string `json:"endpoint"`
	AppKey            string `json:"appkey"`
	RoleName          string `json:"auth_role_name"`
	RoleSecretKey     string `json:"auth_role_secret_key"`
	RestrictedChannel string `json:"auth_restricted_channel"`
}

func getCredentials() (credentialsT, error) {
	var credentials credentialsT
	var content []byte
	var err error
	// Check credentials ENV path
	path := os.Getenv("CREDENTIALS")

	content, err = ioutil.ReadFile(path)
	if err != nil {
		content, err = ioutil.ReadFile("./credentials.json")
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
