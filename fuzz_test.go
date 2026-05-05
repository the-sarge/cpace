package cpace

import "testing"

func FuzzDecodeMessageA(f *testing.F) {
	f.Add([]byte{wireVersion, wireSuite, roleA, 0, pointSize})
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

func FuzzVectorJSONLoader(f *testing.F) {
	f.Add(draft21RistrettoVectorJSON)
	f.Fuzz(func(t *testing.T, in []byte) {
		_, _ = loadDraftVectorJSON(in)
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
		respCfg.AssociatedData = adb
		respCfg.Rand = &repeatingReader{buf: []byte{2}}
		initiator, msgA, err := Start(initCfg)
		if err != nil {
			return
		}
		responder, msgB, err := Respond(respCfg, msgA)
		if err != nil {
			return
		}
		msgC, sI, err := initiator.Finish(msgB)
		if err != nil {
			return
		}
		sR, err := responder.Finish(msgC)
		if err != nil {
			t.Fatalf("responder finish failed after initiator confirmation: %v", err)
		}
		if string(sI.TranscriptID()) != string(sR.TranscriptID()) {
			t.Fatalf("transcript mismatch")
		}
	})
}
