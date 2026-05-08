package cpace

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

type countingFailingReader struct {
	reads int
	err   error
}

func (r *countingFailingReader) Read([]byte) (int, error) {
	r.reads++
	return 0, r.err
}

func TestInternalRandomHelpersDefaultNilRandomness(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")

	initiator, msgA, err := startWithRandom(initCfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := respondWithRandom(respCfg, msgA, nil)
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
		t.Fatal("transcript ids differ")
	}
}

func TestConfirmedExchangeAndExport(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")

	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := respondTestResponder(respCfg, msgA)
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

func TestSessionPeerMetadata(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")

	sI, sR := completeExchange(t, initCfg, respCfg)
	if got := sI.PeerAssociatedData(); !bytes.Equal(got, respCfg.AssociatedData) {
		t.Fatalf("initiator peer AD=%q want %q", got, respCfg.AssociatedData)
	}
	if got := sR.PeerAssociatedData(); !bytes.Equal(got, initCfg.AssociatedData) {
		t.Fatalf("responder peer AD=%q want %q", got, initCfg.AssociatedData)
	}
	if got := sI.PeerID(); !bytes.Equal(got, initCfg.ResponderID) {
		t.Fatalf("initiator peer ID=%q want %q", got, initCfg.ResponderID)
	}
	if got := sR.PeerID(); !bytes.Equal(got, respCfg.InitiatorID) {
		t.Fatalf("responder peer ID=%q want %q", got, respCfg.InitiatorID)
	}

	peerAD := sI.PeerAssociatedData()
	peerAD[0] ^= 0xff
	if bytes.Equal(sI.PeerAssociatedData(), peerAD) {
		t.Fatal("PeerAssociatedData returned mutable session storage")
	}
	peerID := sR.PeerID()
	peerID[0] ^= 0xff
	if bytes.Equal(sR.PeerID(), peerID) {
		t.Fatal("PeerID returned mutable session storage")
	}

	emptySI, emptySR := completeExchange(t, testConfig(), testConfig())
	if got := emptySI.PeerAssociatedData(); len(got) != 0 {
		t.Fatalf("initiator empty peer AD=%q want empty", got)
	}
	if got := emptySR.PeerAssociatedData(); len(got) != 0 {
		t.Fatalf("responder empty peer AD=%q want empty", got)
	}
}

func TestSessionClose(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")

	sI, _ := completeExchange(t, initCfg, respCfg)
	transcriptID := sI.TranscriptID()
	peerAD := sI.PeerAssociatedData()
	peerID := sI.PeerID()
	if _, err := sI.Export([]byte("label"), []byte("ctx"), 32); err != nil {
		t.Fatal(err)
	}
	if err := sI.Close(); err != nil {
		t.Fatal(err)
	}
	if err := sI.Close(); err != nil {
		t.Fatal(err)
	}
	if sI.state.isk != nil {
		t.Fatal("session retained ISK after Close")
	}
	if _, err := sI.Export([]byte("label"), []byte("ctx"), 32); !errors.Is(err, ErrSessionClosed) {
		t.Fatalf("Export after Close err=%v", err)
	}
	if !bytes.Equal(sI.TranscriptID(), transcriptID) {
		t.Fatal("TranscriptID changed after Close")
	}
	if !bytes.Equal(sI.PeerAssociatedData(), peerAD) {
		t.Fatal("PeerAssociatedData changed after Close")
	}
	if !bytes.Equal(sI.PeerID(), peerID) {
		t.Fatal("PeerID changed after Close")
	}
}

func TestSessionValueCopiesShareCloseState(t *testing.T) {
	sI, _ := completeExchange(t, testConfig(), testConfig())
	copied := *sI
	if _, err := copied.Export([]byte("label"), []byte("ctx"), 32); err != nil {
		t.Fatal(err)
	}
	if err := sI.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := copied.Export([]byte("label"), []byte("ctx"), 32); !errors.Is(err, ErrSessionClosed) {
		t.Fatalf("Export from copied closed session err=%v", err)
	}
	if err := copied.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNilSessionClose(t *testing.T) {
	var s *Session
	if err := s.Close(); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Close nil err=%v", err)
	}
	if got := s.PeerAssociatedData(); got != nil {
		t.Fatalf("nil PeerAssociatedData=%q want nil", got)
	}
	if got := s.PeerID(); got != nil {
		t.Fatalf("nil PeerID=%q want nil", got)
	}
}

func TestSessionCloseConcurrentExport(t *testing.T) {
	sI, _ := completeExchange(t, testConfig(), testConfig())
	const workers = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	ready := make(chan struct{}, workers)
	errs := make(chan error, workers)
	var closed atomic.Int64
	for range workers {
		wg.Go(func() {
			<-start
			reportedReady := false
			for {
				_, err := sI.Export([]byte("label"), []byte("ctx"), 32)
				switch {
				case err == nil:
					if !reportedReady {
						ready <- struct{}{}
						reportedReady = true
					}
				case errors.Is(err, ErrSessionClosed):
					closed.Add(1)
					return
				default:
					errs <- err
					return
				}
			}
		})
	}
	close(start)
	for range workers {
		select {
		case <-ready:
		case err := <-errs:
			t.Fatalf("unexpected Export err=%v", err)
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for concurrent exports")
		}
	}
	if err := sI.Close(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected Export err=%v", err)
	}
	if got := closed.Load(); got != workers {
		t.Fatalf("ErrSessionClosed count=%d want %d", got, workers)
	}
}

func TestSessionCloseConcurrentClose(t *testing.T) {
	sI, _ := completeExchange(t, testConfig(), testConfig())
	const workers = 16
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for range workers {
		wg.Go(func() {
			if err := sI.Close(); err != nil {
				errs <- err
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("Close err=%v", err)
	}
	if sI.state.isk != nil {
		t.Fatal("session retained ISK after concurrent Close")
	}
}

func TestSessionMetadataConcurrentClose(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")
	sI, _ := completeExchange(t, initCfg, respCfg)
	transcriptID := sI.TranscriptID()
	peerAD := sI.PeerAssociatedData()
	peerID := sI.PeerID()

	const workers = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	ready := make(chan struct{}, workers)
	stop := make(chan struct{})
	errs := make(chan string, workers)
	for range workers {
		wg.Go(func() {
			<-start
			reportedReady := false
			for {
				select {
				case <-stop:
					return
				default:
				}
				if !bytes.Equal(sI.TranscriptID(), transcriptID) {
					errs <- "TranscriptID changed"
					return
				}
				if !bytes.Equal(sI.PeerAssociatedData(), peerAD) {
					errs <- "PeerAssociatedData changed"
					return
				}
				if !bytes.Equal(sI.PeerID(), peerID) {
					errs <- "PeerID changed"
					return
				}
				if !reportedReady {
					ready <- struct{}{}
					reportedReady = true
				}
			}
		})
	}
	close(start)
	for range workers {
		select {
		case <-ready:
		case msg := <-errs:
			t.Fatal(msg)
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for metadata readers")
		}
	}
	if err := sI.Close(); err != nil {
		t.Fatal(err)
	}
	close(stop)
	wg.Wait()
	close(errs)
	for msg := range errs {
		t.Fatal(msg)
	}
}

func TestMutableInputsAreCopied(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	password := []byte("password")
	initiatorPeerID := []byte("responder")
	initCfg.Password = password
	initCfg.ResponderID = initiatorPeerID
	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := range password {
		password[i] ^= 0xff
	}
	for i := range initCfg.AssociatedData {
		initCfg.AssociatedData[i] ^= 0xff
	}
	for i := range initiatorPeerID {
		initiatorPeerID[i] ^= 0xff
	}

	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")
	responderPeerID := []byte("initiator")
	respCfg.InitiatorID = responderPeerID
	responder, msgB, err := respondTestResponder(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	for i := range responderPeerID {
		responderPeerID[i] ^= 0xff
	}
	msgC, sI, err := initiator.Finish(msgB)
	if err != nil {
		t.Fatal(err)
	}
	sR, err := responder.Finish(msgC)
	if err != nil {
		t.Fatal(err)
	}
	if got := sI.PeerID(); !bytes.Equal(got, []byte("responder")) {
		t.Fatalf("initiator peer ID=%q after caller mutation", got)
	}
	if got := sR.PeerID(); !bytes.Equal(got, []byte("initiator")) {
		t.Fatalf("responder peer ID=%q after caller mutation", got)
	}
}

func TestFinishCleanupDoesNotAliasReturnedSessions(t *testing.T) {
	initCfg := testConfig()
	initCfg.AssociatedData = []byte("ADa")
	respCfg := testConfig()
	respCfg.AssociatedData = []byte("ADb")

	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	initiatorScalar := initiator.scalar
	responder, msgB, err := respondTestResponder(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	responderISK := responder.isk
	responderTranscript := responder.transcript

	msgC, sI, err := initiator.Finish(msgB)
	if err != nil {
		t.Fatal(err)
	}
	if initiator.scalar != nil {
		t.Fatal("initiator scalar reference retained after Finish")
	}
	if initiatorScalar == nil || !allZero(initiatorScalar.Bytes()) {
		t.Fatal("consumed initiator scalar was not cleared")
	}

	sR, err := responder.Finish(msgC)
	if err != nil {
		t.Fatal(err)
	}
	if responder.isk != nil || responder.transcript != nil {
		t.Fatal("responder retained cleared state references after Finish")
	}
	if !allZero(responderISK) {
		t.Fatal("responder-owned ISK was not cleared")
	}
	if !allZero(responderTranscript) {
		t.Fatal("responder-owned transcript was not cleared")
	}
	if allZero(sI.state.isk) || allZero(sR.state.isk) {
		t.Fatal("returned session ISK was cleared through an alias")
	}
	kI, err := sI.Export([]byte("label"), []byte("ctx"), 32)
	if err != nil {
		t.Fatal(err)
	}
	kR, err := sR.Export([]byte("label"), []byte("ctx"), 32)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(kI, kR) {
		t.Fatal("exports differ after cleanup")
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
		{"session id", func(c *Config) { c.SessionID = nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			tc.edit(&cfg)
			if _, _, err := startTestInitiator(cfg); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Start err=%v", err)
			}
		})
	}
}

func TestConfigFieldSizeLimits(t *testing.T) {
	cases := []struct {
		name string
		max  int
		edit func(*Config, []byte)
	}{
		{"password", maxPasswordLength, func(c *Config, b []byte) { c.Password = b }},
		{"initiator id", maxIDLength, func(c *Config, b []byte) { c.InitiatorID = b }},
		{"responder id", maxIDLength, func(c *Config, b []byte) { c.ResponderID = b }},
		{"context", maxContextLength, func(c *Config, b []byte) { c.Context = b }},
		{"session id", maxSessionIDLength, func(c *Config, b []byte) { c.SessionID = b }},
		{"associated data", maxAssociatedDataLength, func(c *Config, b []byte) { c.AssociatedData = b }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" max", func(t *testing.T) {
			cfg := testConfig()
			tc.edit(&cfg, bytes.Repeat([]byte{0x42}, tc.max))
			if _, _, err := startTestInitiator(cfg); err != nil {
				t.Fatalf("Start rejected max-size field: %v", err)
			}
		})
		t.Run(tc.name+" oversized", func(t *testing.T) {
			cfg := testConfig()
			tc.edit(&cfg, bytes.Repeat([]byte{0x42}, tc.max+1))
			if _, _, err := startTestInitiator(cfg); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Start err=%v", err)
			}
		})
	}
}

func TestScalarSamplingRejectsRepeatedZero(t *testing.T) {
	if _, err := sampleScalar(&repeatingReader{buf: []byte{0}}); !errors.Is(err, ErrRandomness) ||
		errors.Is(err, ErrInvalidInput) {
		t.Fatalf("sampleScalar err=%v", err)
	}
}

func TestScalarSamplingMasksDraftRistrettoBits(t *testing.T) {
	in := bytes.Repeat([]byte{0xff}, scalarSize)
	if _, err := scalarFromCanonical(in); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("unmasked all-ones scalar err=%v", err)
	}
	s, err := sampleScalar(&repeatingReader{buf: in})
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Repeat([]byte{0xff}, scalarSize)
	want[31] = 0x0f
	if got := s.Bytes(); !bytes.Equal(got, want) {
		t.Fatalf("sampled scalar=%x want %x", got, want)
	}
}

func TestScalarSamplingWrapsRandomnessReadFailure(t *testing.T) {
	if _, err := sampleScalar(failingReader{err: io.ErrUnexpectedEOF}); !errors.Is(err, ErrRandomness) ||
		errors.Is(err, ErrInvalidInput) ||
		!errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("sampleScalar err=%v", err)
	}
}

func TestProtocolRejectsEmptySessionIDByDefault(t *testing.T) {
	cases := []struct {
		name string
		sid  []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			cfg.SessionID = tc.sid
			if _, _, err := startTestInitiator(cfg); !errors.Is(err, ErrInvalidInput) ||
				!errors.Is(err, ErrEmptySessionID) {
				t.Fatalf("Start err=%v", err)
			}

			msgA := encodeMessageA(tc.sid, bytes.Repeat([]byte{0x42}, pointSize), nil)
			if _, _, err := respondTestResponder(cfg, msgA); !errors.Is(err, ErrInvalidInput) ||
				!errors.Is(err, ErrEmptySessionID) {
				t.Fatalf("Respond err=%v", err)
			}
		})
	}
}

