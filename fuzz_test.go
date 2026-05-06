package cpace

import (
	"bytes"
	"errors"
	"testing"
)

const fuzzProtocolInputCap = 4096

func FuzzDecodeMessageA(f *testing.F) {
	_, _, _, msgA, msgB, _ := makeFuzzExchange(f)
	invalid := fuzzDraftInvalidVector(f)
	f.Add(msgA)
	f.Add(msgA[:len(msgA)-1])
	f.Add(withFuzzRole(msgA, roleB))
	f.Add([]byte{wireFormatV1, wireSuite, roleA, 0x80, 0x00})
	f.Add(encodeMessageA([]byte("sid"), identityEncoding, nil))
	f.Add(encodeMessageA([]byte("sid"), invalid.InvalidY1, nil))
	f.Add(msgB)
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = decodeMessageA(in)
	})
}

func FuzzDecodeMessageB(f *testing.F) {
	_, _, _, _, msgB, msgC := makeFuzzExchange(f)
	invalid := fuzzDraftInvalidVector(f)
	f.Add(msgB)
	f.Add(msgB[:len(msgB)-1])
	f.Add(withFuzzRole(msgB, roleA))
	f.Add([]byte{wireFormatV1, wireSuite, roleB, 0x80, 0x00})
	f.Add(encodeMessageB(identityEncoding, nil, make([]byte, tagSize)))
	f.Add(encodeMessageB(invalid.InvalidY1, nil, make([]byte, tagSize)))
	f.Add(withFuzzTamperedLastByte(msgB))
	f.Add(msgC)
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = decodeMessageB(in)
	})
}

