package cpace

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gtank/ristretto255"
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
	initCfg, respCfg := defaultExchangeInputs()

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
	initCfg, respCfg := defaultExchangeInputs()

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

func TestExportLengthBoundaries(t *testing.T) {
	sI, _ := completeExchange(t, testInitiatorInput(), testResponderInput())

	cases := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"negative", -1, true},
		{"zero", 0, false},
		{"one", 1, false},
		{"max minus one", maxHKDFOutput - 1, false},
		{"max", maxHKDFOutput, false},
		{"over max", maxHKDFOutput + 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := sI.Export([]byte("label"), []byte("ctx"), tc.length)
			if tc.wantErr {
				if err == nil {
					t.Fatal("Export succeeded, want error")
				}
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("Export err=%v want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Export err=%v", err)
			}
			if len(out) != tc.length {
				t.Fatalf("Export len=%d want %d", len(out), tc.length)
			}
		})
	}
}

func TestSessionPeerMetadata(t *testing.T) {
	initCfg, respCfg := defaultExchangeInputs()

	sI, sR := completeExchange(t, initCfg, respCfg)
	if got := sI.PeerAssociatedData(); !bytes.Equal(got, respCfg.LocalAssociatedData) {
		t.Fatalf("initiator peer AD=%q want %q", got, respCfg.LocalAssociatedData)
	}
	if got := sR.PeerAssociatedData(); !bytes.Equal(got, initCfg.LocalAssociatedData) {
		t.Fatalf("responder peer AD=%q want %q", got, initCfg.LocalAssociatedData)
	}
	if got := sI.PeerID(); !bytes.Equal(got, initCfg.PeerID) {
		t.Fatalf("initiator peer ID=%q want %q", got, initCfg.PeerID)
	}
	if got := sR.PeerID(); !bytes.Equal(got, respCfg.PeerID) {
		t.Fatalf("responder peer ID=%q want %q", got, respCfg.PeerID)
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

	emptySI, emptySR := completeExchange(t, testInitiatorInput(), testResponderInput())
	if got := emptySI.PeerAssociatedData(); len(got) != 0 {
		t.Fatalf("initiator empty peer AD=%q want empty", got)
	}
	if got := emptySR.PeerAssociatedData(); len(got) != 0 {
		t.Fatalf("responder empty peer AD=%q want empty", got)
	}
}

func TestSessionClose(t *testing.T) {
	initCfg, respCfg := defaultExchangeInputs()

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
	sI, _ := completeExchange(t, testInitiatorInput(), testResponderInput())
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

func TestClearNilSafe(t *testing.T) {
	var initiator *initiatorCore
	initiator.clear()

	var responder *responderCore
	responder.clear()
}

func TestClearIdempotent(t *testing.T) {
	scalar, err := sampleScalar(repeatingRand(0x11))
	if err != nil {
		t.Fatal(err)
	}
	initiator := &initiatorCore{scalar: scalar}
	initiator.clear()
	initiator.clear()
	if initiator.scalar != nil {
		t.Fatal("initiator core retained scalar reference after clear")
	}
	if !allZero(scalar.Bytes()) {
		t.Fatal("initiator scalar was not zeroed by clear")
	}

	isk := bytes.Repeat([]byte{0x42}, 64)
	tr := newIRTranscript([]byte("ya"), []byte("ada"), []byte("yb"), []byte("adb"))
	transcriptBody := tr.transcript // alias backing array before clear
	responder := &responderCore{
		isk:        isk,
		transcript: tr,
	}
	responder.clear()
	responder.clear()
	if responder.isk != nil {
		t.Fatal("responder core retained isk reference after clear")
	}
	if !allZero(isk) {
		t.Fatal("responder ISK was not zeroed by clear")
	}
	if !allZero(transcriptBody) {
		t.Fatal("responder transcript was not zeroed by clear")
	}
	if responder.transcript.bytes() != nil {
		t.Fatal("responder core retained transcript bytes after clear")
	}
}

func TestNilReceiverMethods(t *testing.T) {
	var s *Session
	if err := s.Close(); err != nil {
		t.Fatalf("nil Close err=%v want nil", err)
	}
	if _, err := s.Export([]byte("label"), nil, 32); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil Export err=%v want ErrInvalidInput", err)
	}
	if got := s.TranscriptID(); got != nil {
		t.Fatalf("nil TranscriptID=%q want nil", got)
	}
	if got := s.PeerAssociatedData(); got != nil {
		t.Fatalf("nil PeerAssociatedData=%q want nil", got)
	}
	if got := s.PeerID(); got != nil {
		t.Fatalf("nil PeerID=%q want nil", got)
	}

	zero := &Session{}
	if err := zero.Close(); !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "nil session") {
		t.Fatalf("zero-value Close err=%v want ErrInvalidInput", err)
	}
	if _, err := zero.Export([]byte("label"), nil, 32); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("zero-value Export err=%v want ErrInvalidInput", err)
	}

	if err := new(Session).Close(); !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "nil session") {
		t.Fatalf("new Session Close err=%v want ErrInvalidInput", err)
	}
}

