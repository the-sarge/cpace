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
