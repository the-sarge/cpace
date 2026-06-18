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

func TestLEB128CanonicalDecode(t *testing.T) {
	tests := []struct {
		name     string
		encoded  []byte
		off      int
		want     int
		wantNext int
	}{
		{"empty", []byte{0x00, 0xff}, 0, 0, 1},
		{"single byte max", []byte{0x7f, 0xff}, 0, 0x7f, 1},
		{"two byte min", []byte{0x80, 0x01, 0xff}, 0, 0x80, 2},
		{"two byte max", []byte{0xff, 0x7f, 0xff}, 0, 0x3fff, 2},
		{"three byte min", []byte{0x80, 0x80, 0x01, 0xff}, 0, 0x4000, 3},
		{"associated data cap", []byte{0x80, 0x80, 0x04, 0xff}, 0, maxAssociatedDataLength, 3},
		{"offset", []byte{0xaa, 0x80, 0x01, 0xff}, 1, 0x80, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, next, err := readLEB128(tt.encoded, tt.off, maxLEB128BytesForField)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("readLEB128 value=%d want %d", got, tt.want)
			}
			if next != tt.wantNext {
				t.Fatalf("readLEB128 next=%d want %d", next, tt.wantNext)
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