func TestSessionCloseConcurrentExport(t *testing.T) {
	sI, _ := completeExchange(t, testInitiatorInput(), testResponderInput())
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
	sI, _ := completeExchange(t, testInitiatorInput(), testResponderInput())
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
	initCfg, respCfg := defaultExchangeInputs()
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
	initCfg, respCfg := defaultExchangeInputs()
	password := []byte("password")
	initiatorSelfID := []byte("initiator")
	initiatorPeerID := []byte("responder")
	initiatorContext := []byte("context")
	initiatorSessionID := []byte("sid")
	initCfg.Password = password
	initCfg.SelfID = initiatorSelfID
	initCfg.PeerID = initiatorPeerID
	initCfg.Context = initiatorContext
	initCfg.SessionID = initiatorSessionID
	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := range password {
		password[i] ^= 0xff
	}
	for i := range initiatorSelfID {
		initiatorSelfID[i] ^= 0xff
	}
	for i := range initCfg.LocalAssociatedData {
		initCfg.LocalAssociatedData[i] ^= 0xff
	}
	for i := range initiatorPeerID {
		initiatorPeerID[i] ^= 0xff
	}
	for i := range initiatorContext {
		initiatorContext[i] ^= 0xff
	}
	for i := range initiatorSessionID {
		initiatorSessionID[i] ^= 0xff
	}

	responderSelfID := []byte("responder")
	responderPeerID := []byte("initiator")
	responderContext := []byte("context")
	responderSessionID := []byte("sid")
	respCfg.SelfID = responderSelfID
	respCfg.PeerID = responderPeerID
	respCfg.Context = responderContext
	respCfg.SessionID = responderSessionID
	responder, msgB, err := respondTestResponder(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	for i := range responderSelfID {
		responderSelfID[i] ^= 0xff
	}
	for i := range responderPeerID {
		responderPeerID[i] ^= 0xff
	}
	for i := range responderContext {
		responderContext[i] ^= 0xff
	}
	for i := range responderSessionID {
		responderSessionID[i] ^= 0xff
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
	if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
		t.Fatal("transcript IDs differ after caller mutation")
	}
}

func TestInputErrorPathsDoNotMutateCallerSlices(t *testing.T) {
	cases := []struct {
		name  string
		input Input
		run   func(Input) error
		want  error
	}{
		{
			name:  "Start randomness error",
			input: testInitiatorInput(),
			run: func(input Input) error {
				_, _, err := startWithRandom(input, failingReader{err: io.ErrUnexpectedEOF})
				return err
			},
			want: ErrRandomness,
		},
		{
			name:  "Respond malformed message",
			input: testResponderInput(),
			run: func(input Input) error {
				_, _, err := Respond(input, []byte("garbage"))
				return err
			},
			want: ErrMessage,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.input
			input.LocalAssociatedData = []byte("AD")
			original := cloneInputBytes(input)
			err := tc.run(input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
			assertInputBytesEqual(t, input, original)
		})
	}
}

func cloneInputBytes(input Input) Input {
	input.Password = clone(input.Password)
	input.SelfID = clone(input.SelfID)
	input.PeerID = clone(input.PeerID)
	input.Context = clone(input.Context)
	input.SessionID = clone(input.SessionID)
	input.LocalAssociatedData = clone(input.LocalAssociatedData)
	return input
}

func assertInputBytesEqual(t *testing.T, got, want Input) {
	t.Helper()
	if !bytes.Equal(got.Password, want.Password) {
		t.Fatalf("Password mutated: got %q want %q", got.Password, want.Password)
	}
	if !bytes.Equal(got.SelfID, want.SelfID) {
		t.Fatalf("SelfID mutated: got %q want %q", got.SelfID, want.SelfID)
	}
	if !bytes.Equal(got.PeerID, want.PeerID) {
		t.Fatalf("PeerID mutated: got %q want %q", got.PeerID, want.PeerID)
	}
	if !bytes.Equal(got.Context, want.Context) {
		t.Fatalf("Context mutated: got %q want %q", got.Context, want.Context)
	}
	if !bytes.Equal(got.SessionID, want.SessionID) {
		t.Fatalf("SessionID mutated: got %q want %q", got.SessionID, want.SessionID)
	}
	if !bytes.Equal(got.LocalAssociatedData, want.LocalAssociatedData) {
		t.Fatalf("LocalAssociatedData mutated: got %q want %q", got.LocalAssociatedData, want.LocalAssociatedData)
	}
}

func TestFinishCleanupDoesNotAliasReturnedSessions(t *testing.T) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(t, initInput, respInput)
	initiatorScalar := exchange.initiator.state.core.scalar
	responderISK := exchange.responder.state.core.isk
	responderTranscript := exchange.responder.state.core.transcript.transcript

	msgC, sI := exchange.finishInitiator()
	if exchange.initiator.state.core.scalar != nil {
		t.Fatal("initiator scalar reference retained after Finish")
	}
	if initiatorScalar == nil || !allZero(initiatorScalar.Bytes()) {
		t.Fatal("consumed initiator scalar was not cleared")
	}

	sR := exchange.finishResponder(msgC)
	if exchange.responder.state.core.isk != nil || exchange.responder.state.core.transcript.bytes() != nil {
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

func TestClearOnFinishFailurePaths(t *testing.T) {
	t.Run("initiator parse failure", func(t *testing.T) {
		initiator, _, err := startTestInitiator(testInitiatorInput())
		if err != nil {
			t.Fatal(err)
		}
		initiatorScalar := initiator.state.core.scalar
		if _, _, err := initiator.Finish([]byte("garbage")); !errors.Is(err, ErrMessage) {
			t.Fatalf("initiator Finish garbage err=%v", err)
		}
		if initiator.state.core.scalar != nil {
			t.Fatal("initiator retained scalar reference after parse failure")
		}
		if initiatorScalar == nil || !allZero(initiatorScalar.Bytes()) {
			t.Fatal("initiator scalar was not cleared after parse failure")
		}
	})

	t.Run("initiator confirmation failure", func(t *testing.T) {
		initCfg := testInitiatorInput()
		initCfg.Password = []byte("password-a")
		respCfg := testResponderInput()
		respCfg.Password = []byte("password-b")
		exchange := newExchange(t, initCfg, respCfg)
		initiatorScalar := exchange.initiator.state.core.scalar
		if _, _, err := exchange.initiator.Finish(exchange.msgB); !errors.Is(err, ErrConfirmationFailed) {
			t.Fatalf("initiator Finish wrong-password err=%v", err)
		}
		if exchange.initiator.state.core.scalar != nil {
			t.Fatal("initiator retained scalar reference after confirmation failure")
		}
		if initiatorScalar == nil || !allZero(initiatorScalar.Bytes()) {
			t.Fatal("initiator scalar was not cleared after confirmation failure")
		}
	})

	t.Run("responder parse failure", func(t *testing.T) {
		initInput, respInput := defaultExchangeInputs()
		exchange := newExchange(t, initInput, respInput)
		responderISK := exchange.responder.state.core.isk
		responderTranscript := exchange.responder.state.core.transcript.transcript
		if _, err := exchange.responder.Finish([]byte("garbage")); !errors.Is(err, ErrMessage) {
			t.Fatalf("responder Finish garbage err=%v", err)
		}
		if exchange.responder.state.core.isk != nil || exchange.responder.state.core.transcript.bytes() != nil {
			t.Fatal("responder retained cleared state references after parse failure")
		}
		if !allZero(responderISK) {
			t.Fatal("responder ISK was not cleared after parse failure")
		}
		if !allZero(responderTranscript) {
			t.Fatal("responder transcript was not cleared after parse failure")
		}
	})

	t.Run("responder confirmation failure", func(t *testing.T) {
		initInput, respInput := defaultExchangeInputs()
		exchange := newExchange(t, initInput, respInput)
		msgC, _ := exchange.finishInitiator()
		msgC[len(msgC)-1] ^= 0xff
		responderISK := exchange.responder.state.core.isk
		responderTranscript := exchange.responder.state.core.transcript.transcript
		if _, err := exchange.responder.Finish(msgC); !errors.Is(err, ErrConfirmationFailed) {
			t.Fatalf("responder Finish tampered tagA err=%v", err)
		}
		if exchange.responder.state.core.isk != nil || exchange.responder.state.core.transcript.bytes() != nil {
			t.Fatal("responder retained cleared state references after confirmation failure")
		}
		if !allZero(responderISK) {
			t.Fatal("responder ISK was not cleared after confirmation failure")
		}
		if !allZero(responderTranscript) {
			t.Fatal("responder transcript was not cleared after confirmation failure")
		}
	})
}

func TestSessionISKSurvivesCoreClear(t *testing.T) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(t, initInput, respInput)
	msgC, sI := exchange.finishInitiator()
	responderISK := exchange.responder.state.core.isk
	sR := exchange.finishResponder(msgC)
	if exchange.responder.state.core.isk != nil {
		t.Fatal("responder retained ISK reference after Finish")
	}
	if !allZero(responderISK) {
		t.Fatal("responder-owned ISK backing array was not cleared")
	}
	kI, err := sI.Export([]byte("label"), []byte("ctx"), 32)
	if err != nil {
		t.Fatal(err)
	}
	kR, err := sR.Export([]byte("label"), []byte("ctx"), 32)
	if err != nil {
		t.Fatal(err)
	}
	if allZero(kR) {
		t.Fatal("responder session export unexpectedly all-zero after core cleanup")
	}
	if !bytes.Equal(kI, kR) {
		t.Fatal("exports differ after responder core cleanup")
	}
}

func TestFinishZeroValueHardening(t *testing.T) {
	var initiator Initiator
	if _, _, err := initiator.Finish([]byte("garbage")); !errors.Is(err, ErrInvalidInput) ||
		!strings.Contains(err.Error(), "uninitialized initiator") {
		t.Fatalf("zero-value Initiator.Finish malformed err=%v", err)
	}
	if initiator.state != nil {
		t.Fatal("zero-value Initiator.Finish consumed state on malformed message")
	}

	v, err := loadDraftVectorJSON(draft21RistrettoVectorJSON)
	if err != nil {
		t.Fatal(err)
	}
	msgB := encodeMessageB(v["Yb"], v["ADb"], bytes.Repeat([]byte{0x99}, tagSize))
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrInvalidInput) ||
		!strings.Contains(err.Error(), "uninitialized initiator") {
		t.Fatalf("zero-value Initiator.Finish shaped msgB err=%v", err)
	}
	if initiator.state != nil {
		t.Fatal("zero-value Initiator.Finish consumed state on shaped message B")
	}

	var responder Responder
	if _, err := responder.Finish([]byte("garbage")); !errors.Is(err, ErrInvalidInput) ||
		!strings.Contains(err.Error(), "uninitialized responder") {
		t.Fatalf("zero-value Responder.Finish malformed err=%v", err)
	}
	if responder.state != nil {
		t.Fatal("zero-value Responder.Finish consumed state on malformed message")
	}

	forgedC := encodeMessageC(confirmationTag(nil, nil, nil, nil))
	sess, err := responder.Finish(forgedC)
	if sess != nil {
		t.Fatal("zero-value Responder.Finish returned a Session for forged message C")
	}
	if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "uninitialized responder") {
		t.Fatalf("zero-value Responder.Finish forged msgC err=%v", err)
	}
	if responder.state != nil {
		t.Fatal("zero-value Responder.Finish consumed state on forged message C")
	}
}

