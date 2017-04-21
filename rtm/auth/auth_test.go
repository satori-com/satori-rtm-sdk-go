package auth

import (
	"testing"
)

func TestGetHmacMD5(t *testing.T) {
	expected := "GnXQHVYdUXdr82yZa4Jtaw=="
	hash := GetHmacMD5("hello", "world")

	if hash != expected {
		t.Fatalf("Hmac hashed do not match. Expected: %s. Actual: %s", expected, hash)
	}
}
