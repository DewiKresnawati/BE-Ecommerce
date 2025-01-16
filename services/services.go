package services

import (
	"github.com/veritrans/go-midtrans"
)

var (
	serverKey = "SB-Mid-server-B7ypxjlgwmDjouHtKbNkAjUo" // Ganti dengan Server Key Sandbox Anda
	clientKey = "SB-Mid-client-VbnC3VuCw9znazL7"         // Ganti dengan Client Key Sandbox Anda
)

// MidtransClient menginisialisasi client Midtrans
func MidtransClient() *midtrans.Client {
	c := midtrans.NewClient()
	c.ServerKey = serverKey
	c.ClientKey = clientKey
	c.APIEnvType = midtrans.Sandbox // Gunakan Sandbox untuk testing, ubah ke Production untuk live
	return &c                       // Mengembalikan pointer ke client
}