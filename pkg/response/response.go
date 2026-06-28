package response

import "github.com/gofiber/fiber/v2"

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{Status: "success", Data: data})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{Status: "success", Data: data})
}

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{Status: "error", Message: message})
}
