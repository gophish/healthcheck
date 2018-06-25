package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/healthcheck/config"
	"github.com/gophish/healthcheck/template"
	"github.com/gophish/healthcheck/util"
)

// DefaultSender is the sender part of the email address used to send messages
const DefaultSender = "no-reply"

// DefaultSenderName is the name part of the email address used to send
// messages
const DefaultSenderName = "Gophish Healthcheck"

// DefaultSubject is the default subject used when sending messages
const DefaultSubject = "Gophish Healthcheck - Test Email"

// DefaultSMTPPort is the default SMTP port used when making outbound SMTP
// connections
const DefaultSMTPPort = 25

// MessageIDLength is the number of bytes to use when generated message IDs
const MessageIDLength = 16

// ErrMissingMailServer occurs when a message is received without specifying
// a valid mail server.
var ErrMissingMailServer = errors.New("no mail server specified")

// ErrMissingRecipient occurs when a message is received without specifying
// a valid recipient
var ErrMissingRecipient = errors.New("no recipient specified")

// Dialer is a wrapper around a standard gomail.Dialer in order
// to implement the mailer.Dialer interface. This allows us to better
// separate the mailer package as opposed to forcing a connection
// between mailer and gomail.
type Dialer struct {
	*gomail.Dialer
}

// Dial wraps the gomail dialer's Dial command
func (d *Dialer) Dial() (mailer.Sender, error) {
	return d.Dialer.Dial()
}

type MessageConfiguration struct {
	SPF   string `json:"spf"`
	DKIM  string `json:"dkim"`
	DMARC string `json:"dmarc"`
	MX    bool   `json:"mx"`
}

type Message struct {
	ID           uint         `gorm:"primary_key" json:"id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at,omitempty"`
	Recipient    string       `gorm:"-" json:"recipient"`
	MailServer   string       `json:"mail_server"`
	MessageID    string       `json:"message_id"`
	DomainHash   string       `json:"domain_hash"`
	Successful   bool         `json:"successful"`
	ErrorMessage string       `json:"error_message"`
	ErrorChan    chan (error) `gorm:"-" json:"-"`

	MessageConfiguration `gorm:"embedded" json:"configuration"`
}

func (m *Message) Validate() error {
	if m.Recipient == "" {
		return ErrMissingRecipient
	}
	if m.MailServer == "" {
		return ErrMissingMailServer
	}
	return nil
}

func (m *Message) generateFromAddress() string {
	return fmt.Sprintf("\"%s\" <%s@%s.%s>", DefaultSenderName, DefaultSender, m.MessageID, config.Config.EmailHostname)
}

func (m *Message) Backoff(reason error) error {
	m.Successful = false
	m.ErrorMessage = reason.Error()
	m.ErrorChan <- reason
	return db.Save(m).Error
}

func (m *Message) Error(err error) error {
	m.Successful = false
	m.ErrorMessage = err.Error()
	m.ErrorChan <- err
	return db.Save(m).Error
}

func (m *Message) Success() error {
	m.Successful = true
	m.ErrorChan <- nil
	return db.Save(m).Error
}

func (m *Message) Generate(msg *gomail.Message) error {
	msg.SetHeader("From", m.generateFromAddress())
	msg.SetHeader("To", m.Recipient)
	msg.SetHeader("Subject", DefaultSubject)
	// Sign with DKIM if needed
	if m.MessageConfiguration.DKIM != "none" {

	}

	text, err := template.ExecuteTemplate(template.TextTemplate, m)
	if err != nil {
		return err
	}
	msg.SetBody("text/plain", text)

	html, err := template.ExecuteTemplate(template.HTMLTemplate, m)
	if err != nil {
		return err
	}
	msg.SetBody("text/html", html)

	// Handle attachments in future versions
	return nil
}

func (m *Message) GetDialer() (mailer.Dialer, error) {
	d := &Dialer{
		&gomail.Dialer{
			Host: m.MailServer,
			Port: DefaultSMTPPort,
		},
	}
	return d, nil
}

// GetMessage retrieves a message by ID from the database
func GetMessage(id string) (*Message, error) {
	message := &Message{}
	err := db.Where("message_id=?", id).First(message).Error
	return message, err
}

// PostMessage saves a message instance into the database
func PostMessage(m *Message) error {
	for {
		// Generate a random ID for the message
		m.MessageID = util.GenerateSecureID(MessageIDLength)
		// Verify the ID doesn't already exist
		_, err := GetMessage(m.MessageID)
		if err == gorm.ErrRecordNotFound {
			break
		}
	}
	err := db.Save(m).Error
	return err
}
