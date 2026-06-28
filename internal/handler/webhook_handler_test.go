package handler

import (
	"testing"
)

func TestVerifySignature_Valid(t *testing.T) {
	payload := []byte(`{"event":"STOCK_UPDATED","data":{}}`)
	secret := "my-webhook-hmac-secret"
	sig := "sha256=" + computeHMAC(payload, secret)

	if !verifySignature(payload, sig, secret) {
		t.Error("valid signature should verify as true")
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	payload := []byte(`{"event":"STOCK_UPDATED","data":{}}`)

	if verifySignature(payload, "sha256=badhash", "secret") {
		t.Error("invalid signature should verify as false")
	}
}

func TestVerifySignature_TamperedPayload(t *testing.T) {
	secret := "my-webhook-hmac-secret"
	original := []byte(`{"event":"STOCK_UPDATED","data":{}}`)
	sig := "sha256=" + computeHMAC(original, secret)

	tampered := []byte(`{"event":"STOCK_UPDATED","data":{},"extra":"injected"}`)
	if verifySignature(tampered, sig, secret) {
		t.Error("tampered payload should fail signature check")
	}
}

func TestComputeHMAC_Deterministic(t *testing.T) {
	payload := []byte("test payload")
	secret := "secret"

	h1 := computeHMAC(payload, secret)
	h2 := computeHMAC(payload, secret)
	if h1 != h2 {
		t.Errorf("HMAC is not deterministic: %s != %s", h1, h2)
	}
}

func TestComputeHMAC_DifferentSecrets(t *testing.T) {
	payload := []byte("test payload")

	h1 := computeHMAC(payload, "secret1")
	h2 := computeHMAC(payload, "secret2")
	if h1 == h2 {
		t.Error("different secrets should produce different HMACs")
	}
}
