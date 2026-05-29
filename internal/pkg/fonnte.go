package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const fontteBaseURL = "https://api.fonnte.com"

// FonnteClient holds credentials and an HTTP client for the Fonnte API.
type FonnteClient struct {
	token  string
	http   *http.Client
	log    *logrus.Logger
}

// NewFonnteClient reads fonnte.token from viper config.
func NewFonnteClient(cfg *viper.Viper, log *logrus.Logger) *FonnteClient {
	return &FonnteClient{
		token: cfg.GetString("fonnte.token"),
		http:  &http.Client{Timeout: 15 * time.Second},
		log:   log,
	}
}

// ── Request / Response ────────────────────────────────────────────────────────

// SendRequest is the payload for the /send endpoint.
type FontteSendRequest struct {
	// Target phone number(s). Single: "08123456789".
	// Multiple: "08123456789,08987654321".
	Target string

	// Text message to send.
	Message string

	// Delay between messages in seconds (default "2").
	Delay string

	// CountryCode without "+", e.g. "62" for Indonesia.
	CountryCode string
}

// FontteResponse is the decoded JSON body from Fonnte.
type FontteResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Target  any    `json:"target,omitempty"`
	Process any    `json:"process,omitempty"`
}

// ── Methods ───────────────────────────────────────────────────────────────────

// Send sends a WhatsApp message via Fonnte /send endpoint.
func (f *FonnteClient) Send(req FontteSendRequest) (*FontteResponse, error) {
	if req.Delay == "" {
		req.Delay = "2"
	}
	if req.CountryCode == "" {
		req.CountryCode = "62"
	}

	body := &strings.Builder{}
	w := multipart.NewWriter(body)
	_ = w.WriteField("target", req.Target)
	_ = w.WriteField("message", req.Message)
	_ = w.WriteField("delay", req.Delay)
	_ = w.WriteField("countryCode", req.CountryCode)
	w.Close()

	httpReq, err := http.NewRequest(http.MethodPost, fontteBaseURL+"/send", strings.NewReader(body.String()))
	if err != nil {
		return nil, fmt.Errorf("fonnte: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", w.FormDataContentType())
	httpReq.Header.Set("Authorization", f.token)

	resp, err := f.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fonnte: send request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fonnte: read response: %w", err)
	}

	var result FontteResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("fonnte: decode response: %w", err)
	}

	if !result.Status {
		f.log.WithField("response", string(raw)).Warn("fonnte: message not delivered")
	}

	return &result, nil
}

// SendOne is a convenience wrapper for sending to a single number.
func (f *FonnteClient) SendOne(target, message string) (*FontteResponse, error) {
	return f.Send(FontteSendRequest{
		Target:  target,
		Message: message,
	})
}

// SendMany sends the same message to multiple numbers at once.
func (f *FonnteClient) SendMany(targets []string, message string) (*FontteResponse, error) {
	return f.Send(FontteSendRequest{
		Target:  strings.Join(targets, ","),
		Message: message,
	})
}
