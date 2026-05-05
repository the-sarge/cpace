package cpace

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"testing"
)

//go:embed testdata/draft21-ristretto255-sha512.json
var draft21RistrettoVectorJSON []byte

//go:embed testdata/draft21-ristretto255-invalid.json
var draft21RistrettoInvalidJSON []byte

const (
	draft21RistrettoVectorJSONSHA256  = "dc74177668cc2374beaf57fcb6e4c08a908238bab6b74d8edf8c86e04bc663ae"
	draft21RistrettoInvalidJSONSHA256 = "6288f7ff96dfb8c2d6c4d743927c5fe6ac4aecbc56da2d1f00f27104000b6dfd"
)

type draftVector map[string][]byte

type draftInvalidVector struct {
	Valid     map[string][]byte
	InvalidY1 []byte
	InvalidY2 []byte
}

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

func loadDraftInvalidVectorJSON(in []byte) (draftInvalidVector, error) {
	var raw struct {
		Valid     map[string]string `json:"Valid"`
		InvalidY1 string            `json:"Invalid Y1"`
		InvalidY2 string            `json:"Invalid Y2"`
	}
	if err := json.Unmarshal(in, &raw); err != nil {
		return draftInvalidVector{}, err
	}
	valid := make(map[string][]byte, len(raw.Valid))
	for k, v := range raw.Valid {
		decoded, err := hex.DecodeString(v)
		if err != nil {
			return draftInvalidVector{}, err
		}
		valid[k] = decoded
	}
	invalidY1, err := hex.DecodeString(raw.InvalidY1)
	if err != nil {
		return draftInvalidVector{}, err
	}
	invalidY2, err := hex.DecodeString(raw.InvalidY2)
	if err != nil {
		return draftInvalidVector{}, err
	}
	return draftInvalidVector{Valid: valid, InvalidY1: invalidY1, InvalidY2: invalidY2}, nil
}

func pinnedJSONHash(in []byte) string {
	sum := sha256.Sum256(bytes.TrimRight(in, "\r\n"))
	return hex.EncodeToString(sum[:])
}

