package cpace

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"
)

type repeatingReader struct {
	buf []byte
	off int
}

func (r *repeatingReader) Read(p []byte) (int, error) {
	for i := range p {
		if len(r.buf) == 0 {
			p[i] = 1
			continue
		}
		p[i] = r.buf[r.off%len(r.buf)]
		r.off++
	}
	return len(p), nil
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func testConfig() Config {
	return Config{
		Password:    []byte("password"),
		InitiatorID: []byte("initiator"),
		ResponderID: []byte("responder"),
		Context:     []byte("context"),
		SessionID:   []byte("sid"),
	}
}

func TestConfirmedExchangeAndExport(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")
	respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}

	initiator, msgA, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := Respond(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	msgC, sI, err := initiator.Finish(msgB)
	if err != nil {
		t.Fatal(err)
	}
	sR, err := responder.Finish(msgC)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
		t.Fatalf("transcript ids differ")
	}
	kI, err := sI.Export([]byte("label"), []byte("ctx"), 64)
	if err != nil {
		t.Fatal(err)
	}
	kR, err := sR.Export([]byte("label"), []byte("ctx"), 64)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(kI, kR) {
		t.Fatalf("exports differ")
	}
}

func TestMutableInputsAreCopied(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	password := []byte("password")
	initCfg.Password = password
	initiator, msgA, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := range password {
		password[i] ^= 0xff
	}
	for i := range initCfg.AssociatedData {
		initCfg.AssociatedData[i] ^= 0xff
	}

	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")
	respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
	responder, msgB, err := Respond(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	msgC, _, err := initiator.Finish(msgB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := responder.Finish(msgC); err != nil {
		t.Fatal(err)
	}
}

func TestInputValidation(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Config)
	}{
		{"password", func(c *Config) { c.Password = nil }},
		{"initiator", func(c *Config) { c.InitiatorID = nil }},
		{"responder", func(c *Config) { c.ResponderID = nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			tc.edit(&cfg)
			if _, _, err := Start(cfg); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Start err=%v", err)
			}
		})
	}
}

func TestScalarSamplingRejectsRepeatedZero(t *testing.T) {
	cfg := testConfig()
	cfg.Rand = &repeatingReader{buf: []byte{0}}
	if _, _, err := Start(cfg); !errors.Is(err, ErrRandomness) || errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Start err=%v", err)
	}
}

func TestScalarSamplingWrapsRandomnessReadFailure(t *testing.T) {
	cfg := testConfig()
	cfg.Rand = failingReader{err: io.ErrUnexpectedEOF}
	if _, _, err := Start(cfg); !errors.Is(err, ErrRandomness) ||
		errors.Is(err, ErrInvalidInput) ||
		!errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Start err=%v", err)
	}
}

func TestProtocolCompletesWithEmptySessionID(t *testing.T) {
	cases := []struct {
		name    string
		initSID []byte
		respSID []byte
	}{
		{"nil nil", nil, nil},
		{"empty empty", []byte{}, []byte{}},
		{"nil empty", nil, []byte{}},
		{"empty nil", []byte{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			initCfg := testConfig()
			initCfg.SessionID = tc.initSID
			initCfg.AssociatedData = []byte("ADa")
			initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
			respCfg := testConfig()
			respCfg.SessionID = tc.respSID
			respCfg.AssociatedData = []byte("ADb")
			respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
			sI, sR := completeExchange(t, initCfg, respCfg)
			if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
				t.Fatalf("transcript IDs differ")
			}
		})
	}
}

func TestProtocolRejectsAsymmetricSessionID(t *testing.T) {
	cases := []struct {
		name    string
		initSID []byte
		respSID []byte
	}{
		{"initiator non-empty responder nil", []byte("sid"), nil},
		{"initiator nil responder non-empty", nil, []byte("sid")},
		{"initiator non-empty responder empty", []byte("sid"), []byte{}},
		{"initiator empty responder non-empty", []byte{}, []byte("sid")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			initCfg := testConfig()
			initCfg.SessionID = tc.initSID
			initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
			respCfg := testConfig()
			respCfg.SessionID = tc.respSID
			respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
			_, msgA, err := Start(initCfg)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := Respond(respCfg, msgA); !errors.Is(err, ErrMessage) {
				t.Fatalf("Respond err=%v", err)
			}
		})
	}
}

func TestConfirmationFailsOnBoundInputMismatch(t *testing.T) {
	initCfg := testConfig()
	initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	respCfg := testConfig()
	respCfg.Context = []byte("different")
	respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}

	initiator, msgA, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	_, msgB, err := Respond(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrConfirmationFailed) {
		t.Fatalf("Finish err=%v", err)
	}
}

