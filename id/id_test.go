package id

import (
	"encoding/base32"
	"encoding/binary"
	"strings"
	"testing"
)

func TestEncodedIDToInt64(t *testing.T) {
	value := int64(1234567890)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))

	encoded := EncodeID(value)
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
