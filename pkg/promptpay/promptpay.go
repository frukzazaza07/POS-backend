// Package promptpay builds EMVCo PromptPay QR payloads (Thai standard).
// Supports phone numbers (10-digit starting with 0) and national/tax IDs (13 digits).
// Optional amount produces a dynamic QR (initiation method 12); omitting amount produces a static QR (11).
package promptpay

import (
	"fmt"
	"strings"
)

const promptPayAppID = "A000000677010111"

// NormalizeID converts a Thai phone number or national/tax ID into the
// format expected by the PromptPay payload.
//   - 10-digit phone (0XXXXXXXXX) → 00660XXXXXXXXX (13 chars)
//   - 13-digit national/tax ID    → used as-is
func NormalizeID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.ReplaceAll(id, "-", "")
	id = strings.ReplaceAll(id, " ", "")
	if len(id) == 10 && id[0] == '0' {
		return "0066" + id[1:]
	}
	return id
}

// BuildPayload returns the full EMVCo PromptPay QR string (including CRC).
// amount <= 0 → static QR (no amount). amount > 0 → dynamic QR with amount in THB.
func BuildPayload(promptPayID string, amount float64) string {
	normalizedID := NormalizeID(promptPayID)

	merchantAccInfo := tlv("00", promptPayAppID) + tlv("01", normalizedID)

	initiationMethod := "11" // static
	if amount > 0 {
		initiationMethod = "12" // dynamic (with amount)
	}

	payload := tlv("00", "01") +
		tlv("01", initiationMethod) +
		tlv("29", merchantAccInfo) +
		tlv("53", "764") + // THB
		tlv("58", "TH")

	if amount > 0 {
		payload += tlv("54", fmt.Sprintf("%.2f", amount))
	}

	payload += "6304" // CRC tag — checksum appended after
	return payload + crc16(payload)
}

// tlv encodes a single EMVCo tag-length-value element.
func tlv(tag, value string) string {
	return fmt.Sprintf("%s%02d%s", tag, len(value), value)
}

// crc16 computes CRC-16/CCITT-FALSE (poly 0x1021, init 0xFFFF).
func crc16(data string) string {
	crc := uint16(0xFFFF)
	for _, b := range []byte(data) {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return fmt.Sprintf("%04X", crc)
}
