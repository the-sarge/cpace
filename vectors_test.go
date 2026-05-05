package cpace

import (
	"bytes"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"testing"
)

//go:embed testdata/draft21-ristretto255-sha512.json
var draft21RistrettoVectorJSON []byte

type draftVector map[string][]byte

func hx(t *testing.T, s string) []byte {
	t.Helper()
	out, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func loadDraftVectorJSON(in []byte) (draftVector, error) {
	var raw map[string]string
	if err := json.Unmarshal(in, &raw); err != nil {
		return nil, err
	}
	out := make(draftVector, len(raw))
	for k, v := range raw {
		decoded, err := hex.DecodeString(v)
		if err != nil {
			return nil, err
		}
		out[k] = decoded
	}
	return out, nil
}

func TestEmbeddedDraftVectorJSON(t *testing.T) {
	v, err := loadDraftVectorJSON(draft21RistrettoVectorJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"PRS", "CI", "sid", "g", "Ya", "Yb", "K", "ISK_IR", "ISK_SY", "sid_output_ir"} {
		if len(v[key]) == 0 {
			t.Fatalf("missing %s", key)
		}
	}
}

func TestRistrettoDraft21Vectors(t *testing.T) {
	prs := []byte("Password")
	ci := hx(t, "0b415f696e69746961746f720b425f726573706f6e646572")
	sid := hx(t, "7e4b4791d6a8ef019b936c79fb7f2c57")
	gs := generatorString([]byte(dsiRistretto255), prs, ci, sid, sha512BlockSize)
	wantGS := hx(t, "11435061636552697374726574746f3235350850617373776f726464")
	wantGS = append(wantGS, make([]byte, 100)...)
	wantGS = append(wantGS, hx(t, "180b415f696e69746961746f720b425f726573706f6e646572107e4b4791d6a8ef019b936c79fb7f2c57")...)
	if !bytes.Equal(gs, wantGS) {
		t.Fatalf("generator string got %x want %x", gs, wantGS)
	}

	sum := sha512.Sum512(gs)
	wantHash := hx(t, "da6d3ddc8802fca9058755ffd3ebde08a9c2c74945901a258482a288b6663af06bf645c93cd1c51512307199c80e84908916d983b34af77205f90851a657ee27")
	if !bytes.Equal(sum[:], wantHash) {
		t.Fatalf("generator hash got %x want %x", sum, wantHash)
	}
	g := calculateGenerator(prs, ci, sid)
	wantG := hx(t, "222b6b195fe84b1652badb6f6a3ae3d24341e7306967f0b8115b40d5698c7e56")
	if !bytes.Equal(g.Bytes(), wantG) {
		t.Fatalf("generator got %x want %x", g.Bytes(), wantG)
	}

	ya, err := scalarFromCanonical(hx(t, "da3d23700a9e5699258aef94dc060dfda5ebb61f02a5ea77fad53f4ff0976d08"))
	if err != nil {
		t.Fatal(err)
	}
	yb, err := scalarFromCanonical(hx(t, "d2316b454718c35362d83d69df6320f38578ed5984651435e2949762d900b80d"))
	if err != nil {
		t.Fatal(err)
	}
	Ya := scalarMult(ya, g)
	Yb := scalarMult(yb, g)
	wantYa := hx(t, "d6bac480f2c386c394efc7c47adb9925dcd2630b64f240c50f8d0eec482b9157")
	wantYb := hx(t, "3ea7e0b19560d7c0b0f5734f63b955286dfa8232b5ebe63324e2d9e7433f7258")
	if !bytes.Equal(Ya, wantYa) {
		t.Fatalf("Ya got %x want %x", Ya, wantYa)
	}
	if !bytes.Equal(Yb, wantYb) {
		t.Fatalf("Yb got %x want %x", Yb, wantYb)
	}

	k1, ok := scalarMultVFY(ya, Yb)
	if !ok {
		t.Fatal("scalarMultVFY(ya,Yb) failed")
	}
	k2, ok := scalarMultVFY(yb, Ya)
	if !ok {
		t.Fatal("scalarMultVFY(yb,Ya) failed")
	}
	wantK := hx(t, "80b69a8a76457ab6a4d7f887a4bf6b55a2f80ac19c333f917a05fc9887c8b40f")
	if !bytes.Equal(k1, wantK) || !bytes.Equal(k2, wantK) {
		t.Fatalf("K got %x/%x want %x", k1, k2, wantK)
	}

	ada := []byte("ADa")
	adb := []byte("ADb")
	trIR := transcriptIR(Ya, ada, Yb, adb)
	wantTrIR := hx(t, "20d6bac480f2c386c394efc7c47adb9925dcd2630b64f240c50f8d0eec482b915703414461203ea7e0b19560d7c0b0f5734f63b955286dfa8232b5ebe63324e2d9e7433f725803414462")
	if !bytes.Equal(trIR, wantTrIR) {
		t.Fatalf("transcript_ir got %x want %x", trIR, wantTrIR)
	}
	iskIR := deriveISK(sid, wantK, trIR)
	wantISKIR := hx(t, "b69effbf61b51d56401c0f65601abe428de8206feaaf0e32198896dcae7b35cd2b38950a39dfd5d4a79164614c2984f7daa460b588c1e80c3fa2068af7900447")
	if !bytes.Equal(iskIR, wantISKIR) {
		t.Fatalf("ISK_IR got %x want %x", iskIR, wantISKIR)
	}

	trOC := transcriptOC(Ya, ada, Yb, adb)
	iskOC := deriveISK(sid, wantK, trOC)
	wantISKOC := hx(t, "544199d71f62f8d9a1fee55727e24fe4a45844593c2b6013c4fa3969d0e5debb2244675c0b43397cbb68d342b01fc0f98fc961469a25134de9f0f813c1a57476")
	if !bytes.Equal(iskOC, wantISKOC) {
		t.Fatalf("ISK_SY got %x want %x", iskOC, wantISKOC)
	}

	sidOut := sha512.Sum512(append([]byte("CPaceSidOutput"), trIR...))
	wantSidOut := hx(t, "bb1c449b35f0ea79a65c209f329a693d475e0ce2387bed9fe4b78f60b2a27c219813fb2cfe175ef40d2222d9261e66da7d78f7c55a303b1b8611dcdfab880c47")
	if !bytes.Equal(sidOut[:], wantSidOut) {
		t.Fatalf("sid_output got %x want %x", sidOut, wantSidOut)
	}
}

func TestScalarMultVFYDraftInvalidVectors(t *testing.T) {
	s, err := scalarFromCanonical(hx(t, "7cd0e075fa7955ba52c02759a6c90dbbfc10e6d40aea8d283e407d88cf538a05"))
	if err != nil {
		t.Fatal(err)
	}
	valid := hx(t, "2c3c6b8c4f3800e7aef6864025b4ed79bd599117e427c41bd47d93d654b4a51c")
	got, ok := scalarMultVFY(s, valid)
	want := hx(t, "7c13645fe790a468f62c39beb7388e541d8405d1ade69d1778c5fe3e7f6b600e")
	if !ok || !bytes.Equal(got, want) {
		t.Fatalf("valid scalar_mult_vfy got ok=%v %x want %x", ok, got, want)
	}
	for _, invalid := range [][]byte{
		hx(t, "2b3c6b8c4f3800e7aef6864025b4ed79bd599117e427c41bd47d93d654b4a51c"),
		hx(t, "0000000000000000000000000000000000000000000000000000000000000000"),
	} {
		got, ok := scalarMultVFY(s, invalid)
		if ok || !bytes.Equal(got, identityEncoding) {
			t.Fatalf("invalid scalar_mult_vfy got ok=%v %x", ok, got)
		}
	}
}
