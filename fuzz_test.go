package cpace

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
)

//go:embed testdata/draft21-ristretto255-sha512.json
var draft21RistrettoVectorJSON []byte

//go:embed testdata/draft21-ristretto255-invalid.json
var draft21RistrettoInvalidJSON []byte

const fuzzProtocolInputCap = 4096

type draftVector map[string][]byte

type draftInvalidVector struct {
	Valid     map[string][]byte
	InvalidY1 []byte
	InvalidY2 []byte
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

func FuzzDecodeMessageA(f *testing.F) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(f, initInput, respInput)
	invalid := fuzzDraftInvalidVector(f)
	for _, seed := range messageAFuzzSeeds(exchange.msgA, exchange.msgB, invalid.InvalidY1) {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = decodeMessageA(in)
	})
}

func FuzzDecodeMessageB(f *testing.F) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(f, initInput, respInput)
	msgC, _ := exchange.finishInitiator()
	invalid := fuzzDraftInvalidVector(f)
	for _, seed := range messageBFuzzSeeds(exchange.msgB, msgC, invalid.InvalidY1) {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = decodeMessageB(in)
	})
}

func FuzzDecodeMessageC(f *testing.F) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(f, initInput, respInput)
	msgC, _ := exchange.finishInitiator()
	for _, seed := range messageCFuzzSeeds(msgC, exchange.msgA) {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = decodeMessageC(in)
	})
}

func FuzzDraftVectorJSONLoader(f *testing.F) {
	f.Add(draft21RistrettoVectorJSON)
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = loadDraftVectorJSON(in)
	})
}

func FuzzDraftInvalidVectorJSONLoader(f *testing.F) {
	f.Add(draft21RistrettoInvalidJSON)
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = loadDraftInvalidVectorJSON(in)
	})
}

func FuzzProtocolConsistency(f *testing.F) {
	f.Add([]byte("sid"), []byte("ctx"), []byte("ADa"), []byte("ADb"))
	f.Add([]byte("sid2"), []byte{}, []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, sid, ctx, ada, adb []byte) {
		if len(sid) == 0 || len(sid) > sessionIDCap.length || len(ctx) > contextCap.length || len(ada) > 1024 || len(adb) > 1024 {
			t.Skip()
		}
		initCfg := Input{
			Password:            []byte("password"),
			SelfID:              []byte("initiator"),
			PeerID:              []byte("responder"),
			Context:             ctx,
			SessionID:           sid,
			LocalAssociatedData: ada,
		}
		respCfg := Input{
			Password:            []byte("password"),
			SelfID:              []byte("responder"),
			PeerID:              []byte("initiator"),
			Context:             ctx,
			SessionID:           sid,
			LocalAssociatedData: adb,
		}
		initiator, msgA, err := startTestInitiator(initCfg)
		if err != nil {
			t.Fatalf("Start failed for bounded valid input: %v", err)
		}
		responder, msgB, err := respondTestResponder(respCfg, msgA)
		if err != nil {
			t.Fatalf("Respond failed for matching input: %v", err)
		}
		msgC, sI, err := initiator.Finish(msgB)
		if err != nil {
			t.Fatalf("initiator Finish failed for matching input: %v", err)
		}
		sR, err := responder.Finish(msgC)
		if err != nil {
			t.Fatalf("responder finish failed after initiator confirmation: %v", err)
		}
		if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
			t.Fatalf("transcript mismatch")
		}
	})
}

func FuzzProtocolMismatch(f *testing.F) {
	f.Add([]byte("sid"), []byte("ctx"), []byte("ADa"), []byte("ADb"))
	f.Add([]byte("sid2"), []byte{}, []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, sid, ctx, ada, adb []byte) {
		// respCfg.Context appends one byte below; ada/adb use a fuzz-budget cap.
		if len(sid) == 0 || len(sid) > sessionIDCap.length || len(ctx) >= contextCap.length || len(ada) > 1024 || len(adb) > 1024 {
			t.Skip()
		}
		initCfg := Input{
			Password:            []byte("password"),
			SelfID:              []byte("initiator"),
			PeerID:              []byte("responder"),
			Context:             ctx,
			SessionID:           sid,
			LocalAssociatedData: ada,
		}
		respCfg := Input{
			Password:            []byte("password"),
			SelfID:              []byte("responder"),
			PeerID:              []byte("initiator"),
			Context:             append(clone(ctx), 0xff),
			SessionID:           sid,
			LocalAssociatedData: adb,
		}
		initiator, msgA, err := startTestInitiator(initCfg)
		if err != nil {
			t.Fatalf("Start failed for bounded valid input: %v", err)
		}
		_, msgB, err := respondTestResponder(respCfg, msgA)
		if err != nil {
			t.Fatalf("Respond failed before expected confirmation mismatch: %v", err)
		}
		if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrConfirmationFailed) {
			t.Fatalf("Finish err=%v", err)
		}
	})
}

