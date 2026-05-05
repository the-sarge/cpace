package cpace

import (
	"bytes"
	"errors"
	"testing"
)

func FuzzDecodeMessageA(f *testing.F) {
	f.Add([]byte{wireFormatV1, wireSuite, roleA, 0, pointSize})
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = decodeMessageA(in)
	})
}

func FuzzDecodeMessageB(f *testing.F) {
	f.Add(encodeMessageB(make([]byte, pointSize), nil, make([]byte, tagSize)))
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = decodeMessageB(in)
	})
}

func FuzzDecodeMessageC(f *testing.F) {
	f.Add(encodeMessageC(make([]byte, tagSize)))
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
	f.Fuzz(func(t *testing.T, sid, ctx, ada, adb []byte) {
		if len(sid) > 1024 || len(ctx) > 1024 || len(ada) > 1024 || len(adb) > 1024 {
			t.Skip()
		}
		initCfg := Config{
			Password:       []byte("password"),
			InitiatorID:    []byte("initiator"),
			ResponderID:    []byte("responder"),
			Context:        ctx,
			SessionID:      sid,
			AssociatedData: ada,
			Rand:           &repeatingReader{buf: []byte{1}},
		}
		respCfg := initCfg
		respCfg.AssociatedData = adb
		respCfg.Rand = &repeatingReader{buf: []byte{2}}
		initiator, msgA, err := Start(initCfg)
		if err != nil {
			t.Fatalf("Start failed for bounded valid config: %v", err)
		}
		responder, msgB, err := Respond(respCfg, msgA)
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
	f.Fuzz(func(t *testing.T, sid, ctx, ada, adb []byte) {
		if len(sid) > 1024 || len(ctx) > 1024 || len(ada) > 1024 || len(adb) > 1024 {
			t.Skip()
		}
		initCfg := Config{
			Password:       []byte("password"),
			InitiatorID:    []byte("initiator"),
			ResponderID:    []byte("responder"),
			Context:        ctx,
			SessionID:      sid,
			AssociatedData: ada,
			Rand:           &repeatingReader{buf: []byte{1}},
		}
		respCfg := initCfg
		respCfg.Context = append(clone(ctx), 0xff)
		respCfg.AssociatedData = adb
		respCfg.Rand = &repeatingReader{buf: []byte{2}}
		initiator, msgA, err := Start(initCfg)
		if err != nil {
			t.Fatalf("Start failed for bounded valid config: %v", err)
		}
		_, msgB, err := Respond(respCfg, msgA)
		if err != nil {
			t.Fatalf("Respond failed before expected confirmation mismatch: %v", err)
		}
		if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrConfirmationFailed) {
			t.Fatalf("Finish err=%v", err)
		}
	})
}