func TestProtocolAllowsEmptySessionIDWithCompatibilityFlag(t *testing.T) {
	cases := []struct {
		name    string
		initSID []byte
		respSID []byte
	}{
		{"nil nil", nil, nil},
		{"empty empty", []byte{}, []byte{}},
		{"nil empty", nil, []byte{}},
		{"empty nil", []byte{}, nil},
		{"non-empty flag enabled", []byte("sid"), []byte("sid")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			initCfg := testConfig()
			initCfg.SessionID = tc.initSID
			initCfg.AllowEmptySessionID = true
			initCfg.AssociatedData = []byte("ADa")
			respCfg := testConfig()
			respCfg.SessionID = tc.respSID
			respCfg.AllowEmptySessionID = true
			respCfg.AssociatedData = []byte("ADb")
			sI, sR := completeExchange(t, initCfg, respCfg)
			if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
				t.Fatalf("transcript IDs differ")
			}
			kI, err := sI.Export([]byte("label"), []byte("ctx"), 32)
			if err != nil {
				t.Fatal(err)
			}
			kR, err := sR.Export([]byte("label"), []byte("ctx"), 32)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(kI, kR) {
				t.Fatal("exports differ")
			}
		})
	}
}

func TestProtocolRejectsAsymmetricSessionID(t *testing.T) {
	cases := []struct {
		name           string
		initSID        []byte
		respSID        []byte
		allowInitEmpty bool
		allowRespEmpty bool
	}{
		{
			name:    "different non-empty",
			initSID: []byte("initiator sid"),
			respSID: []byte("responder sid"),
		},
		{
			name:           "initiator empty compatibility responder non-empty",
			initSID:        nil,
			respSID:        []byte("sid"),
			allowInitEmpty: true,
		},
		{
			name:           "initiator non-empty responder empty compatibility",
			initSID:        []byte("sid"),
			respSID:        nil,
			allowRespEmpty: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			initCfg := testConfig()
			initCfg.SessionID = tc.initSID
			initCfg.AllowEmptySessionID = tc.allowInitEmpty
			respCfg := testConfig()
			respCfg.SessionID = tc.respSID
			respCfg.AllowEmptySessionID = tc.allowRespEmpty
			_, msgA, err := startTestInitiator(initCfg)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := respondTestResponder(respCfg, msgA); !errors.Is(err, ErrMessage) {
				t.Fatalf("Respond err=%v", err)
			}
		})
	}
}

