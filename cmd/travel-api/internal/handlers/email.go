package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dgraph-io/travel/internal/data"
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
	var request data.EmailRequest
	if err := web.Decode(r, &request); err != nil {
		return errors.Wrap(err, "decoding recipient")
	}

	// Add actual email code here.

	resp := data.EmailResponse{
		UserID:  request.UserID,
		Message: fmt.Sprintf("email sent to %q for node type %q with id %q", request.Email, request.NodeType, request.NodeID),
	}
	web.Respond(ctx, w, resp, http.StatusOK)

	return nil
}
