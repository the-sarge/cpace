package cpace

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
)

//go:embed testdata/draft21-ristretto255-generator.json
var draft21RistrettoGeneratorJSON []byte

//go:embed testdata/draft21-ristretto255-confirmation-tags.json
var draft21RistrettoConfirmationTagJSON []byte

const (
	draft21RistrettoGeneratorJSONSHA256       = "05c8a34bd623fbdefd7fbffcd261d2420bd34363efa301d0b0dd9817f7f47c94"
	draft21RistrettoVectorJSONSHA256          = "dc74177668cc2374beaf57fcb6e4c08a908238bab6b74d8edf8c86e04bc663ae"
	draft21RistrettoInvalidJSONSHA256         = "6288f7ff96dfb8c2d6c4d743927c5fe6ac4aecbc56da2d1f00f27104000b6dfd"
	draft21RistrettoConfirmationTagJSONSHA256 = "1d0b59b3b7486dee3569ad4e8d6908e2f575dfcd9f26804c34584cb29515e0d4"
)

type draftGeneratorVector struct {
	H               string
	HsInBytes       int
	ZPADLength      int
	PRS             []byte
	DSI             []byte
	CI              []byte
	SID             []byte
	GeneratorString []byte
	HashResult      []byte
	EncodedG        []byte
}

