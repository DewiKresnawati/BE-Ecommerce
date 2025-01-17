package handler

import (
	"time"

	"be_ecommerce/services"

	"github.com/gofiber/fiber/v2"
	"github.com/veritrans/go-midtrans"
)

// Struktur untuk informasi customer (update sesuai dengan format Midtrans)
type CustomerDetails struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// Struktur request pembayaran yang hanya membutuhkan gross_amount dan customer_details
type PaymentRequest struct {
	GrossAmount     int64           `json:"gross_amount"`
	CustomerDetails CustomerDetails `json:"customer_details"`
}

func CreatePayment(c *fiber.Ctx) error {
	var req PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// Generate order ID jika tidak diberikan
	orderID := "order-" + time.Now().Format("20060102150405")

	// Dapatkan client dari service
	midtransClient := services.MidtransClient()

	// Prepare Snap transaction request
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: req.GrossAmount,
		},
		CustomerDetail: &midtrans.CustDetail{
			FName: req.CustomerDetails.FirstName,
			Email: req.CustomerDetails.Email,
			Phone: req.CustomerDetails.Phone,
		},
	}

	// Dapatkan token Snap
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create payment",
		})
	}

	// Berikan respons token dan redirect URL
	response := fiber.Map{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
	}

	return c.JSON(response)
}
