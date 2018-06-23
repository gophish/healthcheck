package mail

import (
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/healthcheck/db"
)

// SendEmail sends the email to the mailer's queue and returns any error
// encountered
func SendEmail(m *db.Message) error {
	go func() {
		mailer.Mailer.Queue <- []mailer.Mail{m}
	}()
	return <-m.ErrorChan
}
