package db

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/healthcheck/config"
	"github.com/gophish/healthcheck/template"
	"github.com/gophish/healthcheck/util"
)

const (
	// DefaultSender is the sender part of the email address used to send messages
	DefaultSender = "no-reply"

	// DefaultSenderName is the name part of the email address used to send
	// messages
	DefaultSenderName = "Gophish Healthcheck"

	// DefaultSubject is the default subject used when sending messages
	DefaultSubject = "Gophish Healthcheck - Test Email"

	// DefaultSMTPPort is the default SMTP port used when making outbound SMTP
	// connections
	DefaultSMTPPort = 25

	// MessageIDLength is the number of bytes to use when generated message IDs
	MessageIDLength = 16

	// HardFail is the constant value used to indicate the specific configuration
	// should fail
	HardFail = "hardfail"
	// None is the constant value used to indicate the specific configuration
	// should not be present (e.g. no DKIM signing at all)
	None = "none"
	// SoftFail is the constant value that indicates the certain configuration
	// option (if supported) should softfail (e.g. spf softfail).
	SoftFail = "softfail"
	// Pass is the constant value used to indicate the specific configuration
	// value should be valid
	Pass = "pass"
	// Reject indicates that the DMARC response should have the policy set to
	// reject
	Reject = "reject"
	// Quarantine indicates that the DMARC response should have the policy set
	// to quarantine
	Quarantine = "quarantine"
	// Neutral indicates that the SPF response should have a neutral enforcement
	Neutral = "neutral"
)

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

// MessageConfiguration is the configuration for the outbound message.
type MessageConfiguration struct {
	SPF   string `json:"spf"`
	DKIM  string `json:"dkim"`
	DMARC string `json:"dmarc"`
	MX    string `json:"mx"`
}

// Message is the base struct for handling per-message information.
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

// Validate ensures the message is correctly formatted with all the necessary
// fields.
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

// Backoff simply errors out the message, since we don't handle exponential
// backoffs yet.
func (m *Message) Backoff(reason error) error {
	m.Successful = false
	m.ErrorMessage = reason.Error()
	m.ErrorChan <- reason
	return db.Save(m).Error
}

// Error errors out the message.
func (m *Message) Error(err error) error {
	m.Successful = false
	m.ErrorMessage = err.Error()
	m.ErrorChan <- err
	return db.Save(m).Error
}

// Success saves the message as having been sent successfully.
func (m *Message) Success() error {
	m.Successful = true
	m.ErrorChan <- nil
	return db.Save(m).Error
}

// Generate creates a gomail.Message instance from the provided message.
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

// GetDialer creates a mailer.Dialer from the message configuration.
func (m *Message) GetDialer() (mailer.Dialer, error) {
	port := DefaultSMTPPort
	hp := strings.Split(m.MailServer, ":")
	if len(hp) == 2 {
		port, _ = strconv.Atoi(hp[1])
	}
	d := &Dialer{
		&gomail.Dialer{
			Host: hp[0],
			Port: port,
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
