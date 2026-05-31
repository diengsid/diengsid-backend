package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const wapiBaseURL = "https://wapisender.id/api"

// WhatsAppSender is a common interface implemented by both WapiSenderClient
// and FonnteClient so they can be swapped without changing usecases.
type WhatsAppSender interface {
	SendOne(phone, message string) error
	SendMany(phones []string, message string) error
}

// ── WapiSenderClient ─────────────────────────────────────────────────────────

type WapiSenderClient struct {
	apiKey    string
	deviceKey string
	http      *http.Client
	log       *logrus.Logger
}

// NewWapiSenderClient reads wapisender.api_key and wapisender.device_key from config.
func NewWapiSenderClient(cfg *viper.Viper, log *logrus.Logger) *WapiSenderClient {
	return &WapiSenderClient{
		apiKey:    cfg.GetString("wapisender.api_key"),
		deviceKey: cfg.GetString("wapisender.device_key"),
		http:      &http.Client{Timeout: 15 * time.Second},
		log:       log,
	}
}

// ── Request / Response ────────────────────────────────────────────────────────

type wapiTextRequest struct {
	APIKey    string `json:"api_key"`
	DeviceKey string `json:"device_key"`
	To        string `json:"to"`
	Message   string `json:"message"`
}

type wapiResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// ── Methods ───────────────────────────────────────────────────────────────────

func (c *WapiSenderClient) send(phone, message string) error {
	payload := wapiTextRequest{
		APIKey:    c.apiKey,
		DeviceKey: c.deviceKey,
		To:        NormalizePhone(phone),
		Message:   message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wapisender: marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, wapiBaseURL+"/message/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wapisender: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("wapisender: send request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wapisender: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var result wapiResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("wapisender: decode response: %w (body: %s)", err, strings.TrimSpace(string(raw)))
	}

	if !result.Status {
		c.log.WithFields(logrus.Fields{
			"phone":    phone,
			"response": string(raw),
		}).Warn("wapisender: message not delivered")
		return fmt.Errorf("wapisender: %s", result.Message)
	}

	c.log.WithField("phone", phone).Info("wapisender: message sent")
	return nil
}

// SendOne sends a WhatsApp text message to a single number.
func (c *WapiSenderClient) SendOne(phone, message string) error {
	return c.send(phone, message)
}

// SendMany sends the same message to multiple numbers sequentially.
func (c *WapiSenderClient) SendMany(phones []string, message string) error {
	var errs []string
	for _, p := range phones {
		if err := c.send(p, message); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("wapisender: some messages failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// normalizePhone converts a local Indonesian number (08xxx) to international
// format (628xxx) expected by WAPISender.
func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "0") {
		return "62" + phone[1:]
	}
	if strings.HasPrefix(phone, "+") {
		return phone[1:]
	}
	return phone
}
