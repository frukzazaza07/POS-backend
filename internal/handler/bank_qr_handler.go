package handler

import (
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
	"pos-backend/pkg/promptpay"
	"pos-backend/pkg/response"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/gofiber/fiber/v2"
)

type BankQRHandler struct {
	repo *repository.BankQRConfigRepository
}

func NewBankQRHandler(repo *repository.BankQRConfigRepository) *BankQRHandler {
	return &BankQRHandler{repo: repo}
}

func (h *BankQRHandler) Get(c *fiber.Ctx) error {
	config, err := h.repo.Get()
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "bank QR config not configured")
	}
	return response.Success(c, config)
}

func (h *BankQRHandler) Upsert(c *fiber.Ctx) error {
	var config models.BankQRConfig
	if err := c.BodyParser(&config); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if config.BankName == "" || config.AccountName == "" || config.AccountNumber == "" {
		return response.Error(c, fiber.StatusBadRequest, "bank_name, account_name, and account_number are required")
	}
	config.IsActive = true
	if err := h.repo.Upsert(&config); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	result, _ := h.repo.Get()
	return response.Success(c, result)
}

// GetQRCode generates a PromptPay QR code PNG from the stored config.
// Optional query param ?amount=185.00 embeds the amount (dynamic QR).
// Returns image/png directly — use as <img src="/api/v1/config/bank-qr/qrcode" />.
func (h *BankQRHandler) GetQRCode(c *fiber.Ctx) error {
	config, err := h.repo.Get()
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "bank QR config not configured")
	}
	if config.PromptPayID == "" {
		return response.Error(c, fiber.StatusUnprocessableEntity, "promptpay_id not set in bank QR config — update it via PUT /api/v1/config/bank-qr")
	}

	amount := c.QueryFloat("amount", 0)
	payload := promptpay.BuildPayload(config.PromptPayID, amount)

	png, err := qrcode.Encode(payload, qrcode.Medium, 256)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to generate QR code")
	}

	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", "no-store")
	return c.Send(png)
}
