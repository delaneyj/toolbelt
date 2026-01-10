package toolbelt

import (
	"context"
	"encoding/base32"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestEncodedIDToInt64(t *testing.T) {
	value := int64(1234567890)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	got, err := EncodedIDToInt64(encoded)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != value {
		t.Fatalf("expected %d, got %d", value, got)
	}
}

func TestEncodedIDToInt64_AcceptsPaddingAndLowercase(t *testing.T) {
	value := int64(987654321)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))

	padded := base32.StdEncoding.EncodeToString(buf)
	got, err := EncodedIDToInt64(padded)
	if err != nil {
		t.Fatalf("expected no error for padded input, got %v", err)
	}
	if got != value {
		t.Fatalf("expected %d, got %d", value, got)
	}

	lower := strings.ToLower(padded)
	got, err = EncodedIDToInt64(lower)
	if err != nil {
		t.Fatalf("expected no error for lowercase input, got %v", err)
	}
	if got != value {
		t.Fatalf("expected %d, got %d", value, got)
	}
}

func TestEncodedIDToInt64_InvalidLength(t *testing.T) {
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte{1, 2, 3})
	_, err := EncodedIDToInt64(encoded)
	if err == nil {
		t.Fatal("expected error for non-8-byte decoded value")
	}
}

func TestChiParamInt64(t *testing.T) {
	r := requestWithParam("id", "42")
	got, err := ChiParamInt64(r, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestChiParamInt64_Invalid(t *testing.T) {
	r := requestWithParam("id", "nope")
	_, err := ChiParamInt64(r, "id")
	if err == nil {
		t.Fatal("expected error for invalid int64")
	}
}

func TestChiParamEncodedID(t *testing.T) {
	value := int64(2468)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)

	r := requestWithParam("id", encoded)
	got, err := ChiParamEncodedID(r, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != value {
		t.Fatalf("expected %d, got %d", value, got)
	}
}

func TestChiParamEncodedID_Invalid(t *testing.T) {
	r := requestWithParam("id", "invalid")
	_, err := ChiParamEncodedID(r, "id")
	if err == nil {
		t.Fatal("expected error for invalid encoded id")
	}
}

func requestWithParam(name, value string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(name, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}
