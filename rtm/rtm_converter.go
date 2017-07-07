package rtm

import (
	"encoding/json"
	"fmt"
	"github.com/satori-com/satori-rtm-sdk-go/logger"
	"reflect"
	"strconv"
)

func (rtm *RTMClient) ConvertToRawJson(message interface{}) json.RawMessage {
	switch message.(type) {
	case string:
		encoded, _ := json.Marshal(message.(string))
		return json.RawMessage(encoded)

	case int, int8, int16, int32, int64:
		return json.RawMessage(strconv.FormatInt(reflect.ValueOf(message).Int(), 10))

	case uint, uint8, uint16, uint32, uint64:
		return json.RawMessage(strconv.FormatUint(reflect.ValueOf(message).Uint(), 10))

	case float32, float64:
		return json.RawMessage(strconv.FormatFloat(reflect.ValueOf(message).Float(), 'f', -1, 64))

	case bool:
		return json.RawMessage(strconv.FormatBool(message.(bool)))

	case json.RawMessage:
		return message.(json.RawMessage)

	case nil:
		return json.RawMessage(`null`)

	default:
		m, ok := json.Marshal(&message)
		if ok != nil {
			logger.Error(ERROR_UNSUPPORTED_TYPE)
			logger.Error(fmt.Errorf("Type: %s, Value: %v", reflect.TypeOf(message), message))
			return nil
		} else {
			return json.RawMessage(m)
		}
	}
}
