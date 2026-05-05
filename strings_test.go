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

func TestLEB128ReaderRejectsMalformed(t *testing.T) {
	cases := [][]byte{
		{wireVersion, wireSuite, roleC},
		{wireVersion, wireSuite, roleC, 0x80},
		{wireVersion, wireSuite, roleC, 0x80, 0x00},
		{wireVersion, wireSuite, roleC, 0x80, 0x80, 0x80, 0x80, 0x00},
	}
	for _, tc := range cases {
		if _, err := decodeMessageC(tc); err == nil {
			t.Fatalf("decodeMessageC(%x) succeeded", tc)
		}
	}
}