func TestSingleUseStateCloseNilAndZeroValue(t *testing.T) {
	var nilInitiator *Initiator
	if err := nilInitiator.Close(); err != nil {
		t.Fatalf("nil Initiator.Close err=%v want nil", err)
	}
	var nilResponder *Responder
	if err := nilResponder.Close(); err != nil {
		t.Fatalf("nil Responder.Close err=%v want nil", err)
	}

	var initiator Initiator
	if err := initiator.Close(); !errors.Is(err, ErrInvalidInput) ||
		!strings.Contains(err.Error(), "uninitialized initiator") {
		t.Fatalf("zero-value Initiator.Close err=%v want ErrInvalidInput", err)
	}
	if initiator.state != nil {
		t.Fatal("zero-value Initiator.Close consumed state")
	}

	var responder Responder
	if err := responder.Close(); !errors.Is(err, ErrInvalidInput) ||
		!strings.Contains(err.Error(), "uninitialized responder") {
		t.Fatalf("zero-value Responder.Close err=%v want ErrInvalidInput", err)
	}
	if responder.state != nil {
		t.Fatal("zero-value Responder.Close consumed state")
	}
}

func TestSingleUseStateCloseCleansAbandonedState(t *testing.T) {
	initiator, msgA, err := startTestInitiator(testInitiatorInput())
	if err != nil {
		t.Fatal(err)
	}
	initiatorScalar := initiator.state.core.scalar
	if err := initiator.Close(); err != nil {
		t.Fatalf("Initiator.Close err=%v", err)
	}
	if initiator.state.core.scalar != nil {
		t.Fatal("initiator core retained scalar reference after Close")
	}
	if initiatorScalar == nil || !allZero(initiatorScalar.Bytes()) {
		t.Fatal("initiator scalar was not cleared by Close")
	}
	if err := initiator.Close(); err != nil {
		t.Fatalf("second Initiator.Close err=%v", err)
	}
	if _, _, err := initiator.Finish([]byte("garbage")); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("Initiator.Finish after Close err=%v want ErrStateUsed", err)
	}

	responder, _, err := respondTestResponder(testResponderInput(), msgA)
	if err != nil {
		t.Fatal(err)
	}
	responderISK := responder.state.core.isk
	responderTranscript := responder.state.core.transcript.transcript
	if err := responder.Close(); err != nil {
		t.Fatalf("Responder.Close err=%v", err)
	}
	if responder.state.core.isk != nil || responder.state.core.transcript.bytes() != nil {
		t.Fatal("responder core retained cleared state references after Close")
	}
	if !allZero(responderISK) {
		t.Fatal("responder ISK was not cleared by Close")
	}
	if !allZero(responderTranscript) {
		t.Fatal("responder transcript was not cleared by Close")
	}
	if err := responder.Close(); err != nil {
		t.Fatalf("second Responder.Close err=%v", err)
	}
	if _, err := responder.Finish([]byte("garbage")); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("Responder.Finish after Close err=%v want ErrStateUsed", err)
	}
}

func TestSingleUseStateCloseAfterFinish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		initInput, respInput := defaultExchangeInputs()
		exchange := newExchange(t, initInput, respInput)
		msgC, initSession := exchange.finishInitiator()
		if err := exchange.initiator.Close(); err != nil {
			t.Fatalf("Initiator.Close after successful Finish err=%v", err)
		}
		if _, err := initSession.Export([]byte("label"), []byte("ctx"), 32); err != nil {
			t.Fatalf("initiator Session.Export after Close-on-state err=%v", err)
		}
		respSession := exchange.finishResponder(msgC)
		if err := exchange.responder.Close(); err != nil {
			t.Fatalf("Responder.Close after successful Finish err=%v", err)
		}
		if _, err := respSession.Export([]byte("label"), []byte("ctx"), 32); err != nil {
			t.Fatalf("responder Session.Export after Close-on-state err=%v", err)
		}
	})

	t.Run("failed finish", func(t *testing.T) {
		initInput, respInput := defaultExchangeInputs()
		exchange := newExchange(t, initInput, respInput)
		if _, _, err := exchange.initiator.Finish([]byte("garbage")); !errors.Is(err, ErrMessage) {
			t.Fatalf("initiator Finish garbage err=%v", err)
		}
		if err := exchange.initiator.Close(); err != nil {
			t.Fatalf("Initiator.Close after failed Finish err=%v", err)
		}

		if _, err := exchange.responder.Finish([]byte("garbage")); !errors.Is(err, ErrMessage) {
			t.Fatalf("responder Finish garbage err=%v", err)
		}
		if err := exchange.responder.Close(); err != nil {
			t.Fatalf("Responder.Close after failed Finish err=%v", err)
		}

		exchange2 := newExchange(t, initInput, respInput)
		msgC2, _ := exchange2.finishInitiator()
		msgC2[len(msgC2)-1] ^= 0xff
		if _, err := exchange2.responder.Finish(msgC2); !errors.Is(err, ErrConfirmationFailed) {
			t.Fatalf("responder Finish tampered tagA err=%v", err)
		}
		if err := exchange2.responder.Close(); err != nil {
			t.Fatalf("Responder.Close after confirmation failure err=%v", err)
		}
	})
}

func TestSingleUseStateCopiesShareTerminalState(t *testing.T) {
	initiator, msgA, err := startTestInitiator(testInitiatorInput())
	if err != nil {
		t.Fatal(err)
	}
	initiatorCopy := *initiator
	if err := initiator.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := initiatorCopy.Finish([]byte("garbage")); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("copied Initiator.Finish after original Close err=%v want ErrStateUsed", err)
	}
	if err := initiatorCopy.Close(); err != nil {
		t.Fatalf("copied Initiator.Close after original Close err=%v", err)
	}

	responder, _, err := respondTestResponder(testResponderInput(), msgA)
	if err != nil {
		t.Fatal(err)
	}
	responderCopy := *responder
	if err := responderCopy.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := responder.Finish([]byte("garbage")); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("original Responder.Finish after copied Close err=%v want ErrStateUsed", err)
	}
	if err := responder.Close(); err != nil {
		t.Fatalf("original Responder.Close after copied Close err=%v", err)
	}
}