func TestProtocolAllowsNonEmptySessionIDWithAsymmetricCompatibilityFlag(t *testing.T) {
	cases := []struct {
		name           string
		allowInitEmpty bool
		allowRespEmpty bool
	}{
		{"initiator flag only", true, false},
		{"responder flag only", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			initCfg := testConfig()
			initCfg.AllowEmptySessionID = tc.allowInitEmpty
			respCfg := testConfig()
			respCfg.AllowEmptySessionID = tc.allowRespEmpty
			sI, sR := completeExchange(t, initCfg, respCfg)
			if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
				t.Fatalf("transcript IDs differ")
			}
		})
	}
}

func TestConfirmationFailsOnBoundInputMismatch(t *testing.T) {
	initCfg := testConfig()
	respCfg := testConfig()
	respCfg.Context = []byte("different")

	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	_, msgB, err := respondTestResponder(respCfg, msgA)
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
			respCfg := testConfig()
			respCfg.AssociatedData = []byte("ADb")
			if tc.editResp != nil {
				tc.editResp(&respCfg)
			}
			initiator, msgA, err := startTestInitiator(initCfg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.tamperA != nil {
				msgA = tc.tamperA(msgA)
			}
			_, msgB, err := respondTestResponder(respCfg, msgA)
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
	_, msgA, err := startTestInitiator(cfg)
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

func TestMessageParserFieldSizeLimits(t *testing.T) {
	point := bytes.Repeat([]byte{0x42}, pointSize)
	tag := bytes.Repeat([]byte{0x99}, tagSize)

	msgA := encodeMessageA(bytes.Repeat([]byte{0x11}, maxSessionIDLength), point, bytes.Repeat([]byte{0x22}, maxAssociatedDataLength))
	if _, err := decodeMessageA(msgA); err != nil {
		t.Fatalf("decode A max-size fields: %v", err)
	}
	casesA := []struct {
		name string
		msg  []byte
	}{
		{"sid oversized", encodeMessageA(bytes.Repeat([]byte{0x11}, maxSessionIDLength+1), point, nil)},
		{"point oversized", encodeMessageA([]byte("sid"), bytes.Repeat([]byte{0x42}, pointSize+1), nil)},
		{"ad oversized", encodeMessageA([]byte("sid"), point, bytes.Repeat([]byte{0x22}, maxAssociatedDataLength+1))},
	}
	for _, tc := range casesA {
		t.Run("A "+tc.name, func(t *testing.T) {
			if _, err := decodeMessageA(tc.msg); !errors.Is(err, ErrMessage) {
				t.Fatalf("decode A err=%v", err)
			}
		})
	}

	msgB := encodeMessageB(point, bytes.Repeat([]byte{0x33}, maxAssociatedDataLength), tag)
	if _, err := decodeMessageB(msgB); err != nil {
		t.Fatalf("decode B max-size fields: %v", err)
	}
	casesB := []struct {
		name string
		msg  []byte
	}{
		{"point oversized", encodeMessageB(bytes.Repeat([]byte{0x42}, pointSize+1), nil, tag)},
		{"ad oversized", encodeMessageB(point, bytes.Repeat([]byte{0x33}, maxAssociatedDataLength+1), tag)},
		{"tag oversized", encodeMessageB(point, nil, bytes.Repeat([]byte{0x99}, tagSize+1))},
	}
	for _, tc := range casesB {
		t.Run("B "+tc.name, func(t *testing.T) {
			if _, err := decodeMessageB(tc.msg); !errors.Is(err, ErrMessage) {
				t.Fatalf("decode B err=%v", err)
			}
		})
	}

	if _, err := decodeMessageC(encodeMessageC(tag)); err != nil {
		t.Fatalf("decode C exact tag: %v", err)
	}
	if _, err := decodeMessageC(encodeMessageC(bytes.Repeat([]byte{0x99}, tagSize+1))); !errors.Is(err, ErrMessage) {
		t.Fatalf("decode C oversized tag err=%v", err)
	}
}

func TestMessageParsersRejectMalformedBAndC(t *testing.T) {
	msgB := encodeMessageB(bytes.Repeat([]byte{0x42}, pointSize), []byte("ADb"), bytes.Repeat([]byte{0x99}, tagSize))
	oversizedBAd := []byte{wireFormatV1, wireSuite, roleB}
	oversizedBAd = append(oversizedBAd, prependLen(bytes.Repeat([]byte{0x42}, pointSize))...)
	oversizedBAd = append(oversizedBAd, encodeLEB128(maxAssociatedDataLength+1)...)
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
		{"oversized ad", oversizedBAd},
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
		{"oversized tag", encodeMessageC(bytes.Repeat([]byte{0x99}, tagSize+1))},
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
	respCfg := testConfig()
	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := respondTestResponder(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}

	const finishers = 2
	var wg sync.WaitGroup
	errs := make(chan error, finishers)
	for range finishers {
		wg.Go(func() {
			_, _, err := initiator.Finish(msgB)
			errs <- err
		})
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
	initiator2, _, err := startTestInitiator(initCfg)
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
	invalid := mustLoadDraftInvalidVector(t)
	badA := encodeMessageA([]byte("sid"), invalid.InvalidY1, nil)
	if _, _, err := respondTestResponder(cfg, badA); !errors.Is(err, ErrAbort) {
		t.Fatalf("Respond err=%v", err)
	}
}

func TestResponderPrevalidatesInvalidInitiatorShareBeforeRandomness(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	cases := []struct {
		name string
		ya   []byte
	}{
		{"non-canonical", invalid.InvalidY1},
		{"identity", invalid.InvalidY2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			random := &countingFailingReader{err: io.ErrUnexpectedEOF}
			badA := encodeMessageA([]byte("sid"), tc.ya, nil)
			_, _, err := respondWithRandom(cfg, badA, random)
			if !errors.Is(err, ErrAbort) || errors.Is(err, ErrRandomness) {
				t.Fatalf("Respond err=%v", err)
			}
			if random.reads != 0 {
				t.Fatalf("Respond read randomness %d times before rejecting share", random.reads)
			}
		})
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
			initiator, _, err := startTestInitiator(cfg)
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
	initiator, msgA, err := startTestInitiator(cfg)
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
	respCfg := testConfig()
	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := respondTestResponder(respCfg, msgA)
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
	respCfg := testConfig()
	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := respondTestResponder(respCfg, msgA)
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
	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	responder, msgB, err := respondTestResponder(respCfg, msgA)
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

func allZero(in []byte) bool {
	for _, b := range in {
		if b != 0 {
			return false
		}
	}
	return true
}