func TestEmbeddedDraftVectorJSON(t *testing.T) {
	// Hashes pin the embedded fixtures to the decoded JSON blocks in draft-21
	// Appendix B.3.9 and B.3.11.1.
	if got := pinnedJSONHash(draft21RistrettoVectorJSON); got != draft21RistrettoVectorJSONSHA256 {
		t.Fatalf("vector JSON SHA-256 got %s want %s", got, draft21RistrettoVectorJSONSHA256)
	}
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

func TestEmbeddedDraftInvalidVectorJSON(t *testing.T) {
	if got := pinnedJSONHash(draft21RistrettoInvalidJSON); got != draft21RistrettoInvalidJSONSHA256 {
		t.Fatalf("invalid vector JSON SHA-256 got %s want %s", got, draft21RistrettoInvalidJSONSHA256)
	}
	v, err := loadDraftInvalidVectorJSON(draft21RistrettoInvalidJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"s", "X", "G.scalar_mult_vfy(s,X)"} {
		if len(v.Valid[key]) == 0 {
			t.Fatalf("missing valid.%s", key)
		}
	}
	if len(v.InvalidY1) == 0 || len(v.InvalidY2) == 0 {
		t.Fatal("missing invalid vectors")
	}
	if !bytes.Equal(v.InvalidY2, identityEncoding) {
		t.Fatalf("Invalid Y2 got %x want identity encoding", v.InvalidY2)
	}
}

func TestRistrettoDraft21Vectors(t *testing.T) {
	v, err := loadDraftVectorJSON(draft21RistrettoVectorJSON)
	if err != nil {
		t.Fatal(err)
	}
	prs := v["PRS"]
	ci := v["CI"]
	sid := v["sid"]
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
	wantG := v["g"]
	if !bytes.Equal(g.Bytes(), wantG) {
		t.Fatalf("generator got %x want %x", g.Bytes(), wantG)
	}
	sidMutations := []struct {
		name string
		sid  []byte
	}{
		{"first byte", func() []byte {
			out := clone(sid)
			out[0] ^= 0x01
			return out
		}()},
		{"last byte", func() []byte {
			out := clone(sid)
			out[len(out)-1] ^= 0x01
			return out
		}()},
		{"nil", nil},
		{"empty", []byte{}},
		{"appended", append(clone(sid), 0x00)},
		{"truncated", clone(sid[:len(sid)-1])},
	}
	for _, tc := range sidMutations {
		t.Run("generator sid mutation "+tc.name, func(t *testing.T) {
			alteredG := calculateGenerator(prs, ci, tc.sid)
			if bytes.Equal(alteredG.Bytes(), wantG) {
				t.Fatal("generator unexpectedly matched official vector after sid mutation")
			}
		})
	}

	ya, err := scalarFromCanonical(v["ya"])
	if err != nil {
		t.Fatal(err)
	}
	yb, err := scalarFromCanonical(v["yb"])
	if err != nil {
		t.Fatal(err)
	}
	Ya := scalarMult(ya, g)
	Yb := scalarMult(yb, g)
	wantYa := v["Ya"]
	wantYb := v["Yb"]
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
	wantK := v["K"]
	if !bytes.Equal(k1, wantK) || !bytes.Equal(k2, wantK) {
		t.Fatalf("K got %x/%x want %x", k1, k2, wantK)
	}

	ada := v["ADa"]
	adb := v["ADb"]
	trIR := transcriptIR(Ya, ada, Yb, adb)
	wantTrIR := hx(t, "20d6bac480f2c386c394efc7c47adb9925dcd2630b64f240c50f8d0eec482b915703414461203ea7e0b19560d7c0b0f5734f63b955286dfa8232b5ebe63324e2d9e7433f725803414462")
	if !bytes.Equal(trIR, wantTrIR) {
		t.Fatalf("transcript_ir got %x want %x", trIR, wantTrIR)
	}
	iskIR := deriveISK(sid, wantK, trIR)
	wantISKIR := v["ISK_IR"]
	if !bytes.Equal(iskIR, wantISKIR) {
		t.Fatalf("ISK_IR got %x want %x", iskIR, wantISKIR)
	}

	trOC := transcriptOC(Ya, ada, Yb, adb)
	iskOC := deriveISK(sid, wantK, trOC)
	wantISKOC := v["ISK_SY"]
	if !bytes.Equal(iskOC, wantISKOC) {
		t.Fatalf("ISK_SY got %x want %x", iskOC, wantISKOC)
	}

	sidOut := sha512.Sum512(append([]byte("CPaceSidOutput"), trIR...))
	wantSidOut := v["sid_output_ir"]
	if !bytes.Equal(sidOut[:], wantSidOut) {
		t.Fatalf("sid_output got %x want %x", sidOut, wantSidOut)
	}
}

func TestScalarMultVFYDraftInvalidVectors(t *testing.T) {
	v, err := loadDraftInvalidVectorJSON(draft21RistrettoInvalidJSON)
	if err != nil {
		t.Fatal(err)
	}
	s, err := scalarFromCanonical(v.Valid["s"])
	if err != nil {
		t.Fatal(err)
	}
	valid := v.Valid["X"]
	got, ok := scalarMultVFY(s, valid)
	want := v.Valid["G.scalar_mult_vfy(s,X)"]
	if !ok || !bytes.Equal(got, want) {
		t.Fatalf("valid scalar_mult_vfy got ok=%v %x want %x", ok, got, want)
	}
	for _, invalid := range [][]byte{
		v.InvalidY1,
		v.InvalidY2,
	} {
		got, ok := scalarMultVFY(s, invalid)
		if ok || !bytes.Equal(got, identityEncoding) {
			t.Fatalf("invalid scalar_mult_vfy got ok=%v %x", ok, got)
		}
	}
}