func FuzzRespondWithFuzzedMessageA(f *testing.F) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(f, initInput, respInput)
	invalid := fuzzDraftInvalidVector(f)
	for _, seed := range messageAProtocolFuzzSeeds(exchange.msgA, exchange.msgB, invalid.InvalidY1) {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, messageA []byte) {
		if len(messageA) > fuzzProtocolInputCap {
			t.Skip()
		}
		_, respInput := defaultExchangeInputs()
		_, msgB, err := respondWithRandom(respInput, messageA, repeatingRand(0x22))
		if err == nil {
			if _, err := decodeMessageB(msgB); err != nil {
				t.Fatalf("Respond returned malformed message B: %v", err)
			}
		}
	})
}

func FuzzInitiatorFinishWithFuzzedMessageB(f *testing.F) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(f, initInput, respInput)
	msgC, _ := exchange.finishInitiator()
	invalid := fuzzDraftInvalidVector(f)
	for _, seed := range messageBFuzzSeeds(exchange.msgB, msgC, invalid.InvalidY1) {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, messageB []byte) {
		if len(messageB) > fuzzProtocolInputCap {
			t.Skip()
		}
		initInput, _ := defaultExchangeInputs()
		initiator, _, err := startTestInitiator(initInput)
		if err != nil {
			t.Fatalf("Start failed for fixed fuzz config: %v", err)
		}
		msgC, sess, err := initiator.Finish(messageB)
		if err == nil {
			if sess == nil {
				t.Fatalf("Finish returned nil session without error")
			}
			if _, err := decodeMessageC(msgC); err != nil {
				t.Fatalf("Finish returned malformed message C: %v", err)
			}
		}
	})
}

func FuzzResponderFinishWithFuzzedMessageC(f *testing.F) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(f, initInput, respInput)
	msgC, _ := exchange.finishInitiator()
	for _, seed := range messageCFuzzSeeds(msgC, exchange.msgA) {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, messageC []byte) {
		if len(messageC) > fuzzProtocolInputCap {
			t.Skip()
		}
		initInput, respInput := defaultExchangeInputs()
		exchange := newExchange(t, initInput, respInput)
		sess, err := exchange.responder.Finish(messageC)
		if err == nil && sess == nil {
			t.Fatalf("Finish returned nil session without error")
		}
	})
}

func FuzzScalarMultVFY(f *testing.F) {
	invalid := fuzzDraftInvalidVector(f)
	validX := invalid.Valid["X"]
	if len(validX) == pointSize {
		f.Add(validX)
		f.Add(validX[:pointSize-1])
	}
	if len(invalid.InvalidY1) > 0 {
		f.Add(invalid.InvalidY1)
	}
	if len(invalid.InvalidY2) > 0 {
		f.Add(invalid.InvalidY2)
	}
	f.Add([]byte{})
	f.Add(bytes.Repeat([]byte{0xff}, pointSize))

	s, err := scalarFromCanonical(invalid.Valid["s"])
	if err != nil {
		f.Fatalf("invalid scalar fixture: %v", err)
	}
	f.Fuzz(func(t *testing.T, encoded []byte) {
		if len(encoded) > 128 {
			t.Skip()
		}
		out, err := scalarMultVFY(s, encoded)
		if err == nil {
			if len(out) != pointSize {
				t.Fatalf("scalarMultVFY output length=%d", len(out))
			}
			if bytes.Equal(out, identityEncoding) {
				t.Fatalf("scalarMultVFY accepted identity output")
			}
			return
		}
		if out != nil {
			t.Fatalf("scalarMultVFY rejection out=%x want nil", out)
		}
		if !errors.Is(err, ErrAbort) {
			t.Fatalf("scalarMultVFY rejection err=%v does not wrap ErrAbort", err)
		}
		// A canonical decode round-trips its input, and the harness scalar is
		// the fixed non-zero draft-fixture scalar, so the post-multiply
		// identity branch is unreachable and the rejection cause is fully
		// determined by the encoded bytes: wrong length is the internal
		// defensive branch, the canonical identity encoding is the identity
		// sentinel, and anything else 32 bytes long failed canonical
		// decoding. Do not fuzz the scalar without revisiting this oracle.
		switch {
		case len(encoded) != pointSize:
			if errors.Is(err, ErrPeerShareEncoding) || errors.Is(err, ErrPeerShareIdentity) {
				t.Fatalf("length rejection err=%v wraps a peer-share sentinel", err)
			}
		case bytes.Equal(encoded, identityEncoding):
			if !errors.Is(err, ErrPeerShareIdentity) || errors.Is(err, ErrPeerShareEncoding) {
				t.Fatalf("identity rejection err=%v want ErrPeerShareIdentity only", err)
			}
		default:
			if !errors.Is(err, ErrPeerShareEncoding) || errors.Is(err, ErrPeerShareIdentity) {
				t.Fatalf("encoding rejection err=%v want ErrPeerShareEncoding only", err)
			}
		}
	})
}

