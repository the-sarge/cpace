package cpace

import (
	"bytes"
	"io"
	"testing"

	"github.com/gtank/ristretto255"
)

type exchangeFixture struct {
	tb        testing.TB
	initiator *Initiator
	responder *Responder
	msgA      []byte
	msgB      []byte
}

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

func repeatingRand(fill byte) io.Reader {
	return &repeatingReader{buf: bytes.Repeat([]byte{fill}, scalarSize)}
}

func testInitiatorInput() Input {
	return Input{
		Password:  []byte("password"),
		SelfID:    []byte("initiator"),
		PeerID:    []byte("responder"),
		Context:   []byte("context"),
		SessionID: []byte("sid"),
	}
}

func testResponderInput() Input {
	return Input{
		Password:  []byte("password"),
		SelfID:    []byte("responder"),
		PeerID:    []byte("initiator"),
		Context:   []byte("context"),
		SessionID: []byte("sid"),
	}
}

func defaultExchangeInputs() (Input, Input) {
	initInput := testInitiatorInput()
	initInput.LocalAssociatedData = []byte("ADa")
	respInput := testResponderInput()
	respInput.LocalAssociatedData = []byte("ADb")
	return initInput, respInput
}

func startTestInitiator(cfg Input) (*Initiator, []byte, error) {
	return startWithRandom(cfg, repeatingRand(0x11))
}

func respondTestResponder(cfg Input, messageA []byte) (*Responder, []byte, error) {
	return respondWithRandom(cfg, messageA, repeatingRand(0x22))
}

func newExchange(tb testing.TB, initInput, respInput Input) *exchangeFixture {
	tb.Helper()
	initiator, msgA, err := startTestInitiator(initInput)
	if err != nil {
		tb.Fatalf("Start failed for fixed exchange config: %v", err)
	}
	responder, msgB, err := respondTestResponder(respInput, msgA)
	if err != nil {
		tb.Fatalf("Respond failed for fixed exchange config: %v", err)
	}
	return &exchangeFixture{
		tb:        tb,
		initiator: initiator,
		responder: responder,
		msgA:      msgA,
		msgB:      msgB,
	}
}

func (x *exchangeFixture) finishInitiator() ([]byte, *Session) {
	x.tb.Helper()
	msgC, session, err := x.initiator.Finish(x.msgB)
	if err != nil {
		x.tb.Fatalf("initiator Finish failed for fixed exchange config: %v", err)
	}
	return msgC, session
}

func (x *exchangeFixture) finishResponder(msgC []byte) *Session {
	x.tb.Helper()
	session, err := x.responder.Finish(msgC)
	if err != nil {
		x.tb.Fatalf("responder Finish failed for fixed exchange config: %v", err)
	}
	return session
}

func (x *exchangeFixture) complete() (*Session, *Session) {
	x.tb.Helper()
	msgC, initSession := x.finishInitiator()
	respSession := x.finishResponder(msgC)
	return initSession, respSession
}

// The secret snapshots below are the only place hygiene tests may learn the
// cores' field layout. Snapshot before the terminal operation (the aliases
// must be captured while the secrets are still live), then assertCleared
// after it.

type initiatorSecretSnapshot struct {
	tb        testing.TB
	initiator *Initiator
	scalar    *ristretto255.Scalar
}

func snapshotInitiatorSecrets(tb testing.TB, i *Initiator) initiatorSecretSnapshot {
	tb.Helper()
	scalar := i.state.core.scalar
	if scalar == nil {
		tb.Fatal("snapshot taken after initiator scalar was already cleared")
	}
	return initiatorSecretSnapshot{tb: tb, initiator: i, scalar: scalar}
}

func (s initiatorSecretSnapshot) assertCleared() {
	s.tb.Helper()
	if s.initiator.state.core.scalar != nil {
		s.tb.Fatal("initiator core retained scalar reference after terminal operation")
	}
	if !allZero(s.scalar.Bytes()) {
		s.tb.Fatal("initiator scalar backing array was not cleared")
	}
}

type responderSecretSnapshot struct {
	tb         testing.TB
	responder  *Responder
	isk        []byte
	transcript [][]byte
}

func snapshotResponderSecrets(tb testing.TB, r *Responder) responderSecretSnapshot {
	tb.Helper()
	core := r.state.core
	if core.isk == nil {
		tb.Fatal("snapshot taken after responder ISK was already cleared")
	}
	tr := &core.transcript
	return responderSecretSnapshot{
		tb:         tb,
		responder:  r,
		isk:        core.isk,
		transcript: [][]byte{tr.transcript, tr.ya, tr.ada, tr.yb, tr.adb},
	}
}

func (s responderSecretSnapshot) assertCleared() {
	s.tb.Helper()
	core := s.responder.state.core
	if core.isk != nil {
		s.tb.Fatal("responder core retained ISK reference after terminal operation")
	}
	tr := &core.transcript
	if tr.transcript != nil || tr.ya != nil || tr.ada != nil || tr.yb != nil || tr.adb != nil {
		s.tb.Fatal("responder core retained transcript references after terminal operation")
	}
	if !allZero(s.isk) {
		s.tb.Fatal("responder-owned ISK backing array was not cleared")
	}
	for _, b := range s.transcript {
		if !allZero(b) {
			s.tb.Fatal("responder-owned transcript backing array was not cleared")
		}
	}
}