func TestSingleUseTerminalClaimsDoNotReturnCoreOnLosingPaths(t *testing.T) {
	initCore := &initiatorCore{}
	assertLosingTerminalClaimDoesNotReturnCore(t, &initiatorState{used: true, core: initCore}, initCore)

	respCore := &responderCore{}
	assertLosingTerminalClaimDoesNotReturnCore(t, &responderState{used: true, core: respCore}, respCore)
}

func TestSingleUseTerminalNilCoreReturnsUninitializedDiagnostic(t *testing.T) {
	assertNilCoreTerminalClaimReturnsUninitialized(t, "initiator", "uninitialized initiator", func() *initiatorState {
		return &initiatorState{uninitialized: "uninitialized initiator"}
	})
	assertNilCoreTerminalClaimReturnsUninitialized(t, "responder", "uninitialized responder", func() *responderState {
		return &responderState{uninitialized: "uninitialized responder"}
	})
}

func assertLosingTerminalClaimDoesNotReturnCore[C singleUseCore](t *testing.T, state *singleUseState[C], core C) {
	t.Helper()
	if got, err := state.claimClose(); err != nil || got != nil {
		t.Fatalf("losing close got core=%v err=%v, want nil nil", got, err)
	}
	if got, err := state.claimFinish(); !errors.Is(err, ErrStateUsed) || got != nil {
		t.Fatalf("losing finish got core=%v err=%v, want nil ErrStateUsed", got, err)
	}
	if core == nil {
		t.Fatal("test must provide a non-nil core")
	}
	if state.core != core {
		t.Fatal("losing claims mutated the stored core pointer")
	}
}

func assertNilCoreTerminalClaimReturnsUninitialized[C singleUseCore](t *testing.T, role, want string, newState func() *singleUseState[C]) {
	t.Helper()
	t.Run(role+" finish", func(t *testing.T) {
		state := newState()
		got, err := state.claimFinish()
		if got != nil {
			t.Fatalf("nil-core finish got core=%v want nil", got)
		}
		if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), want) {
			t.Fatalf("nil-core finish err=%v want ErrInvalidInput containing %q", err, want)
		}
		if state.used {
			t.Fatal("nil-core finish consumed terminal state")
		}
	})
	t.Run(role+" close", func(t *testing.T) {
		state := newState()
		got, err := state.claimClose()
		if got != nil {
			t.Fatalf("nil-core close got core=%v want nil", got)
		}
		if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), want) {
			t.Fatalf("nil-core close err=%v want ErrInvalidInput containing %q", err, want)
		}
		if state.used {
			t.Fatal("nil-core close consumed terminal state")
		}
	})
}

func TestInputValidation(t *testing.T) {
	cases := []struct {
		name       string
		edit       func(*Input)
		want       string
		wantErrors []error
	}{
		{
			name:       "password",
			edit:       func(c *Input) { c.Password = nil },
			want:       "cpace: invalid input: empty password",
			wantErrors: []error{ErrInvalidInput},
		},
		{
			name:       "self id",
			edit:       func(c *Input) { c.SelfID = nil },
			want:       "cpace: invalid input: empty self id",
			wantErrors: []error{ErrInvalidInput},
		},
		{
			name:       "peer id",
			edit:       func(c *Input) { c.PeerID = nil },
			want:       "cpace: invalid input: empty peer id",
			wantErrors: []error{ErrInvalidInput},
		},
		{
			name:       "session id",
			edit:       func(c *Input) { c.SessionID = nil },
			want:       "cpace: invalid input: cpace: empty session id",
			wantErrors: []error{ErrInvalidInput, ErrEmptySessionID},
		},
		{
			name: "self id before oversized context",
			edit: func(c *Input) {
				c.SelfID = nil
				c.Context = bytes.Repeat([]byte{0x42}, contextCap.length+1)
			},
			want:       "cpace: invalid input: empty self id",
			wantErrors: []error{ErrInvalidInput},
		},
		{
			name: "session id before oversized context",
			edit: func(c *Input) {
				c.SessionID = nil
				c.Context = bytes.Repeat([]byte{0x42}, contextCap.length+1)
			},
			want:       "cpace: invalid input: cpace: empty session id",
			wantErrors: []error{ErrInvalidInput, ErrEmptySessionID},
		},
		{
			name: "oversized context before oversized session id",
			edit: func(c *Input) {
				c.Context = bytes.Repeat([]byte{0x42}, contextCap.length+1)
				c.SessionID = bytes.Repeat([]byte{0x42}, sessionIDCap.length+1)
			},
			want:       "cpace: invalid input: context too large",
			wantErrors: []error{ErrInvalidInput},
		},
		{
			name: "oversized context before oversized local associated data",
			edit: func(c *Input) {
				c.Context = bytes.Repeat([]byte{0x42}, contextCap.length+1)
				c.LocalAssociatedData = bytes.Repeat([]byte{0x42}, localAssociatedDataCap.length+1)
			},
			want:       "cpace: invalid input: context too large",
			wantErrors: []error{ErrInvalidInput},
		},
		{
			name: "oversized session id before oversized local associated data",
			edit: func(c *Input) {
				c.SessionID = bytes.Repeat([]byte{0x42}, sessionIDCap.length+1)
				c.LocalAssociatedData = bytes.Repeat([]byte{0x42}, localAssociatedDataCap.length+1)
			},
			want:       "cpace: invalid input: session id too large",
			wantErrors: []error{ErrInvalidInput},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testInitiatorInput()
			tc.edit(&cfg)
			assertInputValidationError(t, cfg, tc.want, tc.wantErrors...)
		})
	}
}

func assertInputValidationError(t *testing.T, cfg Input, want string, wantErrors ...error) {
	t.Helper()
	calls := []struct {
		name string
		run  func(Input) error
	}{
		{
			name: "Start",
			run: func(input Input) error {
				_, _, err := Start(input)
				return err
			},
		},
		{
			name: "Respond",
			run: func(input Input) error {
				_, _, err := Respond(input, nil)
				return err
			},
		},
	}
	for _, call := range calls {
		t.Run(call.name, func(t *testing.T) {
			err := call.run(cfg)
			if err == nil {
				t.Fatal("err=nil")
			}
			for _, wantErr := range wantErrors {
				if !errors.Is(err, wantErr) {
					t.Fatalf("err=%v want errors.Is(..., %v)", err, wantErr)
				}
			}
			if err.Error() != want {
				t.Fatalf("err=%q want %q", err.Error(), want)
			}
		})
	}
}

func TestInputFieldSizeLimits(t *testing.T) {
	cases := []struct {
		field packageCapField
		edit  func(*Input, []byte)
	}{
		{passwordCap, func(c *Input, b []byte) { c.Password = b }},
		{selfIDCap, func(c *Input, b []byte) { c.SelfID = b }},
		{peerIDCap, func(c *Input, b []byte) { c.PeerID = b }},
		{contextCap, func(c *Input, b []byte) { c.Context = b }},
		{sessionIDCap, func(c *Input, b []byte) { c.SessionID = b }},
		{localAssociatedDataCap, func(c *Input, b []byte) { c.LocalAssociatedData = b }},
	}
	for _, tc := range cases {
		t.Run(tc.field.name+" max", func(t *testing.T) {
			cfg := testInitiatorInput()
			tc.edit(&cfg, bytes.Repeat([]byte{0x42}, tc.field.length))
			if _, _, err := startTestInitiator(cfg); err != nil {
				t.Fatalf("Start rejected max-size field: %v", err)
			}
		})
		t.Run(tc.field.name+" oversized", func(t *testing.T) {
			cfg := testInitiatorInput()
			tc.edit(&cfg, bytes.Repeat([]byte{0x42}, tc.field.length+1))
			if _, _, err := startTestInitiator(cfg); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Start err=%v", err)
			} else if want := ErrInvalidInput.Error() + ": " + tc.field.name + " too large"; err.Error() != want {
				t.Fatalf("Start err=%q want %q", err.Error(), want)
			}
			if _, _, err := Respond(cfg, nil); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Respond err=%v", err)
			} else if want := ErrInvalidInput.Error() + ": " + tc.field.name + " too large"; err.Error() != want {
				t.Fatalf("Respond err=%q want %q", err.Error(), want)
			}
		})
	}
}

