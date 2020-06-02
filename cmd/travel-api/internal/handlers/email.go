package handlers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/smtp"

	"github.com/dgraph-io/travel/internal/platform/web"
	"github.com/pkg/errors"
)

// EmailConfig defines the configuration required to send an email.
type EmailConfig struct {
	User     string
	Password string
	Host     string
	Port     string
}

type email struct {
	EmailConfig
}

func (e *email) send(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var recipient struct {
		Email   string `validate:"email"`
		Subject string `validate:"required"`
	}
	if err := web.Decode(r, &recipient); err != nil {
		return errors.Wrap(err, "decoding recipient")
	}

	auth := smtp.PlainAuth("", e.User, e.Password, e.Host)
	to := []string{recipient.Email}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\nThis is the email body.\r\n", recipient.Email, recipient.Subject))
	addr := net.JoinHostPort(e.Host, e.Port)

	if err := smtp.SendMail(addr, auth, e.User, to, msg); err != nil {
		return errors.Wrap(err, "sending email")
	}

	return nil
}
