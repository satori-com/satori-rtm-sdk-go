package pdu

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetResponseCode(t *testing.T) {
	query := RTMQuery{
		Action: "rtm/subscribe/ok",
	}
	if GetResponseCode(query) != CODE_OK_REQUEST {
		t.Error("[CODE_OK] Response code mismatch")
	}

	query = RTMQuery{
		Action: "rtm/subscribe/unknown",
	}

	query = RTMQuery{
		Action: "aaa/bbb/ccc",
	}
	if GetResponseCode(query) != CODE_BAD_REQUEST {
		t.Error("[CODE_BAD] Response code mismatch")
	}

	query = RTMQuery{
		Action: "", // Check empty action
	}
	if GetResponseCode(query) != CODE_BAD_REQUEST {
		t.Error("[CODE_BAD] Response code mismatch")
	}

	query = RTMQuery{
		Action: "rtm/search/data",
	}
	if GetResponseCode(query) != CODE_DATA_REQUEST {
		t.Error("[CODE_DATA_REQUEST] Response code mismatch")
	}
}

func TestGetResponseError(t *testing.T) {
	query := RTMQuery{
		Action: "rtm/subscribe/error",
		Body:   json.RawMessage("{error: \"code_123\", reason:\"Error reason\"}"),
	}
	if GetResponseCode(query) != CODE_ERROR_REQUEST {
		t.Error("[CODE_BAD] Response code mismatch")
	}
	if err := GetResponseError(query); err.Error() != "{error: \"code_123\", reason:\"Error reason\"}" {
		t.Error("Got wrong error reason")
	}

	query = RTMQuery{
		Action: "rtm/subscribe/ok",
	}
	if err := GetResponseError(query); err != nil {
		t.Error("Got response error message, but should not")
	}
}

func TestRTMQuery_String(t *testing.T) {
	query := RTMQuery{
		Action: "rtm/publish/ok",
		Body:   json.RawMessage("{\"Position\": \"12345.12345\"}"),
	}

	if fmt.Sprint(query) != "{\"action\":\"rtm/publish/ok\",\"body\":{\"Position\":\"12345.12345\"}}" {
		t.Fatal(query)
	}
}
