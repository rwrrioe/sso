package resend

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/resend/resend-go/v3"
	"github.com/rwrrioe/sso/internal/usecase/code"
)

const resendURL = "https://api.resend.com/emails"

type Config struct {
	From string
	Name string
}

type ResendAPI struct {
	log     *slog.Logger
	options *Config
	client  *resend.Client
}

func New(
	log *slog.Logger,
	cfg *Config,
) code.MailProvider {
	apiKey := os.Getenv("RESEND_API_KEY")

	cl := resend.NewClient(apiKey)

	return &ResendAPI{
		log:     log,
		options: cfg,
		client:  cl,
	}
}

func (api *ResendAPI) SendCode(
	ctx context.Context,
	email, code string,
	codeType code.CodeType,
) error {
	const op = "resend.SendCode"

	params := &resend.SendEmailRequest{
		From:    api.options.From,
		To:      []string{email},
		Subject: templates[codeType].Subject,
		Html:    fmt.Sprintf(templates[codeType].HTML, code),
	}

	_, err := api.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		api.log.Error("failed to send code", err.Error())
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

type MailTemplate struct {
	Subject string
	HTML    string
}

var templates = map[code.CodeType]MailTemplate{
	code.TypeResetCode: {
		Subject: "Password Reset",
		HTML:    "<p>Your reset code: <b>%s</b></p>",
	},
	code.TypeEmailVerificationCode: {
		Subject: "Email Verification",
		HTML:    "<p>Your verification code: <b>%s</b></p>",
	},
	code.Type2FACode: {
		Subject: "Two-Factor Authentication",
		HTML:    "<p>Your 2FA code: <b>%s</b></p>",
	},
}