func hx(t *testing.T, s string) []byte {
	t.Helper()
	out, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func loadDraftGeneratorJSON(in []byte) (draftGeneratorVector, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(in, &raw); err != nil {
		return draftGeneratorVector{}, err
	}
	stringField := func(key string) (string, error) {
		var out string
		if err := json.Unmarshal(raw[key], &out); err != nil {
			return "", err
		}
		return out, nil
	}
	intField := func(key string) (int, error) {
		var out int
		if err := json.Unmarshal(raw[key], &out); err != nil {
			return 0, err
		}
		return out, nil
	}
	hexField := func(key string) ([]byte, error) {
		s, err := stringField(key)
		if err != nil {
			return nil, err
		}
		return hex.DecodeString(s)
	}
	h, err := stringField("H")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	hsInBytes, err := intField("H.s_in_bytes")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	zpadLength, err := intField("ZPAD length")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	prs, err := hexField("PRS")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	dsi, err := hexField("DSI")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	ci, err := hexField("CI")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	sid, err := hexField("sid")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	genStr, err := hexField("generator_string(G.DSI,PRS,CI,sid,H.s_in_bytes)")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	hashResult, err := hexField("hash result")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	encodedG, err := hexField("encoded generator g")
	if err != nil {
		return draftGeneratorVector{}, err
	}
	return draftGeneratorVector{
		H:               h,
		HsInBytes:       hsInBytes,
		ZPADLength:      zpadLength,
		PRS:             prs,
		DSI:             dsi,
		CI:              ci,
		SID:             sid,
		GeneratorString: genStr,
		HashResult:      hashResult,
		EncodedG:        encodedG,
	}, nil
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
	for _, key := range []string{"PRS", "CI", "sid", "g", "Ya", "Yb", "K", "ISK_IR", "ISK_SY", "sid_output_ir", "sid_output_oc"} {
		if len(v[key]) == 0 {
			t.Fatalf("missing %s", key)
		}
	}
}

func TestEmbeddedDraftGeneratorJSON(t *testing.T) {
	// Hash pins the decoded JSON block in draft-21 Appendix B.3.1.1.
	if got := pinnedJSONHash(draft21RistrettoGeneratorJSON); got != draft21RistrettoGeneratorJSONSHA256 {
		t.Fatalf("generator JSON SHA-256 got %s want %s", got, draft21RistrettoGeneratorJSONSHA256)
	}
	v, err := loadDraftGeneratorJSON(draft21RistrettoGeneratorJSON)
	if err != nil {
		t.Fatal(err)
	}
	if v.H != "SHA-512" || v.HsInBytes != sha512BlockSize || v.ZPADLength != 100 {
		t.Fatalf("unexpected generator metadata H=%q H.s_in_bytes=%d ZPAD=%d", v.H, v.HsInBytes, v.ZPADLength)
	}
	gotGS := generatorString(v.DSI, v.PRS, v.CI, v.SID, v.HsInBytes)
	if !bytes.Equal(gotGS, v.GeneratorString) {
		t.Fatalf("generator_string got %x want %x", gotGS, v.GeneratorString)
	}
	sum := sha512.Sum512(v.GeneratorString)
	if !bytes.Equal(sum[:], v.HashResult) {
		t.Fatalf("hash result got %x want %x", sum, v.HashResult)
	}
	g := calculateGenerator(v.PRS, v.CI, v.SID)
	if !bytes.Equal(g.Bytes(), v.EncodedG) {
		t.Fatalf("encoded generator got %x want %x", g.Bytes(), v.EncodedG)
	}
}

func TestEmbeddedDraftConfirmationTagGoldens(t *testing.T) {
	// Hash pins package-local confirmation-tag goldens captured at the
	// primitive seam from the draft-21 Appendix B.3.9 vector.
	if got := pinnedJSONHash(draft21RistrettoConfirmationTagJSON); got != draft21RistrettoConfirmationTagJSONSHA256 {
		t.Fatalf("confirmation tag JSON SHA-256 got %s want %s", got, draft21RistrettoConfirmationTagJSONSHA256)
	}
	tags, err := loadDraftVectorJSON(draft21RistrettoConfirmationTagJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"tagA", "tagB"} {
		if len(tags[key]) != tagSize {
			t.Fatalf("%s length=%d want %d", key, len(tags[key]), tagSize)
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

	k1, err := scalarMultVFY(ya, Yb)
	if err != nil {
		t.Fatalf("scalarMultVFY(ya,Yb): %v", err)
	}
	k2, err := scalarMultVFY(yb, Ya)
	if err != nil {
		t.Fatalf("scalarMultVFY(yb,Ya): %v", err)
	}
	wantK := v["K"]
	if !bytes.Equal(k1, wantK) || !bytes.Equal(k2, wantK) {
		t.Fatalf("K got %x/%x want %x", k1, k2, wantK)
	}

	ada := v["ADa"]
	adb := v["ADb"]
	transcriptIR := newIRTranscript(Ya, ada, Yb, adb)
	trIR := transcriptIR.bytes()
	wantTrIR := hx(t, "20d6bac480f2c386c394efc7c47adb9925dcd2630b64f240c50f8d0eec482b915703414461203ea7e0b19560d7c0b0f5734f63b955286dfa8232b5ebe63324e2d9e7433f725803414462")
	if !bytes.Equal(trIR, wantTrIR) {
		t.Fatalf("transcript_ir got %x want %x", trIR, wantTrIR)
	}
	iskIR := deriveISK(sid, wantK, trIR)
	wantISKIR := v["ISK_IR"]
	if !bytes.Equal(iskIR, wantISKIR) {
		t.Fatalf("ISK_IR got %x want %x", iskIR, wantISKIR)
	}
	tags, err := loadDraftVectorJSON(draft21RistrettoConfirmationTagJSON)
	if err != nil {
		t.Fatal(err)
	}
	if got := confirmationTag(iskIR, sid, Yb, adb); !bytes.Equal(got, tags["tagB"]) {
		t.Fatalf("tagB got %x want %x", got, tags["tagB"])
	}
	if got := confirmationTag(iskIR, sid, Ya, ada); !bytes.Equal(got, tags["tagA"]) {
		t.Fatalf("tagA got %x want %x", got, tags["tagA"])
	}

	trOC := testTranscriptOC(Ya, ada, Yb, adb)
	iskOC := deriveISK(sid, wantK, trOC)
	wantISKOC := v["ISK_SY"]
	if !bytes.Equal(iskOC, wantISKOC) {
		t.Fatalf("ISK_SY got %x want %x", iskOC, wantISKOC)
	}

	wantSidOut := v["sid_output_ir"]
	if got := transcriptIR.transcriptID(); !bytes.Equal(got, wantSidOut) {
		t.Fatalf("sid_output got %x want %x", got, wantSidOut)
	}
	wantSidOutOC := v["sid_output_oc"]
	if got := transcriptID(trOC); !bytes.Equal(got, wantSidOutOC) {
		t.Fatalf("sid_output_oc got %x want %x", got, wantSidOutOC)
	}
}

func TestCoreDraft21Vectors(t *testing.T) {
	v, err := loadDraftVectorJSON(draft21RistrettoVectorJSON)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := loadDraftVectorJSON(draft21RistrettoConfirmationTagJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"ya", "yb"} {
		if v[key][31]&0xf0 != 0 {
			t.Fatalf("%s top nibble is not sampler-injectable: %x", key, v[key][31])
		}
	}

	initNC := draftVectorInput(v, v["ADa"])
	defer initNC.wipe()
	initCore, gotYa, err := newInitiatorCore(initNC, &repeatingReader{buf: v["ya"]})
	if err != nil {
		t.Fatal(err)
	}
	defer clearScalar(initCore.scalar)
	if !bytes.Equal(gotYa, v["Ya"]) {
		t.Fatalf("newInitiatorCore Ya got %x want %x", gotYa, v["Ya"])
	}
	gotTagA, initSession, err := initCore.finish(v["Yb"], v["ADb"], tags["tagB"])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotTagA, tags["tagA"]) {
		t.Fatalf("initiator core tagA got %x want %x", gotTagA, tags["tagA"])
	}
	if !bytes.Equal(initSession.state.isk, v["ISK_IR"]) {
		t.Fatalf("initiator core session ISK got %x want %x", initSession.state.isk, v["ISK_IR"])
	}
	wantTranscriptID := v["sid_output_ir"]
	initTranscriptID := initSession.TranscriptID()
	if !bytes.Equal(initTranscriptID, wantTranscriptID) {
		t.Fatalf("initiator core session TranscriptID got %x want %x", initTranscriptID, wantTranscriptID)
	}

	respNC := draftVectorInput(v, v["ADb"])
	defer respNC.wipe()
	respCore, gotYb, gotTagB, err := newResponderCore(respNC, v["Ya"], v["ADa"], &repeatingReader{buf: v["yb"]})
	if err != nil {
		t.Fatal(err)
	}
	defer clearBytes(respCore.isk)
	defer clearBytes(respCore.transcript)
	if !bytes.Equal(gotYb, v["Yb"]) {
		t.Fatalf("newResponderCore Yb got %x want %x", gotYb, v["Yb"])
	}
	if !bytes.Equal(respCore.isk, v["ISK_IR"]) {
		t.Fatalf("responder core ISK got %x want %x", respCore.isk, v["ISK_IR"])
	}
	if !bytes.Equal(gotTagB, tags["tagB"]) {
		t.Fatalf("responder core tagB got %x want %x", gotTagB, tags["tagB"])
	}
	respSession, err := respCore.finish(tags["tagA"])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(respSession.state.isk, v["ISK_IR"]) {
		t.Fatalf("responder core session ISK got %x want %x", respSession.state.isk, v["ISK_IR"])
	}
	respTranscriptID := respSession.TranscriptID()
	if !bytes.Equal(respTranscriptID, wantTranscriptID) {
		t.Fatalf("responder core session TranscriptID got %x want %x", respTranscriptID, wantTranscriptID)
	}
	if !bytes.Equal(initTranscriptID, respTranscriptID) {
		t.Fatalf("core session TranscriptIDs differ: initiator %x responder %x", initTranscriptID, respTranscriptID)
	}
}

func draftVectorInput(v draftVector, ad []byte) normalizedInput {
	return normalizedInput{
		password:    clone(v["PRS"]),
		initiatorID: []byte("A_initiator"),
		responderID: []byte("B_responder"),
		ci:          clone(v["CI"]),
		sid:         clone(v["sid"]),
		ad:          clone(ad),
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
	got, err := scalarMultVFY(s, valid)
	want := v.Valid["G.scalar_mult_vfy(s,X)"]
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("valid scalar_mult_vfy got err=%v %x want %x", err, got, want)
	}
	for _, tc := range []struct {
		name     string
		encoded  []byte
		sentinel error
	}{
		{"Y1 non-canonical", v.InvalidY1, ErrPeerShareEncoding},
		{"Y2 identity", v.InvalidY2, ErrPeerShareIdentity},
	} {
		got, err := scalarMultVFY(s, tc.encoded)
		if got != nil {
			t.Fatalf("%s: invalid scalar_mult_vfy out=%x want nil", tc.name, got)
		}
		if !errors.Is(err, ErrAbort) || !errors.Is(err, tc.sentinel) {
			t.Fatalf("%s: invalid scalar_mult_vfy err=%v want ErrAbort and %v", tc.name, err, tc.sentinel)
		}
	}
}