func FuzzMessageARoundTrip(f *testing.F) {
	f.Add([]byte("sid"), exactMessageFieldBytes(messageASpec, pointSize, 0x42, 0), []byte("ADa"))
	f.Add([]byte{}, identityEncoding, []byte{})
	f.Add(bytes.Repeat([]byte{0x01}, 8), exactMessageFieldBytes(messageASpec, pointSize, 0x02, -1), bytes.Repeat([]byte{0x03}, 8))
	f.Fuzz(func(t *testing.T, sid, ya, ada []byte) {
		if len(sid) > fuzzProtocolInputCap || len(ya) > fuzzProtocolInputCap || len(ada) > fuzzProtocolInputCap {
			t.Skip()
		}
		msg := encodeMessageA(sid, ya, ada)
		got, err := decodeMessageA(msg)
		if !messageFieldsMatchFramingShape(messageASpec, sid, ya, ada) {
			if err == nil {
				t.Fatalf("decodeMessageA accepted lengths sid=%d ya=%d ada=%d", len(sid), len(ya), len(ada))
			}
			return
		}
		if err != nil {
			t.Fatalf("decodeMessageA round trip failed: %v", err)
		}
		if !bytes.Equal(got.sid, sid) || !bytes.Equal(got.ya, ya) || !bytes.Equal(got.ada, ada) {
			t.Fatalf("message A round trip mismatch")
		}
	})
}

func FuzzMessageBRoundTrip(f *testing.F) {
	f.Add(exactMessageFieldBytes(messageBSpec, pointSize, 0x42, 0), []byte("ADb"), exactMessageFieldBytes(messageBSpec, tagSize, 0x99, 0))
	f.Add(identityEncoding, []byte{}, exactMessageFieldBytes(messageBSpec, tagSize, 0x00, 0))
	f.Add(exactMessageFieldBytes(messageBSpec, pointSize, 0x02, -1), bytes.Repeat([]byte{0x03}, 8), exactMessageFieldBytes(messageBSpec, tagSize, 0x04, -1))
	f.Fuzz(func(t *testing.T, yb, adb, tag []byte) {
		if len(yb) > fuzzProtocolInputCap || len(adb) > fuzzProtocolInputCap || len(tag) > fuzzProtocolInputCap {
			t.Skip()
		}
		msg := encodeMessageB(yb, adb, tag)
		got, err := decodeMessageB(msg)
		if !messageFieldsMatchFramingShape(messageBSpec, yb, adb, tag) {
			if err == nil {
				t.Fatalf("decodeMessageB accepted lengths yb=%d adb=%d tag=%d", len(yb), len(adb), len(tag))
			}
			return
		}
		if err != nil {
			t.Fatalf("decodeMessageB round trip failed: %v", err)
		}
		if !bytes.Equal(got.yb, yb) || !bytes.Equal(got.adb, adb) || !bytes.Equal(got.tag, tag) {
			t.Fatalf("message B round trip mismatch")
		}
	})
}

func FuzzMessageCRoundTrip(f *testing.F) {
	f.Add(exactMessageFieldBytes(messageCSpec, tagSize, 0x99, 0))
	f.Add(exactMessageFieldBytes(messageCSpec, tagSize, 0x00, 0))
	f.Add(exactMessageFieldBytes(messageCSpec, tagSize, 0x04, -1))
	f.Fuzz(func(t *testing.T, tag []byte) {
		if len(tag) > fuzzProtocolInputCap {
			t.Skip()
		}
		msg := encodeMessageC(tag)
		got, err := decodeMessageC(msg)
		if !messageFieldsMatchFramingShape(messageCSpec, tag) {
			if err == nil {
				t.Fatalf("decodeMessageC accepted tag length %d", len(tag))
			}
			return
		}
		if err != nil {
			t.Fatalf("decodeMessageC round trip failed: %v", err)
		}
		if !bytes.Equal(got.tag, tag) {
			t.Fatalf("message C round trip mismatch")
		}
	})
}

func fuzzDraftInvalidVector(tb testing.TB) draftInvalidVector {
	tb.Helper()
	v, err := loadDraftInvalidVectorJSON(draft21RistrettoInvalidJSON)
	if err != nil {
		tb.Fatalf("invalid vector fixture failed to load: %v", err)
	}
	return v
}