func TestProtocolAllowsEmptyLocalAssociatedData(t *testing.T) {
	cases := []struct {
		name   string
		initAD []byte
		respAD []byte
	}{
		{"nil nil", nil, nil},
		{"empty empty", []byte{}, []byte{}},
		{"nil empty", nil, []byte{}},
		{"empty nil", []byte{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			initCfg := testInitiatorInput()
			initCfg.LocalAssociatedData = tc.initAD
			respCfg := testResponderInput()
			respCfg.LocalAssociatedData = tc.respAD
			sI, sR := completeExchange(t, initCfg, respCfg)
			if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
				t.Fatal("transcript IDs differ")
			}
			if got := sI.PeerAssociatedData(); len(got) != 0 {
				t.Fatalf("initiator peer associated data=%q want empty", got)
			}
			if got := sR.PeerAssociatedData(); len(got) != 0 {
				t.Fatalf("responder peer associated data=%q want empty", got)
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
			cfg := testInitiatorInput()
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
			initCfg, respCfg := defaultExchangeInputs()
			initCfg.SessionID = tc.initSID
			initCfg.AllowEmptySessionID = true
			respCfg.SessionID = tc.respSID
			respCfg.AllowEmptySessionID = true
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
			initCfg := testInitiatorInput()
			initCfg.SessionID = tc.initSID
			initCfg.AllowEmptySessionID = tc.allowInitEmpty
			respCfg := testResponderInput()
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
			initCfg := testInitiatorInput()
			initCfg.AllowEmptySessionID = tc.allowInitEmpty
			respCfg := testResponderInput()
			respCfg.AllowEmptySessionID = tc.allowRespEmpty
			sI, sR := completeExchange(t, initCfg, respCfg)
			if !bytes.Equal(sI.TranscriptID(), sR.TranscriptID()) {
				t.Fatalf("transcript IDs differ")
			}
		})
	}
}

