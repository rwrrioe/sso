package resend

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v3"
	"github.com/rwrrioe/sso/internal/usecase/code"
)

const resendURL = "https://api.resend.com/emails"

type Options struct {
	From    string
	Name    string
	Subject string
	HTML    string
}

type ResendAPI struct {
	log     *slog.Logger
	options *Options
	client  *resend.Client
}

func NewResendAPI(
	log *slog.Logger,
	options *Options,
	apiKey string,
) code.MailProvider {
	cl := resend.NewClient(apiKey)

	return &ResendAPI{
		log:     log,
		options: options,
		client:  cl,
	}
}

func (api *ResendAPI) SendCode(
	ctx context.Context, email, code string) error {
	const op = "resend.SendCode"

	params := &resend.SendEmailRequest{
		From:    api.options.From,
		To:      []string{email},
		Subject: api.options.Subject,
		Html:    fmt.Sprintf(api.options.HTML, code),
	}

	_, err := api.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		api.log.Error("failed to send code", err.Error())
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
