package pdu

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Gets PDU response code
func GetResponseCode(query RTMQuery) int {
	lastSlashIndex := strings.LastIndex(query.Action, "/")
	if lastSlashIndex < 0 {
		return CODE_BAD_REQUEST
	}
	switch query.Action[lastSlashIndex:] {
	case "/ok":
		return CODE_OK_REQUEST
	case "/error":
		return CODE_ERROR_REQUEST
	default:
		return CODE_BAD_REQUEST
	}
}

// Gets error as type "error" from PDU
func GetResponseError(response RTMQuery) error {
	responseCode := GetResponseCode(response)
	if responseCode == CODE_ERROR_REQUEST {
		return fmt.Errorf(string(response.Body))
	}

	return nil
}

// Stringify query. Returns struct as JSON
func (rq RTMQuery) String() string {
	message, _ := json.Marshal(&rq)
	return string(message)
}