func TestConfirmationFailsOnBoundInputMismatch(t *testing.T) {
	initCfg := testInitiatorInput()
	respCfg := testResponderInput()
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

func TestRoleLocalIdentityReversalFailsConfirmation(t *testing.T) {
	initCfg := testInitiatorInput()
	respCfg := testInitiatorInput()

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
		editResp   func(*Input)
		tamperA    func([]byte) []byte
		tamperB    func([]byte) []byte
		respondErr error
		finishErr  error
	}{
		{
			name:      "initiator identity",
			editResp:  func(c *Input) { c.SelfID = []byte("other initiator") },
			finishErr: ErrConfirmationFailed,
		},
		{
			name:      "responder identity",
			editResp:  func(c *Input) { c.PeerID = []byte("other responder") },
			finishErr: ErrConfirmationFailed,
		},
		{
			name:      "context",
			editResp:  func(c *Input) { c.Context = []byte("other context") },
			finishErr: ErrConfirmationFailed,
		},
		{
			name:       "session id",
			editResp:   func(c *Input) { c.SessionID = []byte("other sid") },
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
			initCfg, respCfg := defaultExchangeInputs()
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

func TestStateReuseAndConcurrentFinish(t *testing.T) {
	initCfg := testInitiatorInput()
	respCfg := testResponderInput()
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

func TestSingleUseStateConcurrentFinishClose(t *testing.T) {
	t.Run("initiator original finish versus copied close", func(t *testing.T) {
		initiator, msgA, err := startTestInitiator(testInitiatorInput())
		if err != nil {
			t.Fatal(err)
		}
		_, msgB, err := respondTestResponder(testResponderInput(), msgA)
		if err != nil {
			t.Fatal(err)
		}
		initiatorCopy := *initiator

		var wg sync.WaitGroup
		finishErrs := make(chan error, 1)
		closeErrs := make(chan error, 1)
		sessions := make(chan *Session, 1)
		wg.Go(func() {
			_, sess, err := initiator.Finish(msgB)
			finishErrs <- err
			sessions <- sess
		})
		wg.Go(func() {
			closeErrs <- initiatorCopy.Close()
		})
		wg.Wait()
		close(finishErrs)
		close(closeErrs)
		close(sessions)

		if err := <-closeErrs; err != nil {
			t.Fatalf("copied Initiator.Close err=%v", err)
		}
		finishErr := <-finishErrs
		sess := <-sessions
		switch {
		case finishErr == nil:
			if sess == nil {
				t.Fatal("successful Finish returned nil Session")
			}
			if err := sess.Close(); err != nil {
				t.Fatal(err)
			}
		case errors.Is(finishErr, ErrStateUsed):
			if sess != nil {
				t.Fatal("ErrStateUsed Finish returned Session")
			}
		default:
			t.Fatalf("Finish err=%v, want nil or ErrStateUsed", finishErr)
		}
	})

	t.Run("responder copied finish versus original close", func(t *testing.T) {
		initiator, msgA, err := startTestInitiator(testInitiatorInput())
		if err != nil {
			t.Fatal(err)
		}
		responder, msgB, err := respondTestResponder(testResponderInput(), msgA)
		if err != nil {
			t.Fatal(err)
		}
		msgC, _, err := initiator.Finish(msgB)
		if err != nil {
			t.Fatal(err)
		}
		responderCopy := *responder

		var wg sync.WaitGroup
		finishErrs := make(chan error, 1)
		closeErrs := make(chan error, 1)
		sessions := make(chan *Session, 1)
		wg.Go(func() {
			sess, err := responderCopy.Finish(msgC)
			finishErrs <- err
			sessions <- sess
		})
		wg.Go(func() {
			closeErrs <- responder.Close()
		})
		wg.Wait()
		close(finishErrs)
		close(closeErrs)
		close(sessions)

		if err := <-closeErrs; err != nil {
			t.Fatalf("Responder.Close err=%v", err)
		}
		finishErr := <-finishErrs
		sess := <-sessions
		switch {
		case finishErr == nil:
			if sess == nil {
				t.Fatal("successful Finish returned nil Session")
			}
			if err := sess.Close(); err != nil {
				t.Fatal(err)
			}
		case errors.Is(finishErr, ErrStateUsed):
			if sess != nil {
				t.Fatal("ErrStateUsed Finish returned Session")
			}
		default:
			t.Fatalf("Finish err=%v, want nil or ErrStateUsed", finishErr)
		}
	})
}

func TestProtocolAbortsOnInvalidRistrettoEncoding(t *testing.T) {
	cfg := testResponderInput()
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
			cfg := testResponderInput()
			random := &countingFailingReader{err: io.ErrUnexpectedEOF}
			badA := encodeMessageA([]byte("sid"), tc.ya, nil)
			_, _, err := respondWithRandom(cfg, badA, random)
			if !errors.Is(err, ErrAbort) || errors.Is(err, ErrRandomness) {
				t.Fatalf("Respond err=%v", err)
			}
			if random.reads != 0 {
				t.Fatalf("Respond read randomness %d times before rejecting share", random.reads)
			}

			nc, err := normalizeRespondInput(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer nc.wipe()
			random = &countingFailingReader{err: io.ErrUnexpectedEOF}
			core, yb, tagB, err := newResponderCore(nc, tc.ya, nil, random)
			if core != nil || yb != nil || tagB != nil {
				t.Fatalf("newResponderCore returned core=%v yb=%x tagB=%x on invalid share", core, yb, tagB)
			}
			if !errors.Is(err, ErrAbort) || errors.Is(err, ErrRandomness) {
				t.Fatalf("newResponderCore err=%v", err)
			}
			if random.reads != 0 {
				t.Fatalf("newResponderCore read randomness %d times before rejecting share", random.reads)
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
			cfg := testInitiatorInput()
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

func TestPeerShareErrorsWrapErrAbort(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	cases := []struct {
		name      string
		share     []byte
		sentinel  error
		other     error
		wantError string
		viaFinish bool
	}{
		{
			name:      "respond non-canonical",
			share:     invalid.InvalidY1,
			sentinel:  ErrPeerShareEncoding,
			other:     ErrPeerShareIdentity,
			wantError: "cpace: protocol abort: invalid initiator share: cpace: peer share encoding",
		},
		{
			name:      "respond identity",
			share:     invalid.InvalidY2,
			sentinel:  ErrPeerShareIdentity,
			other:     ErrPeerShareEncoding,
			wantError: "cpace: protocol abort: invalid initiator share: cpace: peer share identity",
		},
		{
			name:      "finish non-canonical",
			share:     invalid.InvalidY1,
			sentinel:  ErrPeerShareEncoding,
			other:     ErrPeerShareIdentity,
			wantError: "cpace: protocol abort: invalid responder share: cpace: peer share encoding",
			viaFinish: true,
		},
		{
			name:      "finish identity",
			share:     invalid.InvalidY2,
			sentinel:  ErrPeerShareIdentity,
			other:     ErrPeerShareEncoding,
			wantError: "cpace: protocol abort: invalid responder share: cpace: peer share identity",
			viaFinish: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.viaFinish {
				initiator, _, startErr := Start(testInitiatorInput())
				if startErr != nil {
					t.Fatal(startErr)
				}
				msgB := encodeMessageB(tc.share, nil, bytes.Repeat([]byte{0x99}, tagSize))
				_, _, err = initiator.Finish(msgB)
			} else {
				badA := encodeMessageA([]byte("sid"), tc.share, nil)
				_, _, err = Respond(testInitiatorInput(), badA)
			}
			if err == nil {
				t.Fatal("expected peer-share rejection, got nil error")
			}
			if !errors.Is(err, ErrAbort) {
				t.Fatalf("err=%v does not wrap ErrAbort", err)
			}
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("err=%v does not wrap %v", err, tc.sentinel)
			}
			if errors.Is(err, tc.other) {
				t.Fatalf("err=%v wraps unrelated sentinel %v", err, tc.other)
			}
			// Exact-string match pins the role context and the single
			// "cpace: protocol abort" prefix mandated by ADR-0003.
			if err.Error() != tc.wantError {
				t.Fatalf("err=%q want %q", err.Error(), tc.wantError)
			}
		})
	}
}

func TestPeerShareEncodingRejection(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	badA := encodeMessageA([]byte("sid"), invalid.InvalidY1, nil)
	_, _, err := Respond(testInitiatorInput(), badA)
	if !errors.Is(err, ErrPeerShareEncoding) {
		t.Fatalf("Respond err=%v want ErrPeerShareEncoding", err)
	}
	if !errors.Is(err, ErrAbort) {
		t.Fatalf("Respond err=%v does not wrap ErrAbort", err)
	}
}

func TestPeerShareIdentityRejection(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	badA := encodeMessageA([]byte("sid"), invalid.InvalidY2, nil)
	_, _, err := Respond(testInitiatorInput(), badA)
	if !errors.Is(err, ErrPeerShareIdentity) {
		t.Fatalf("Respond err=%v want ErrPeerShareIdentity", err)
	}
	if !errors.Is(err, ErrAbort) {
		t.Fatalf("Respond err=%v does not wrap ErrAbort", err)
	}
}

func TestPeerShareLengthDefenseInternal(t *testing.T) {
	// Malformed wire lengths surface as ErrMessage from framing before any
	// share reaches decodePublicShare, so the length branch is only reachable
	// through a direct internal call. It must stay an ErrAbort-wrapped
	// defensive error with no peer-share sentinel.
	invalid := mustLoadDraftInvalidVector(t)
	s, err := scalarFromCanonical(invalid.Valid["s"])
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, pointSize - 1, pointSize + 1} {
		short := make([]byte, n)
		p, err := decodePublicShare(short)
		if p != nil {
			t.Fatalf("len=%d: decodePublicShare returned non-nil element", n)
		}
		assertLengthDefenseError(t, n, err)
		out, err := scalarMultVFY(s, short)
		if out != nil {
			t.Fatalf("len=%d: scalarMultVFY out=%x want nil", n, out)
		}
		assertLengthDefenseError(t, n, err)
	}
}

func assertLengthDefenseError(t *testing.T, n int, err error) {
	t.Helper()
	if !errors.Is(err, ErrAbort) {
		t.Fatalf("len=%d: err=%v does not wrap ErrAbort", n, err)
	}
	if errors.Is(err, ErrPeerShareEncoding) || errors.Is(err, ErrPeerShareIdentity) {
		t.Fatalf("len=%d: err=%v wraps a peer-share sentinel", n, err)
	}
	if !strings.Contains(err.Error(), "invalid peer share length") {
		t.Fatalf("len=%d: err=%q missing length diagnostic", n, err)
	}
}

func TestPeerShareRolePassesThroughNonSentinelErrors(t *testing.T) {
	// The ADR-0003 call-site mapping sanctions exactly two behaviors: rewrap
	// the two exported sentinels with role context, and pass every other
	// error through unchanged. Pin the pass-through half with the two real
	// non-sentinel errors the helpers can produce, asserting pointer identity
	// so both a role-context rewrap (the duplicated-prefix shape) and an
	// error-swallowing regression fail loudly.
	invalid := mustLoadDraftInvalidVector(t)
	_, lengthErr := decodePublicShare(make([]byte, pointSize-1))
	if lengthErr == nil {
		t.Fatal("expected length defense error")
	}
	_, neutralErr := scalarMultVFY(ristretto255.NewScalar().Zero(), invalid.Valid["X"])
	if neutralErr == nil {
		t.Fatal("expected post-multiply neutral-element error")
	}
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"wrong length", lengthErr},
		{"post-multiply neutral element", neutralErr},
	} {
		for _, role := range []peerShareRole{initiatorPeerShare, responderPeerShare} {
			if got := role.wrapError(tc.err); !samePointerErrorValue(got, tc.err) {
				t.Fatalf("%s/%s: wrapError returned %v, want the identical error value", tc.name, role, got)
			}
		}
	}
}

func samePointerErrorValue(got, want error) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	gotValue := reflect.ValueOf(got)
	wantValue := reflect.ValueOf(want)
	if gotValue.Type() != wantValue.Type() {
		return false
	}
	if gotValue.Kind() != reflect.Pointer {
		return false
	}
	return gotValue.Pointer() == wantValue.Pointer()
}

func TestPeerShareRoleSharedSecretAddsRoleContext(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	s, err := scalarFromCanonical(invalid.Valid["s"])
	if err != nil {
		t.Fatal(err)
	}

	got, err := responderPeerShare.sharedSecret(s, invalid.Valid["X"])
	if err != nil {
		t.Fatal(err)
	}
	if want := invalid.Valid["G.scalar_mult_vfy(s,X)"]; !bytes.Equal(got, want) {
		t.Fatalf("sharedSecret got %x want %x", got, want)
	}

	_, err = responderPeerShare.sharedSecret(s, invalid.InvalidY1)
	if err == nil {
		t.Fatal("expected responder share rejection")
	}
	if !errors.Is(err, ErrAbort) || !errors.Is(err, ErrPeerShareEncoding) {
		t.Fatalf("sharedSecret err=%v want ErrAbort and ErrPeerShareEncoding", err)
	}
	if want := "cpace: protocol abort: invalid responder share: cpace: peer share encoding"; err.Error() != want {
		t.Fatalf("sharedSecret err=%q want %q", err.Error(), want)
	}

	_, err = initiatorPeerShare.sharedSecret(s, invalid.InvalidY2)
	if err == nil {
		t.Fatal("expected initiator share rejection")
	}
	if !errors.Is(err, ErrAbort) || !errors.Is(err, ErrPeerShareIdentity) {
		t.Fatalf("sharedSecret err=%v want ErrAbort and ErrPeerShareIdentity", err)
	}
	if want := "cpace: protocol abort: invalid initiator share: cpace: peer share identity"; err.Error() != want {
		t.Fatalf("sharedSecret err=%q want %q", err.Error(), want)
	}
}

func TestScalarMultVFYElementMatchesEncodedPeerShare(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	s, err := scalarFromCanonical(invalid.Valid["s"])
	if err != nil {
		t.Fatal(err)
	}
	p, err := decodePublicShare(invalid.Valid["X"])
	if err != nil {
		t.Fatal(err)
	}
	got, err := scalarMultVFYElement(s, p)
	if err != nil {
		t.Fatal(err)
	}
	want, err := scalarMultVFY(s, invalid.Valid["X"])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("scalarMultVFYElement got %x want %x", got, want)
	}
}

