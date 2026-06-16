package cpace

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestLengthValueEncodingBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		length     int
		wantPrefix string
	}{
		{"empty", 0, "00"},
		{"single byte max", 0x7f, "7f"},
		{"two byte min", 0x80, "8001"},
		{"two byte max", 0x3fff, "ff7f"},
		{"three byte min", 0x4000, "808001"},
		{"associated data cap", maxAssociatedDataLength, "808004"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := bytes.Repeat([]byte{0xaa}, tt.length)
			got := prependLen(payload)
			wantPrefix, err := hex.DecodeString(tt.wantPrefix)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(got, wantPrefix) {
				t.Fatalf("prefix=%x want %x", got[:len(wantPrefix)], wantPrefix)
			}
			if !bytes.Equal(got[len(wantPrefix):], payload) {
				t.Fatalf("payload=%x want %x", got[len(wantPrefix):], payload)
			}
			if len(got) != len(wantPrefix)+tt.length {
				t.Fatalf("encoded len=%d want %d", len(got), len(wantPrefix)+tt.length)
			}
		})
	}
}

func TestLEB128LengthInvariant(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		encoded []byte
	}{
		{"empty", 0, appendLEB128(nil, 0)},
		{"single byte max", 0x7f, appendLEB128(nil, 0x7f)},
		{"two byte min", 0x80, appendLEB128(nil, 0x80)},
		{"two byte max", 0x3fff, appendLEB128(nil, 0x3fff)},
		{"three byte min", 0x4000, appendLEB128(nil, 0x4000)},
		{"associated data cap", maxAssociatedDataLength, appendLEB128(nil, maxAssociatedDataLength)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := leb128LenInt(tt.length); got != len(tt.encoded) {
				t.Fatalf("leb128LenInt(%d)=%d want %d", tt.length, got, len(tt.encoded))
			}
			if got := lengthValueLen(tt.length); got != len(tt.encoded)+tt.length {
				t.Fatalf("lengthValueLen(%d)=%d want %d", tt.length, got, len(tt.encoded)+tt.length)
			}
		})
	}
}

func TestLVCatBoundaryComposition(t *testing.T) {
	first := bytes.Repeat([]byte{0xaa}, 0x80)
	second := []byte("z")
	got := lvCat(first, second)
	wantPrefix, err := hex.DecodeString("8001")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(got, wantPrefix) {
		t.Fatalf("first prefix=%x want %x", got[:len(wantPrefix)], wantPrefix)
	}
	wantSecond := append([]byte{0x01}, second...)
	if !bytes.Equal(got[len(wantPrefix)+len(first):], wantSecond) {
		t.Fatalf("second field=%x want %x", got[len(wantPrefix)+len(first):], wantSecond)
	}
}
