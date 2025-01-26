package model

// Struktur request untuk pembayaran
type PaymentRequest struct {
	Shipping string      `json:"shipping"`
	Amount   int64       `json:"amount"`
	Items    []CartItem `json:"items"`
}
