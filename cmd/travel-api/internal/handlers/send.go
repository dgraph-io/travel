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

type send struct {
	Email
}

func (s *send) email(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var recipient struct {
		Email   string `validate:"email"`
		Subject string `validate:"required"`
	}
	if err := web.Decode(r, &recipient); err != nil {
		return errors.Wrap(err, "decoding recipient")
	}

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	to := []string{recipient.Email}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\nThis is the email body.\r\n", recipient.Email, recipient.Subject))
	addr := net.JoinHostPort(s.Host, s.Port)

	if err := smtp.SendMail(addr, auth, s.User, to, msg); err != nil {
		return errors.Wrap(err, "sending email")
	}

	return nil
}
