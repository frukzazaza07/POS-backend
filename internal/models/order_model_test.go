package models_test

import (
	"strings"
	"testing"
	"time"

	"pos-backend/internal/models"
)

func TestGeneratePosOrderID_Format(t *testing.T) {
	id := models.GeneratePosOrderID()

	if !strings.HasPrefix(id, "ORDER-") {
		t.Errorf("want ORDER- prefix, got %s", id)
	}

	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("want 3 dash-separated parts, got %d in %q", len(parts), id)
	}

	wantDate := time.Now().Format("20060102")
	if parts[1] != wantDate {
		t.Errorf("want date %s, got %s", wantDate, parts[1])
	}

	if len(parts[2]) != 8 {
		t.Errorf("want 8-char suffix, got %q (len=%d)", parts[2], len(parts[2]))
	}

	if parts[2] != strings.ToUpper(parts[2]) {
		t.Errorf("suffix should be uppercase, got %s", parts[2])
	}
}

func TestGeneratePosOrderID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := models.GeneratePosOrderID()
		if seen[id] {
			t.Fatalf("duplicate ID on iteration %d: %s", i, id)
		}
		seen[id] = true
	}
}
