package pkg

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const (
	dokuSandboxBase    = "https://api-sandbox.doku.com"
	dokuProductionBase = "https://api.doku.com"
	dokuPaymentPath    = "/checkout/v1/payment"
)

type DokuClient struct {
	ClientID        string
	SecretKey       string
	BaseURL         string
	SuccessURL      string
	FailedURL       string
	NotificationURL string
}

func NewDokuClient(cfg *viper.Viper) *DokuClient {
	base := dokuProductionBase
	if cfg.GetBool("doku.is_sandbox") {
		base = dokuSandboxBase
	}
	return &DokuClient{
		ClientID:        cfg.GetString("doku.client_id"),
		SecretKey:       cfg.GetString("doku.secret_key"),
		BaseURL:         base,
		SuccessURL:      cfg.GetString("doku.success_url"),
		FailedURL:       cfg.GetString("doku.failed_url"),
		NotificationURL: cfg.GetString("doku.notification_url"),
	}
}

// ── Request / Response types ──────────────────────────────────────────────────

type DokuPaymentRequest struct {
	Client struct {
		ID string `json:"id"`
	} `json:"client"`
	Order struct {
		InvoiceNumber     string         `json:"invoice_number"`
		LineItems         []DokuLineItem `json:"line_items"`
		Amount            int64          `json:"amount"`
		Currency          string         `json:"currency"`
		CallbackURL       string         `json:"callback_url"`
		CallbackURLCancel string         `json:"callback_url_cancel"`
	} `json:"order"`
	Payment struct {
		PaymentDueDate int `json:"payment_due_date"` // minutes
	} `json:"payment"`
	Customer       DokuCustomer `json:"customer"`
	AdditionalInfo struct {
		OverrideNotificationURL string `json:"override_notification_url,omitempty"`
	} `json:"additional_info,omitempty"`
}

type DokuLineItem struct {
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
}

type DokuCustomer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type DokuPaymentResponse struct {
	Response struct {
		Payment struct {
			URL        string `json:"url"`
			Date       string `json:"date"`
			ExpiryDate string `json:"expiry_date"`
		} `json:"payment"`
	} `json:"response"`
}

// ── CreatePayment ─────────────────────────────────────────────────────────────

func (c *DokuClient) CreatePayment(invoiceNumber string, amount int64, description string, customer DokuCustomer) (string, error) {
	req := DokuPaymentRequest{}
	req.Client.ID = c.ClientID
	req.Order.InvoiceNumber = invoiceNumber
	req.Order.Amount = amount
	req.Order.Currency = "IDR"
	req.Order.CallbackURL = fmt.Sprintf("%s&invoice=%s", c.SuccessURL, invoiceNumber)
	req.Order.CallbackURLCancel = fmt.Sprintf("%s&invoice=%s", c.FailedURL, invoiceNumber)
	req.Order.LineItems = []DokuLineItem{
		{Name: description, Price: amount, Quantity: 1},
	}
	req.Payment.PaymentDueDate = 60 // 60 minutes
	req.Customer = customer
	if c.NotificationURL != "" {
		req.AdditionalInfo.OverrideNotificationURL = c.NotificationURL
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal doku request: %w", err)
	}

	requestID := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	signature := c.sign(requestID, dokuPaymentPath, string(body), timestamp)

	httpReq, err := http.NewRequest("POST", c.BaseURL+dokuPaymentPath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Client-Id", c.ClientID)
	httpReq.Header.Set("Request-Id", requestID)
	httpReq.Header.Set("Request-Timestamp", timestamp)
	httpReq.Header.Set("Signature", signature)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("doku request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("doku error %d: %s", resp.StatusCode, string(respBody))
	}

	var dokuResp DokuPaymentResponse
	if err := json.Unmarshal(respBody, &dokuResp); err != nil {
		return "", fmt.Errorf("unmarshal doku response: %w", err)
	}

	if dokuResp.Response.Payment.URL == "" {
		return "", fmt.Errorf("doku returned empty payment url: %s", string(respBody))
	}

	return dokuResp.Response.Payment.URL, nil
}

// ── VerifyNotification ────────────────────────────────────────────────────────

func (c *DokuClient) VerifyNotification(requestID, requestTimestamp, notifyBody, signatureHeader string) bool {
	expected := c.sign(requestID, "/api/payment/notify", notifyBody, requestTimestamp)
	return hmac.Equal([]byte(expected), []byte(signatureHeader))
}

// ── sign ──────────────────────────────────────────────────────────────────────

func (c *DokuClient) sign(requestID, requestTarget, body, timestamp string) string {
	h := sha256.New()
	h.Write([]byte(body))
	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	components := []string{
		"Client-Id:" + c.ClientID,
		"Request-Id:" + requestID,
		"Request-Timestamp:" + timestamp,
		"Request-Target:" + requestTarget,
		"Digest:" + digest,
	}
	stringToSign := strings.Join(components, "\n")

	mac := hmac.New(sha256.New, []byte(c.SecretKey))
	mac.Write([]byte(stringToSign))
	return "HMACSHA256=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