func FuzzDecodeMessageC(f *testing.F) {
	_, _, _, msgA, _, msgC := makeFuzzExchange(f)
	f.Add(msgC)
	f.Add(msgC[:len(msgC)-1])
	f.Add(withFuzzRole(msgC, roleA))
	f.Add([]byte{wireFormatV1, wireSuite, roleC, 0x80, 0x00})
	f.Add(encodeMessageC(make([]byte, tagSize-1)))
	f.Add(withFuzzTamperedLastByte(msgC))
	f.Add(msgA)
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
		if len(sid) == 0 || len(sid) > 1024 || len(ctx) > 1024 || len(ada) > 1024 || len(adb) > 1024 {
			t.Skip()
		}
		initCfg := Config{
			Password:       []byte("password"),
			InitiatorID:    []byte("initiator"),
			ResponderID:    []byte("responder"),
			Context:        ctx,
			SessionID:      sid,
			AssociatedData: ada,
		}
		respCfg := initCfg
		respCfg.AssociatedData = adb
		initiator, msgA, err := startTestInitiator(initCfg)
		if err != nil {
			t.Fatalf("Start failed for bounded valid config: %v", err)
		}
		responder, msgB, err := respondTestResponder(respCfg, msgA)
		if err != nil {
			t.Fatalf("Respond failed for matching config: %v", err)
		}
		msgC, sI, err := initiator.Finish(msgB)
		if err != nil {
			t.Fatalf("initiator Finish failed for matching config: %v", err)
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
		if len(sid) == 0 || len(sid) > maxSessionIDLength || len(ctx) >= maxContextLength || len(ada) > 1024 || len(adb) > 1024 {
			t.Skip()
		}
		initCfg := Config{
			Password:       []byte("password"),
			InitiatorID:    []byte("initiator"),
			ResponderID:    []byte("responder"),
			Context:        ctx,
			SessionID:      sid,
			AssociatedData: ada,
		}
		respCfg := initCfg
		respCfg.Context = append(clone(ctx), 0xff)
		respCfg.AssociatedData = adb
		initiator, msgA, err := startTestInitiator(initCfg)
		if err != nil {
			t.Fatalf("Start failed for bounded valid config: %v", err)
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
	_, _, _, msgA, msgB, _ := makeFuzzExchange(f)
	invalid := fuzzDraftInvalidVector(f)
	f.Add(msgA)
	f.Add(msgA[:len(msgA)-1])
	f.Add(withFuzzRole(msgA, roleB))
	f.Add([]byte{wireFormatV1, wireSuite, roleA, 0x80, 0x00})
	f.Add(encodeMessageA([]byte("sid"), identityEncoding, nil))
	f.Add(encodeMessageA([]byte("sid"), invalid.InvalidY1, nil))
	f.Add(encodeMessageA([]byte("other sid"), bytes.Repeat([]byte{0x42}, pointSize), nil))
	f.Add(msgB)
	f.Fuzz(func(t *testing.T, messageA []byte) {
		if len(messageA) > fuzzProtocolInputCap {
			t.Skip()
		}
		cfg := fuzzResponderConfig()
		_, msgB, err := respondWithRandom(cfg, messageA, repeatingRand(0x22))
		if err == nil {
			if _, err := decodeMessageB(msgB); err != nil {
				t.Fatalf("Respond returned malformed message B: %v", err)
			}
		}
	})
}

func FuzzInitiatorFinishWithFuzzedMessageB(f *testing.F) {
	_, _, _, _, msgB, msgC := makeFuzzExchange(f)
	invalid := fuzzDraftInvalidVector(f)
	f.Add(msgB)
	f.Add(msgB[:len(msgB)-1])
	f.Add(withFuzzRole(msgB, roleA))
	f.Add([]byte{wireFormatV1, wireSuite, roleB, 0x80, 0x00})
	f.Add(encodeMessageB(identityEncoding, nil, make([]byte, tagSize)))
	f.Add(encodeMessageB(invalid.InvalidY1, nil, make([]byte, tagSize)))
	f.Add(withFuzzTamperedLastByte(msgB))
	f.Add(msgC)
	f.Fuzz(func(t *testing.T, messageB []byte) {
		if len(messageB) > fuzzProtocolInputCap {
			t.Skip()
		}
		initiator, _, err := startTestInitiator(fuzzInitiatorConfig())
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
	_, _, _, msgA, _, msgC := makeFuzzExchange(f)
	f.Add(msgC)
	f.Add(msgC[:len(msgC)-1])
	f.Add(withFuzzRole(msgC, roleA))
	f.Add([]byte{wireFormatV1, wireSuite, roleC, 0x80, 0x00})
	f.Add(encodeMessageC(make([]byte, tagSize-1)))
	f.Add(withFuzzTamperedLastByte(msgC))
	f.Add(msgA)
	f.Fuzz(func(t *testing.T, messageC []byte) {
		if len(messageC) > fuzzProtocolInputCap {
			t.Skip()
		}
		_, responder, _, _, _, _ := makeFuzzExchange(t)
		sess, err := responder.Finish(messageC)
		if err == nil && sess == nil {
			t.Fatalf("Finish returned nil session without error")
		}
	})
}

func FuzzScalarMultVFY(f *testing.F) {
	invalid := fuzzDraftInvalidVector(f)
	f.Add(invalid.Valid["X"])
	f.Add(invalid.InvalidY1)
	f.Add(invalid.InvalidY2)
	f.Add([]byte{})
	f.Add(invalid.Valid["X"][:pointSize-1])
	f.Add(bytes.Repeat([]byte{0xff}, pointSize))

	s, err := scalarFromCanonical(invalid.Valid["s"])
	if err != nil {
		f.Fatalf("invalid scalar fixture: %v", err)
	}
	f.Fuzz(func(t *testing.T, encoded []byte) {
		if len(encoded) > 128 {
			t.Skip()
		}
		out, ok := scalarMultVFY(s, encoded)
		if len(out) != pointSize {
			t.Fatalf("scalarMultVFY output length=%d", len(out))
		}
		if ok && bytes.Equal(out, identityEncoding) {
			t.Fatalf("scalarMultVFY accepted identity output")
		}
		if !ok && !bytes.Equal(out, identityEncoding) {
			t.Fatalf("scalarMultVFY rejection output=%x", out)
		}
	})
}

func FuzzMessageARoundTrip(f *testing.F) {
	f.Add([]byte("sid"), bytes.Repeat([]byte{0x42}, pointSize), []byte("ADa"))
	f.Add([]byte{}, identityEncoding, []byte{})
	f.Add(bytes.Repeat([]byte{0x01}, 8), bytes.Repeat([]byte{0x02}, pointSize-1), bytes.Repeat([]byte{0x03}, 8))
	f.Fuzz(func(t *testing.T, sid, ya, ada []byte) {
		if len(sid) > fuzzProtocolInputCap || len(ya) > fuzzProtocolInputCap || len(ada) > fuzzProtocolInputCap {
			t.Skip()
		}
		msg := encodeMessageA(sid, ya, ada)
		got, err := decodeMessageA(msg)
		if len(sid) > maxSessionIDLength || len(ya) != pointSize || len(ada) > maxAssociatedDataLength {
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
	f.Add(bytes.Repeat([]byte{0x42}, pointSize), []byte("ADb"), bytes.Repeat([]byte{0x99}, tagSize))
	f.Add(identityEncoding, []byte{}, bytes.Repeat([]byte{0x00}, tagSize))
	f.Add(bytes.Repeat([]byte{0x02}, pointSize-1), bytes.Repeat([]byte{0x03}, 8), bytes.Repeat([]byte{0x04}, tagSize-1))
	f.Fuzz(func(t *testing.T, yb, adb, tag []byte) {
		if len(yb) > fuzzProtocolInputCap || len(adb) > fuzzProtocolInputCap || len(tag) > fuzzProtocolInputCap {
			t.Skip()
		}
		msg := encodeMessageB(yb, adb, tag)
		got, err := decodeMessageB(msg)
		if len(yb) != pointSize || len(adb) > maxAssociatedDataLength || len(tag) != tagSize {
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
	f.Add(bytes.Repeat([]byte{0x99}, tagSize))
	f.Add(bytes.Repeat([]byte{0x00}, tagSize))
	f.Add(bytes.Repeat([]byte{0x04}, tagSize-1))
	f.Fuzz(func(t *testing.T, tag []byte) {
		if len(tag) > fuzzProtocolInputCap {
			t.Skip()
		}
		msg := encodeMessageC(tag)
		got, err := decodeMessageC(msg)
		if len(tag) != tagSize {
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

func makeFuzzExchange(tb testing.TB) (*Initiator, *Responder, *Session, []byte, []byte, []byte) {
	tb.Helper()
	initiator, msgA, err := startTestInitiator(fuzzInitiatorConfig())
	if err != nil {
		tb.Fatalf("Start failed for fixed fuzz config: %v", err)
	}
	responder, msgB, err := respondTestResponder(fuzzResponderConfig(), msgA)
	if err != nil {
		tb.Fatalf("Respond failed for fixed fuzz config: %v", err)
	}
	msgC, sess, err := initiator.Finish(msgB)
	if err != nil {
		tb.Fatalf("initiator Finish failed for fixed fuzz config: %v", err)
	}
	return initiator, responder, sess, msgA, msgB, msgC
}

func fuzzInitiatorConfig() Config {
	cfg := testConfig()
	cfg.AssociatedData = []byte("ADa")
	return cfg
}

func fuzzResponderConfig() Config {
	cfg := testConfig()
	cfg.AssociatedData = []byte("ADb")
	return cfg
}

func fuzzDraftInvalidVector(tb testing.TB) draftInvalidVector {
	tb.Helper()
	v, err := loadDraftInvalidVectorJSON(draft21RistrettoInvalidJSON)
	if err != nil {
		tb.Fatalf("invalid vector fixture failed to load: %v", err)
	}
	return v
}

func withFuzzRole(msg []byte, role byte) []byte {
	out := clone(msg)
	if len(out) > 2 {
		out[2] = role
	}
	return out
}

func withFuzzTamperedLastByte(msg []byte) []byte {
	out := clone(msg)
	if len(out) > 0 {
		out[len(out)-1] ^= 0x01
	}
	return out
}
