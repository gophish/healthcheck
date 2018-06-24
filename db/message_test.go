package db

import (
	"testing"
)

func createMessage() *Message {
	return &Message{
		Recipient:  "test@example.com",
		MailServer: "localhost",
	}
}

func TestSuccessfulMessageValidation(t *testing.T) {
	m := createMessage()
	err := m.Validate()
	if err != nil {
		t.Fatalf("Received unexpected error during messages validation: %s", err)
	}
}

func TestMessageNoRecipient(t *testing.T) {
	m := createMessage()
	m.Recipient = ""
	err := m.Validate()
	if err != ErrMissingRecipient {
		t.Fatalf("Didn't receive expected error with empty recipient. Got: %s", err)
	}
}

func TestMessageNoMailServer(t *testing.T) {
	m := createMessage()
	m.MailServer = ""
	err := m.Validate()
	if err != ErrMissingMailServer {
		t.Fatalf("Didn't receive expected error with empty mail server. Got: %s", err)
	}
}