func TestTranscriptLockingMismatches(t *testing.T) {
	cases := []struct {
		name       string
		editResp   func(*Config)
		tamperA    func([]byte) []byte
		tamperB    func([]byte) []byte
		respondErr error
		finishErr  error
	}{
		{
			name:      "initiator identity",
			editResp:  func(c *Config) { c.InitiatorID = []byte("other initiator") },
			finishErr: ErrConfirmationFailed,
		},
		{
			name:      "responder identity",
			editResp:  func(c *Config) { c.ResponderID = []byte("other responder") },
			finishErr: ErrConfirmationFailed,
		},
		{
			name:      "context",
			editResp:  func(c *Config) { c.Context = []byte("other context") },
			finishErr: ErrConfirmationFailed,
		},
		{
			name:       "session id",
			editResp:   func(c *Config) { c.SessionID = []byte("other sid") },
			respondErr: ErrMessage,
		},
		{
			name: "ADa",
			tamperA: func(msg []byte) []byte {
				a, err := decodeMessageA(msg)
				if err != nil {
					panic(err)
				}
				return encodeMessageA(a.sid, a.ya, []byte("tampered ADa"))
			},
			finishErr: ErrConfirmationFailed,
		},
		{
			name: "ADb",
			tamperB: func(msg []byte) []byte {
				b, err := decodeMessageB(msg)
				if err != nil {
					panic(err)
				}
				return encodeMessageB(b.yb, []byte("tampered ADb"), b.tag)
			},
			finishErr: ErrConfirmationFailed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			initCfg := testConfig()
			initCfg.AssociatedData = []byte("ADa")
			initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
			respCfg := testConfig()
			respCfg.AssociatedData = []byte("ADb")
			respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
			if tc.editResp != nil {
				tc.editResp(&respCfg)
			}
			initiator, msgA, err := Start(initCfg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.tamperA != nil {
				msgA = tc.tamperA(msgA)
			}
			_, msgB, err := Respond(respCfg, msgA)
			if tc.respondErr != nil {
				if !errors.Is(err, tc.respondErr) {
					t.Fatalf("Respond err=%v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.tamperB != nil {
				msgB = tc.tamperB(msgB)
			}
			_, _, err = initiator.Finish(msgB)
			if !errors.Is(err, tc.finishErr) {
				t.Fatalf("Finish err=%v", err)
			}
		})
	}
}

func TestMessageParserRejectsMalformed(t *testing.T) {
	cfg := testConfig()
	cfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	_, msgA, err := Start(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		msg  []byte
	}{
		{"truncated", msgA[:len(msgA)-1]},
		{"trailing", append(clone(msgA), 0)},
		{"format", append([]byte{0}, msgA[1:]...)},
		{"suite", append([]byte{msgA[0], 0}, msgA[2:]...)},
		{"role", append([]byte{msgA[0], msgA[1], roleB}, msgA[3:]...)},
		{"swapped format suite", append([]byte{wireSuite, wireFormatV1}, msgA[2:]...)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decodeMessageA(tc.msg); err == nil {
				t.Fatalf("decode succeeded")
			}
		})
	}
}

func TestMessageParsersRejectMalformedBAndC(t *testing.T) {
	msgB := encodeMessageB(bytes.Repeat([]byte{0x42}, pointSize), []byte("ADb"), bytes.Repeat([]byte{0x99}, tagSize))
	casesB := []struct {
		name string
		msg  []byte
	}{
		{"truncated", msgB[:len(msgB)-1]},
		{"trailing", append(clone(msgB), 0)},
		{"format", append([]byte{0}, msgB[1:]...)},
		{"suite", append([]byte{msgB[0], 0}, msgB[2:]...)},
		{"role", append([]byte{msgB[0], msgB[1], roleA}, msgB[3:]...)},
		{"swapped format suite", append([]byte{wireSuite, wireFormatV1}, msgB[2:]...)},
		{"point len", encodeMessageB(bytes.Repeat([]byte{0x42}, pointSize-1), nil, bytes.Repeat([]byte{0x99}, tagSize))},
		{"tag len", encodeMessageB(bytes.Repeat([]byte{0x42}, pointSize), nil, bytes.Repeat([]byte{0x99}, tagSize-1))},
		{"leb128", []byte{wireFormatV1, wireSuite, roleB, 0x80, 0x00}},
		{"oversized", append([]byte{wireFormatV1, wireSuite, roleB}, encodeLEB128(maxFieldLength+1)...)},
	}
	for _, tc := range casesB {
		t.Run("B "+tc.name, func(t *testing.T) {
			if _, err := decodeMessageB(tc.msg); err == nil {
				t.Fatalf("decode B succeeded")
			}
		})
	}

	msgC := encodeMessageC(bytes.Repeat([]byte{0x99}, tagSize))
	casesC := []struct {
		name string
		msg  []byte
	}{
		{"truncated", msgC[:len(msgC)-1]},
		{"trailing", append(clone(msgC), 0)},
		{"format", append([]byte{0}, msgC[1:]...)},
		{"suite", append([]byte{msgC[0], 0}, msgC[2:]...)},
		{"role", append([]byte{msgC[0], msgC[1], roleA}, msgC[3:]...)},
		{"swapped format suite", append([]byte{wireSuite, wireFormatV1}, msgC[2:]...)},
		{"tag len", encodeMessageC(bytes.Repeat([]byte{0x99}, tagSize-1))},
		{"leb128", []byte{wireFormatV1, wireSuite, roleC, 0x80, 0x00}},
		{"oversized", append([]byte{wireFormatV1, wireSuite, roleC}, encodeLEB128(maxFieldLength+1)...)},
	}
	for _, tc := range casesC {
		t.Run("C "+tc.name, func(t *testing.T) {
			if _, err := decodeMessageC(tc.msg); err == nil {
				t.Fatalf("decode C succeeded")
			}
		})
	}
}

func TestStateReuseAndConcurrentFinish(t *testing.T) {
	initCfg := testConfig()
	initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	respCfg := testConfig()
	respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
	initiator, msgA, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := Respond(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := initiator.Finish(msgB)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	var used int
	var ok int
	var msgC []byte
	for err := range errs {
		if errors.Is(err, ErrStateUsed) {
			used++
		} else if err == nil {
			ok++
		} else {
			t.Fatalf("unexpected err %v", err)
		}
	}
	if ok != 1 || used != 1 {
		t.Fatalf("ok=%d used=%d", ok, used)
	}

	// A fresh initiator produces the message C needed to exercise responder reuse.
	initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	initiator2, _, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	msgC, _, err = initiator2.Finish(msgB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := responder.Finish(msgC); err != nil {
		t.Fatal(err)
	}
	if _, err := responder.Finish(msgC); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("second responder finish err=%v", err)
	}
}

func TestProtocolAbortsOnInvalidRistrettoEncoding(t *testing.T) {
	cfg := testConfig()
	cfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
	invalid := mustLoadDraftInvalidVector(t)
	badA := encodeMessageA([]byte("sid"), invalid.InvalidY1, nil)
	if _, _, err := Respond(cfg, badA); !errors.Is(err, ErrAbort) {
		t.Fatalf("Respond err=%v", err)
	}
}

func TestInitiatorAbortsOnInvalidResponderShare(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	cases := []struct {
		name string
		yb   []byte
	}{
		{"non-canonical", invalid.InvalidY1},
		{"identity", invalid.InvalidY2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			cfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
			initiator, _, err := Start(cfg)
			if err != nil {
				t.Fatal(err)
			}
			msgB := encodeMessageB(tc.yb, nil, bytes.Repeat([]byte{0x99}, tagSize))
			if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrAbort) {
				t.Fatalf("Finish err=%v", err)
			}
		})
	}
}

