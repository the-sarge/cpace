package cpace

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestStringUtilitiesDraftVectors(t *testing.T) {
	tests := []struct {
		name string
		got  []byte
		want string
	}{
		{"prepend empty", prependLen(nil), "00"},
		{"prepend 1234", prependLen([]byte("1234")), "0431323334"},
		{"lv_cat", lvCat([]byte("1234"), []byte("5"), nil, []byte("678")), "043132333401350003363738"},
		{"o_cat first", oCat([]byte("ABCD"), []byte("BCD")), "6f6342434441424344"},
		{"o_cat second", oCat([]byte("BCD"), []byte("ABCDE")), "6f634243444142434445"},
		{"transcript_ir", transcriptIR([]byte("123"), []byte("PartyA"), []byte("234"), []byte("PartyB")), "03313233065061727479410332333406506172747942"},
		{"transcript_oc", transcriptOC([]byte("123"), []byte("PartyA"), []byte("234"), []byte("PartyB")), "6f6303323334065061727479420331323306506172747941"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := hex.DecodeString(tt.want)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(tt.got, want) {
				t.Fatalf("got %x want %x", tt.got, want)
			}
		})
	}
}

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
			got := prependLen(bytes.Repeat([]byte{0xaa}, tt.length))
			wantPrefix, err := hex.DecodeString(tt.wantPrefix)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(got, wantPrefix) {
				t.Fatalf("prefix=%x want %x", got[:len(wantPrefix)], wantPrefix)
			}
			if len(got) != len(wantPrefix)+tt.length {
				t.Fatalf("encoded len=%d want %d", len(got), len(wantPrefix)+tt.length)
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

func TestWireFormatPrefixByte(t *testing.T) {
	if wireFormatV1 != 0xc1 {
		t.Fatalf("wireFormatV1=%#x, want 0xc1", wireFormatV1)
	}
	if wireSuite != 0x01 {
		t.Fatalf("wireSuite=%#x, want 0x01", wireSuite)
	}
	cases := []struct {
		name string
		msg  []byte
	}{
		{"A", encodeMessageA(nil, make([]byte, pointSize), nil)},
		{"B", encodeMessageB(make([]byte, pointSize), nil, make([]byte, tagSize))},
		{"C", encodeMessageC(make([]byte, tagSize))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.msg[0] != wireFormatV1 {
				t.Fatalf("format prefix=%#x, want %#x", tc.msg[0], wireFormatV1)
			}
			if tc.msg[1] != wireSuite {
				t.Fatalf("suite byte=%#x, want %#x", tc.msg[1], wireSuite)
			}
		})
	}
}
