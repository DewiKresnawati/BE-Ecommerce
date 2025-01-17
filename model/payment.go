package model

// Struktur untuk informasi barang
type ItemDetails struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"`
}

// Struktur request pembayaran yang menyesuaikan dengan format baru
type PaymentRequest struct {
	Shipping string        `json:"shipping"`
	Amount   int64         `json:"amount"`
	Items    []ItemDetails `json:"items"`
}