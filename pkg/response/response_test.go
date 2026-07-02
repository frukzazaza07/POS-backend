package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"pos-backend/pkg/response"
)

type genericResp struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decode(t *testing.T, resp *http.Response) genericResp {
	t.Helper()
	var body genericResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func TestSuccess(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return response.Success(c, "ok", fiber.Map{"key": "value"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	body := decode(t, resp)
	if body.Status != "success" {
		t.Errorf("want status=success, got %q", body.Status)
	}
}

func TestCreated(t *testing.T) {
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		return response.Created(c, "created", fiber.Map{"id": "123"})
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want 201, got %d", resp.StatusCode)
	}
	body := decode(t, resp)
	if body.Status != "success" {
		t.Errorf("want status=success, got %q", body.Status)
	}
}

func TestError(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusBadRequest, "bad input")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
	body := decode(t, resp)
	if body.Status != "error" {
		t.Errorf("want status=error, got %q", body.Status)
	}
	if body.Message != "bad input" {
		t.Errorf("want message=bad input, got %q", body.Message)
	}
}
