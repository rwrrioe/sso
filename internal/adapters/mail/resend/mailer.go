package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/resend/resend-go/v3"
	"github.com/rwrrioe/sso/internal/lib/logger/sl"
)

const url = "https://api.resend.com/emails"

type Options struct {
	From    string
	Name    string
	Subject string
}

type ResendAPI struct {
	apiKey  string
	log     *slog.Logger
	options *Options
}

func (api *ResendAPI) SendCode(ctx context.Context, name, email, code string) error {
	const op = "resend.SendCode"

	payload := &resend.SendEmailRequest{
		From:    api.options.From,
		To:      []string{email},
		Subject: api.options.Subject,
		Html:    "TODO add HTML",
	}

	b, err := json.Marshal(&payload)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	mailReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	mailReq.Header.Set("Authorization", "Bearer "+api.apiKey)
	mailReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(mailReq)

	if err != nil {
		api.log.Error("failed to send mail via resend", sl.Err(err))
		return fmt.Errorf("%s:%w", op, err)
	}

	if resp != nil {
		defer resp.Body.Close()

		b, _ = io.ReadAll(resp.Body)
		var result Response

		err = json.Unmarshal(b, &result)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("%s: resend error: %s %s", op, result.Name, result.Error)
		}

		api.log.Info("successfully sent; id:", result.Id)

	}

	return nil
}
