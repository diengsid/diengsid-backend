package model

type CreatePaymentResponse struct {
	PaymentURL string `json:"payment_url"`
	InvoiceNo  string `json:"invoice_no"`
}

type PaymentInfoResponse struct {
	InvoiceNo  string `json:"invoice_no"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	PaymentURL string `json:"payment_url"`
}

// DokuNotification is the webhook body DOKU POSTs to /api/payment/notify.
type DokuNotification struct {
	Order struct {
		InvoiceNumber string `json:"invoice_number"`
		Amount        int64  `json:"amount"`
	} `json:"order"`
	Transaction struct {
		Status            string `json:"status"`
		Date              string `json:"date"`
		OriginalRequestID string `json:"original_request_id"`
	} `json:"transaction"`
}
