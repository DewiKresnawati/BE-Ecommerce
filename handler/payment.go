package handler

import (
	
	"net/http"
	"time"
	"fmt"

	"github.com/gofiber/fiber/v2"
	`be_ecommerce/model`    
	`be_ecommerce/services`

	`github.com/veritrans/go-midtrans`
)

func CreatePayment(c *fiber.Ctx) error {
	var req model.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
	}

	orderID := "order-" + time.Now().Format("20060102150405")

	midtransClient := services.MidtransClient()

	snapGateway := midtrans.SnapGateway{Client: *midtransClient}
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: req.Amount,
		},
		Items: &[]midtrans.ItemDetail{},
	}

	for _, item := range req.Items {
		*snapReq.Items = append(*snapReq.Items, midtrans.ItemDetail{
			Name:  item.Name,
			Qty:   int32(item.Quantity),
			Price: item.Price,
		})
	}

	// Tambahkan log untuk memastikan snapReq
	fmt.Printf("Snap Request: %+v\n", snapReq)

	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		// Tambahkan log error detail dari Midtrans
		fmt.Printf("Midtrans Error: %+v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create payment",
			"details": err.Error(),
		})
	}

	fmt.Printf("Snap Response: %+v\n", snapResp)

	return c.JSON(fiber.Map{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
	})
}