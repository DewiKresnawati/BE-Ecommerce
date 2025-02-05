package handler

import (
	"be_ecommerce/model"
	"io"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func GetIPaddress() string {

	resp, err := http.Get("https://icanhazip.com/")

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

func GetHome(c *fiber.Ctx) error {
	var resp model.Response
	resp.Response = GetIPaddress()
	return c.JSON(fiber.Map{"message": "berhasil", "data": resp.Response})
}
