package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delaneyj/toolbelt/id"
	"github.com/go-chi/chi/v5"
)

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
	encoded := id.EncodeID(value)

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