func TestPeerShareRoleDecodeSharedSecretAddsRoleContext(t *testing.T) {
	invalid := mustLoadDraftInvalidVector(t)
	s, err := scalarFromCanonical(invalid.Valid["s"])
	if err != nil {
		t.Fatal(err)
	}

	p, err := responderPeerShare.decode(invalid.Valid["X"])
	if err != nil {
		t.Fatal(err)
	}
	got, err := responderPeerShare.sharedSecretElement(s, p)
	if err != nil {
		t.Fatal(err)
	}
	if want := invalid.Valid["G.scalar_mult_vfy(s,X)"]; !bytes.Equal(got, want) {
		t.Fatalf("sharedSecretElement got %x want %x", got, want)
	}

	_, err = responderPeerShare.decode(invalid.InvalidY1)
	if err == nil {
		t.Fatal("expected responder share rejection")
	}
	if !errors.Is(err, ErrAbort) || !errors.Is(err, ErrPeerShareEncoding) {
		t.Fatalf("decode err=%v want ErrAbort and ErrPeerShareEncoding", err)
	}
	if want := "cpace: protocol abort: invalid responder share: cpace: peer share encoding"; err.Error() != want {
		t.Fatalf("decode err=%q want %q", err.Error(), want)
	}

	_, err = initiatorPeerShare.decode(invalid.InvalidY2)
	if err == nil {
		t.Fatal("expected initiator share rejection")
	}
	if !errors.Is(err, ErrAbort) || !errors.Is(err, ErrPeerShareIdentity) {
		t.Fatalf("decode err=%v want ErrAbort and ErrPeerShareIdentity", err)
	}
	if want := "cpace: protocol abort: invalid initiator share: cpace: peer share identity"; err.Error() != want {
		t.Fatalf("decode err=%q want %q", err.Error(), want)
	}
}

func TestWireLengthRejectionIsMessageNotPeerShare(t *testing.T) {
	// Pins the layering claim made by ADR-0003 and docs/integration-guidance.md:
	// malformed wire lengths surface as ErrMessage from framing and never as
	// ErrAbort or a peer-share sentinel.
	for _, n := range []int{pointSize - 1, pointSize + 1} {
		badA := encodeMessageA([]byte("sid"), make([]byte, n), nil)
		_, _, err := Respond(testInitiatorInput(), badA)
		if !errors.Is(err, ErrMessage) {
			t.Fatalf("len=%d: Respond err=%v want ErrMessage", n, err)
		}
		if errors.Is(err, ErrAbort) || errors.Is(err, ErrPeerShareEncoding) || errors.Is(err, ErrPeerShareIdentity) {
			t.Fatalf("len=%d: wire-length rejection err=%v leaked an abort-layer error", n, err)
		}
	}
}

func TestScalarMultVFYPostMultiplyIdentityDefense(t *testing.T) {
	// For prime-order Ristretto255 the post-multiply identity branch is
	// unreachable from the wire: s·p is non-identity for any decoded
	// (non-identity) p and any scalar sampleScalar can return, since zero
	// samples are rejected. A zero scalar — an input sampleScalar can never
	// produce — is therefore the one direct-call input that forces s·p to
	// the identity and exercises the branch.
	invalid := mustLoadDraftInvalidVector(t)
	out, err := scalarMultVFY(ristretto255.NewScalar().Zero(), invalid.Valid["X"])
	if out != nil {
		t.Fatalf("out=%x want nil", out)
	}
	if !errors.Is(err, ErrAbort) {
		t.Fatalf("err=%v does not wrap ErrAbort", err)
	}
	if errors.Is(err, ErrPeerShareEncoding) || errors.Is(err, ErrPeerShareIdentity) {
		t.Fatalf("err=%v wraps a peer-share sentinel", err)
	}
	if !strings.Contains(err.Error(), "neutral-element shared secret") {
		t.Fatalf("err=%q missing neutral-element diagnostic", err)
	}

	p, err := decodePublicShare(invalid.Valid["X"])
	if err != nil {
		t.Fatal(err)
	}
	out, err = scalarMultVFYElement(ristretto255.NewScalar().Zero(), p)
	if out != nil {
		t.Fatalf("element out=%x want nil", out)
	}
	if !errors.Is(err, ErrAbort) {
		t.Fatalf("element err=%v does not wrap ErrAbort", err)
	}
	if errors.Is(err, ErrPeerShareEncoding) || errors.Is(err, ErrPeerShareIdentity) {
		t.Fatalf("element err=%v wraps a peer-share sentinel", err)
	}
	if !strings.Contains(err.Error(), "neutral-element shared secret") {
		t.Fatalf("element err=%q missing neutral-element diagnostic", err)
	}
}

