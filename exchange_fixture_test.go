package cpace

import (
	"bytes"
	"io"
	"testing"
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
