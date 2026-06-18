package cpace

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func testOCat(a, b []byte) []byte {
	out := []byte("oc")
	if bytes.Compare(a, b) > 0 {
		out = append(out, a...)
		out = append(out, b...)
		return out
	}
	out = append(out, b...)
	out = append(out, a...)
	return out
}

func testTranscriptOC(ya, ada, yb, adb []byte) []byte {
	return testOCat(lvCat(ya, ada), lvCat(yb, adb))
}

func TestStringUtilitiesDraftVectors(t *testing.T) {
	tests := []struct {
		name string
		got  []byte
		want string
	}{
		{"prepend empty", prependLen(nil), "00"},
		{"prepend 1234", prependLen([]byte("1234")), "0431323334"},
		{"lv_cat", lvCat([]byte("1234"), []byte("5"), nil, []byte("678")), "043132333401350003363738"},
		{"o_cat first", testOCat([]byte("ABCD"), []byte("BCD")), "6f6342434441424344"},
		{"o_cat second", testOCat([]byte("BCD"), []byte("ABCDE")), "6f634243444142434445"},
		{"transcript_ir", newIRTranscript([]byte("123"), []byte("PartyA"), []byte("234"), []byte("PartyB")).bytes(), "03313233065061727479410332333406506172747942"},
		{"transcript_oc", testTranscriptOC([]byte("123"), []byte("PartyA"), []byte("234"), []byte("PartyB")), "6f6303323334065061727479420331323306506172747941"},
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

func TestIRTranscriptDraftVectorFlow(t *testing.T) {
	v, err := loadDraftVectorJSON(draft21RistrettoVectorJSON)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := loadDraftVectorJSON(draft21RistrettoConfirmationTagJSON)
	if err != nil {
		t.Fatal(err)
	}
	tr := newIRTranscript(v["Ya"], v["ADa"], v["Yb"], v["ADb"])
	wantTranscript := hx(t, "20d6bac480f2c386c394efc7c47adb9925dcd2630b64f240c50f8d0eec482b915703414461203ea7e0b19560d7c0b0f5734f63b955286dfa8232b5ebe63324e2d9e7433f725803414462")
	if !bytes.Equal(tr.bytes(), wantTranscript) {
		t.Fatalf("IR transcript=%x want %x", tr.bytes(), wantTranscript)
	}
	isk := tr.deriveISK(v["sid"], v["K"])
	if !bytes.Equal(isk, v["ISK_IR"]) {
		t.Fatalf("ISK_IR=%x want %x", isk, v["ISK_IR"])
	}
	if got := tr.responderConfirmationTag(isk, v["sid"]); !bytes.Equal(got, tags["tagB"]) {
		t.Fatalf("responder confirmation tag=%x want %x", got, tags["tagB"])
	}
	if got := tr.initiatorConfirmationTag(isk, v["sid"]); !bytes.Equal(got, tags["tagA"]) {
		t.Fatalf("initiator confirmation tag=%x want %x", got, tags["tagA"])
	}
}

func TestIRTranscriptOwnsInputsAndOutput(t *testing.T) {
	ya := []byte("ya")
	ada := []byte("ada")
	yb := []byte("yb")
	adb := []byte("adb")
	tr := newIRTranscript(ya, ada, yb, adb)
	sid := []byte("sid")
	isk := []byte("isk")
	wantTranscript := tr.bytes()
	wantTagA := tr.initiatorConfirmationTag(isk, sid)
	wantTagB := tr.responderConfirmationTag(isk, sid)

	ya[0] = 'Y'
	ada[0] = 'A'
	yb[0] = 'Z'
	adb[0] = 'B'
	gotTranscript := tr.bytes()
	gotTranscript[0] ^= 0xff

	if !bytes.Equal(tr.bytes(), wantTranscript) {
		t.Fatalf("transcript changed after caller mutation")
	}
	if got := tr.initiatorConfirmationTag(isk, sid); !bytes.Equal(got, wantTagA) {
		t.Fatalf("initiator tag changed after caller mutation")
	}
	if got := tr.responderConfirmationTag(isk, sid); !bytes.Equal(got, wantTagB) {
		t.Fatalf("responder tag changed after caller mutation")
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
