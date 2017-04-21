// RoleKey auth provider.
// Covers auth protocol (handshake). Sends handshake request and processes server response.
//
// The role-based authentication method is a two-step authentication process
// based on the HMAC process, using the MD5 hashing routine:
//  - The client obtains a nonce from the server in a handshake request.
//  - The client then sends an authorization request with its role secret key hashed with the received nonce.
//
// Obtain a role secret key from the https://developer.satori.com for your application.
package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/connection"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/pdu"
)

var (
	ERROR_BROKEN_CONNECTION = errors.New("Auth: Broken connection")
)

type Auth struct {
	conn       *connection.Connection
	role       string
	roleSecret string
}

type nonceType struct {
	Data struct {
		Nonce string
	}
}

func New(role, roleSecret string) *Auth {
	auth := &Auth{
		role:       role,
		roleSecret: roleSecret,
	}
	return auth
}

func (auth *Auth) Authenticate(conn *connection.Connection) error {
	auth.conn = conn

	logger.Info("Auth: Starting authentication")
	action := "auth/handshake"
	body := json.RawMessage(`{ "method": "role_secret", "data": {"role": "` + auth.role + `"} }`)

	ch, err := auth.conn.SendAck(action, body)

	if err != nil {
		return err
	}

	handshake, ok := <-ch
	if !ok {
		return ERROR_BROKEN_CONNECTION
	}
	if pdu.GetResponseCode(handshake) != pdu.CODE_OK_REQUEST {
		return pdu.GetResponseError(handshake)
	}

	logger.Debug("Auth: Handshake response:", handshake.String())

	nonce := nonceType{}
	err = json.Unmarshal(handshake.Body, &nonce)
	if err != nil {
		return err
	}

	logger.Debug("Auth: Got nonce. Trying to authenticate")
	hash := GetHmacMD5(string(nonce.Data.Nonce), auth.roleSecret)

	action = "auth/authenticate"
	body = json.RawMessage(`{"method": "role_secret", "credentials": {"hash": "` + hash + `"} }`)
	ch, err = auth.conn.SendAck(action, body)

	authenticated, ok := <-ch
	if !ok {
		return ERROR_BROKEN_CONNECTION
	}
	if pdu.GetResponseCode(authenticated) != pdu.CODE_OK_REQUEST {
		return pdu.GetResponseError(authenticated)
	}
	logger.Info("Auth: Succesfully authenticated")

	return nil
}

// Gets Hmac hash for message using Secret key.
//
// Returns base64 encoded string
func GetHmacMD5(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(md5.New, key)
	h.Write([]byte(message))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