func TestInitiatorReflectedShareFailsConfirmationNotAbort(t *testing.T) {
	cfg := testConfig()
	cfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	initiator, msgA, err := Start(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a, err := decodeMessageA(msgA)
	if err != nil {
		t.Fatal(err)
	}
	msgB := encodeMessageB(a.ya, nil, bytes.Repeat([]byte{0x99}, tagSize))
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrConfirmationFailed) {
		t.Fatalf("Finish err=%v", err)
	}
}

func TestResponderRejectsTamperedMessageC(t *testing.T) {
	initCfg := testConfig()
	initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	respCfg := testConfig()
	respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
	initiator, msgA, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := Respond(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	msgC, _, err := initiator.Finish(msgB)
	if err != nil {
		t.Fatal(err)
	}
	msgC[len(msgC)-1] ^= 0x01
	if _, err := responder.Finish(msgC); !errors.Is(err, ErrConfirmationFailed) {
		t.Fatalf("Finish err=%v", err)
	}
}

func TestFinishConsumesStateOnParseFailure(t *testing.T) {
	initCfg := testConfig()
	initCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x11}, 32)}
	respCfg := testConfig()
	respCfg.Rand = &repeatingReader{buf: bytes.Repeat([]byte{0x22}, 32)}
	initiator, msgA, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := Respond(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := initiator.Finish([]byte("garbage")); !errors.Is(err, ErrMessage) {
		t.Fatalf("initiator Finish garbage err=%v", err)
	}
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("initiator second Finish err=%v", err)
	}

	if _, err := responder.Finish([]byte("garbage")); !errors.Is(err, ErrMessage) {
		t.Fatalf("responder Finish garbage err=%v", err)
	}
	msgC := encodeMessageC(bytes.Repeat([]byte{0x99}, tagSize))
	if _, err := responder.Finish(msgC); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("responder second Finish err=%v", err)
	}
}

func mustLoadDraftInvalidVector(t *testing.T) draftInvalidVector {
	t.Helper()
	v, err := loadDraftInvalidVectorJSON(draft21RistrettoInvalidJSON)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func completeExchange(t *testing.T, initCfg, respCfg Config) (*Session, *Session) {
	t.Helper()
	initiator, msgA, err := Start(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := Respond(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	msgC, sI, err := initiator.Finish(msgB)
	if err != nil {
		t.Fatal(err)
	}
	sR, err := responder.Finish(msgC)
	if err != nil {
		t.Fatal(err)
	}
	return sI, sR
}