func TestInitiatorReflectedShareFailsConfirmationNotAbort(t *testing.T) {
	cfg := testInitiatorInput()
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
	initCfg := testInitiatorInput()
	respCfg := testResponderInput()
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
	initCfg := testInitiatorInput()
	respCfg := testResponderInput()
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

func TestConfirmationFailsOnPasswordMismatch(t *testing.T) {
	initCfg := testInitiatorInput()
	initCfg.Password = []byte("password-a")
	respCfg := testResponderInput()
	respCfg.Password = []byte("password-b")

	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := respondTestResponder(respCfg, msgA); err != nil {
		t.Fatalf("Respond err=%v: Respond success does not authenticate by itself; it must succeed even with a wrong-password peer", err)
	}
	_, msgB, err := respondTestResponder(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrConfirmationFailed) {
		t.Fatalf("initiator Finish with mismatched password err=%v want ErrConfirmationFailed", err)
	}
}

func TestBuildCIWireStability(t *testing.T) {
	if got, want := DraftVersion, "draft-irtf-cfrg-cpace-21"; got != want {
		t.Fatalf("DraftVersion=%q want %q", got, want)
	}
	if got, want := suiteName, "CPACE-RISTR255-SHA512"; got != want {
		t.Fatalf("suiteName=%q want %q", got, want)
	}
	if got, want := currentSuite, byte(0x01); got != want {
		t.Fatalf("currentSuite=0x%02x want 0x%02x", got, want)
	}

	// Pin the exact byte output of buildCI for fixed inputs. Any change to
	// the contributing strings, their layout order, or the LV encoding will
	// fail this assertion. This is the primary guard against silent
	// protocol-identity drift; the keyed material derived through this CI
	// is load-bearing for every session.
	var want []byte
	appendLV := func(s []byte) {
		n := len(s)
		if n > 0x7f {
			t.Fatalf("test inputs must fit in single-byte LEB128; len=%d", n)
			return // proves the byte(n) bound to gosec G115; Fatalf's no-return is invisible to it
		}
		want = append(want, byte(n))
		want = append(want, s...)
	}
	appendLV([]byte("CPace-Go-CI"))
	appendLV([]byte("draft-irtf-cfrg-cpace-21"))
	appendLV([]byte("CPACE-RISTR255-SHA512"))
	appendLV([]byte("initiator"))
	appendLV([]byte("initiator-id"))
	appendLV([]byte("responder"))
	appendLV([]byte("responder-id"))
	appendLV([]byte("context"))
	appendLV([]byte("ctx-value"))

	got := buildCI([]byte("initiator-id"), []byte("responder-id"), []byte("ctx-value"))
	if !bytes.Equal(got, want) {
		t.Fatalf("buildCI drift\n got=%x\nwant=%x", got, want)
	}
}

func TestNilReceiverFinishAndExport(t *testing.T) {
	var i *Initiator
	if _, _, err := i.Finish([]byte("msgB")); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil Initiator.Finish err=%v want ErrInvalidInput", err)
	}

	var r *Responder
	if _, err := r.Finish([]byte("msgC")); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil Responder.Finish err=%v want ErrInvalidInput", err)
	}

	var s *Session
	if _, err := s.Export([]byte("label"), nil, 32); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil Session.Export err=%v want ErrInvalidInput", err)
	}
	if got := s.TranscriptID(); got != nil {
		t.Fatalf("nil Session.TranscriptID=%v want nil", got)
	}
}

func TestExportDomainSeparation(t *testing.T) {
	sI, _ := completeExchange(t, testInitiatorInput(), testResponderInput())

	cases := []struct {
		name                       string
		labelA, ctxA, labelB, ctxB []byte
	}{
		{"different label", []byte("a"), []byte("ctx"), []byte("b"), []byte("ctx")},
		{"different context", []byte("label"), []byte("a"), []byte("label"), []byte("b")},
		{"label/context boundary 1", []byte("ab"), []byte("c"), []byte("a"), []byte("bc")},
		{"label/context boundary 2", []byte("label"), []byte("ctx"), []byte("labelctx"), []byte("")},
		{"empty vs non-empty label", []byte(""), []byte("ctx"), []byte("a"), []byte("ctx")},
		{"empty vs non-empty context", []byte("label"), []byte(""), []byte("label"), []byte("a")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kA, err := sI.Export(tc.labelA, tc.ctxA, 32)
			if err != nil {
				t.Fatal(err)
			}
			kB, err := sI.Export(tc.labelB, tc.ctxB, 32)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(kA, kB) {
				t.Fatalf("Export(%q,%q)==Export(%q,%q): prefix-free domain separation broken",
					tc.labelA, tc.ctxA, tc.labelB, tc.ctxB)
			}
		})
	}

	t.Run("nil and empty slice equivalent", func(t *testing.T) {
		kNil, err := sI.Export([]byte("label"), nil, 32)
		if err != nil {
			t.Fatal(err)
		}
		kEmpty, err := sI.Export([]byte("label"), []byte{}, 32)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(kNil, kEmpty) {
			t.Fatal("Export(label,nil) != Export(label,[]byte{}): nil/empty inconsistent")
		}
	})

	t.Run("deterministic", func(t *testing.T) {
		k1, err := sI.Export([]byte("label"), []byte("ctx"), 32)
		if err != nil {
			t.Fatal(err)
		}
		k2, err := sI.Export([]byte("label"), []byte("ctx"), 32)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(k1, k2) {
			t.Fatal("Export is not deterministic for identical inputs")
		}
	})
}

func TestFinishConsumesStateOnConfirmationFailure(t *testing.T) {
	initCfg := testInitiatorInput()
	initCfg.Password = []byte("password-a")
	respCfg := testResponderInput()
	respCfg.Password = []byte("password-b")

	initiator, msgA, err := startTestInitiator(initCfg)
	if err != nil {
		t.Fatal(err)
	}
	_, msgB, err := respondTestResponder(respCfg, msgA)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrConfirmationFailed) {
		t.Fatalf("initiator Finish wrong-password err=%v want ErrConfirmationFailed", err)
	}
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("initiator second Finish after ErrConfirmationFailed err=%v want ErrStateUsed", err)
	}

	initiator2, msgA2, err := startTestInitiator(testInitiatorInput())
	if err != nil {
		t.Fatal(err)
	}
	responder2, msgB2, err := respondTestResponder(testResponderInput(), msgA2)
	if err != nil {
		t.Fatal(err)
	}
	msgC2, _, err := initiator2.Finish(msgB2)
	if err != nil {
		t.Fatal(err)
	}
	tampered := append([]byte(nil), msgC2...)
	tampered[len(tampered)-1] ^= 0xff
	if _, err := responder2.Finish(tampered); !errors.Is(err, ErrConfirmationFailed) {
		t.Fatalf("responder Finish tampered tagA err=%v want ErrConfirmationFailed", err)
	}
	if _, err := responder2.Finish(msgC2); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("responder second Finish after ErrConfirmationFailed err=%v want ErrStateUsed", err)
	}
}

func TestInitiatorFinishConsumesStateOnAbort(t *testing.T) {
	initiator, _, err := startTestInitiator(testInitiatorInput())
	if err != nil {
		t.Fatal(err)
	}
	// Identity Yb is rejected by scalarMultVFY (via decodePublicShare),
	// surfacing as ErrAbort. State must be consumed before the rejection
	// so that a retry returns ErrStateUsed rather than re-running the abort.
	identityYb := make([]byte, pointSize)
	msgB := encodeMessageB(identityYb, []byte("adb"), bytes.Repeat([]byte{0xaa}, tagSize))
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrAbort) {
		t.Fatalf("initiator Finish identity-Yb err=%v want ErrAbort", err)
	}
	if _, _, err := initiator.Finish(msgB); !errors.Is(err, ErrStateUsed) {
		t.Fatalf("initiator second Finish after ErrAbort err=%v want ErrStateUsed", err)
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

func completeExchange(t *testing.T, initCfg, respCfg Input) (*Session, *Session) {
	t.Helper()
	return newExchange(t, initCfg, respCfg).complete()
}

func allZero(in []byte) bool {
	for _, b := range in {
		if b != 0 {
			return false
		}
	}
	return true
}
