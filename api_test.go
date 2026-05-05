package cpace

import (
	"bytes"
	"errors"
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
	if _, _, err := Start(cfg); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Start err=%v", err)
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
	for name, msg := range map[string][]byte{
		"truncated": msgA[:len(msgA)-1],
		"trailing":  append(clone(msgA), 0),
		"version":   append([]byte{0}, msgA[1:]...),
		"suite":     append([]byte{msgA[0], 0}, msgA[2:]...),
		"role":      append([]byte{msgA[0], msgA[1], roleB}, msgA[3:]...),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeMessageA(msg); err == nil {
				t.Fatalf("decode succeeded")
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
	badA := encodeMessageA([]byte("sid"), hx(t, "2b3c6b8c4f3800e7aef6864025b4ed79bd599117e427c41bd47d93d654b4a51c"), nil)
	if _, _, err := Respond(cfg, badA); !errors.Is(err, ErrAbort) {
		t.Fatalf("Respond err=%v", err)
	}
}
