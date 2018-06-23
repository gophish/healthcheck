package util

import (
	"testing"
)

func TestDomainHashFromAddress(t *testing.T) {
	address := "test@example.com"
	expected := "0caaf24ab1a0c33440c06afe99df986365b0781f"
	got, err := DomainHashFromAddress(address)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if expected != got {
		t.Fatalf("Invalid response. Got: %s Expected %s", got, expected)
	}
}

func TestInvalidEmailAddress(t *testing.T) {
	address := "test"
	_, err := DomainHashFromAddress(address)
	if err == nil {
		t.Fatalf("Didn't receive expected error")
	}
}
