package db

import (
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/gophish/healthcheck/config"
)

func setupConfig(t *testing.T) {
	config.Config.DBName = "sqlite3"
	config.Config.DBPath = ":memory:"
	config.Config.MigrationsPath = "../db/sqlite3/migrations/"
	err := Setup()
	if err != nil {
		t.Fatalf("Failed setting up the database: %s", err.Error())
	}
}

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

func TestGetMessage(t *testing.T) {
	setupConfig(t)
	m := createMessage()
	err := PostMessage(m)

	got, err := GetMessage(m.MessageID)
	if err != nil {
		t.Fatalf("Unexpected error when getting message: %s", err.Error())
	}
	if got.ID != m.ID {
		t.Fatalf("Invalid message received. Expected ID %d Got %d", m.ID, got.ID)
	}
}

func TestInvalidGetMessage(t *testing.T) {
	setupConfig(t)
	_, err := GetMessage("InvalidID")
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("Unexpected error received when fetching invalid message. Expected %s Got %s", gorm.ErrRecordNotFound.Error(), err.Error())
	}
}
