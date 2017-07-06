package rtm

import (
	"encoding/json"
	"testing"
)

var rtm = &RTMClient{}

func TestRTM_ConvertToRawJson_String(t *testing.T) {
	actual := rtm.ConvertToRawJson("test")
	if string(actual) != "\"test\"" {
		t.Error("Error to convert to string")
	}

	actual = rtm.ConvertToRawJson("\"[hello']")
	if string(actual) != "\"\\\"[hello']\"" {
		t.Error("Error to convert to string")
	}
}

func TestRTM_ConvertToRawJson_String_Ints(t *testing.T) {
	var i int = -1
	actual := rtm.ConvertToRawJson(i)
	if string(actual) != "-1" {
		t.Error("Error to convert to int")
	}

	var ui uint = 1
	actual = rtm.ConvertToRawJson(ui)
	if string(actual) != "1" {
		t.Error("Error to convert to uint")
	}

	var i8 int8 = -8
	actual = rtm.ConvertToRawJson(i8)
	if string(actual) != "-8" {
		t.Error("Error to convert to int8")
	}

	var i32 int32 = 32
	actual = rtm.ConvertToRawJson(i32)
	if string(actual) != "32" {
		t.Error("Error to convert to int32")
	}

	var i64 int64 = -64
	actual = rtm.ConvertToRawJson(i64)
	if string(actual) != "-64" {
		t.Error("Error to convert to int64")
	}

	var ui8 uint8 = 8
	actual = rtm.ConvertToRawJson(ui8)
	if string(actual) != "8" {
		t.Error("Error to convert to uint8")
	}

	var ui32 uint32 = 32
	actual = rtm.ConvertToRawJson(ui32)
	if string(actual) != "32" {
		t.Error("Error to convert to uint32")
	}

	var ui64 uint64 = 64
	actual = rtm.ConvertToRawJson(ui64)
	if string(actual) != "64" {
		t.Error("Error to convert to uint64")
	}
}

func TestRTM_ConvertToRawJson_Floats(t *testing.T) {
	var f32 float32 = 1.123
	actual := rtm.ConvertToRawJson(f32)
	if string(actual)[:5] != "1.123" {
		t.Error("Error to convert to float32")
	}

	var f64 float64 = -2.123
	actual = rtm.ConvertToRawJson(f64)
	if string(actual)[:5] != "-2.12" {
		t.Error("Error to convert to float64")
	}
}

func TestRTM_ConvertToRawJson_Complexes(t *testing.T) {
	var c64 complex64 = 1.123
	actual := rtm.ConvertToRawJson(c64)
	if actual != nil {
		t.Error("complex64 should not be converted")
	}
}

func TestRTM_ConvertToRawJson_Bool(t *testing.T) {
	var tbool bool = true
	var fbool bool = false

	actual := rtm.ConvertToRawJson(tbool)
	if string(actual) != "true" {
		t.Error("Error to convert to bool (true)")
	}

	actual = rtm.ConvertToRawJson(fbool)
	if string(actual) != "false" {
		t.Error("Error to convert to bool (false)")
	}
}

func TestRTM_ConvertToRawJson_Raw(t *testing.T) {
	message := json.RawMessage("{a: 3, b: {b: true}}")
	actual := rtm.ConvertToRawJson(message)
	if string(actual) != "{a: 3, b: {b: true}}" {
		t.Error("Error to convert to RawJson")
	}
}

func TestRTM_ConvertToRawJson_Structs(t *testing.T) {
	type A struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		HaveCar bool   `json:"have_car"`
	}

	obj := A{
		Name:    "Gopher",
		Age:     15,
		HaveCar: false,
	}
	actual := rtm.ConvertToRawJson(obj)
	if string(actual) != "{\"name\":\"Gopher\",\"age\":15,\"have_car\":false}" {
		t.Error("Error to convert Struct")
	}
}

func TestRTM_ConvertToRawJson_Nil(t *testing.T) {
	actual := rtm.ConvertToRawJson(nil)
	if string(actual) != "null" {
		t.Error("Error to convert nil")
	}
}

func TestRTM_ConvertToRawJson_CustomType(t *testing.T) {
	type Custom int
	type Custom2 Custom
	var a Custom2 = 5
	actual := rtm.ConvertToRawJson(a)
	if string(actual) != "5" {
		t.Error("Error to convert CustomType")
	}
}

func TestRTM_ConvertToRawJson_List(t *testing.T) {
	list := []string{"a", "b", "c"}
	actual := rtm.ConvertToRawJson(list)
	if string(actual) != "[\"a\",\"b\",\"c\"]" {
		t.Error("Error to convert []string")
	}
}

func TestRTM_ConvertToRawJson_Map(t *testing.T) {
	m := make(map[string]int)
	m["test"] = 1
	m["aaa"] = 2
	actual := rtm.ConvertToRawJson(m)
	if string(actual) != "{\"aaa\":2,\"test\":1}" {
		t.Error("Error to convert map[string]int")
	}
}

func TestRTM_ConvertToRawJson_ComplexMap(t *testing.T) {
	m := make(map[int]func() int)
	m[1] = func() int { return 1 }
	m[2] = func() int { return 2 }
	actual := rtm.ConvertToRawJson(m)

	if actual != nil {
		t.Error("Complex map shuold not be converted")
	}
}
