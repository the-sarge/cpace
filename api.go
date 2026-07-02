package cpace

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
)

const (
	// DraftVersion identifies the CPace Internet-Draft revision implemented by
	// this package.
	DraftVersion = "draft-irtf-cfrg-cpace-21"

	// currentSuite is the only suite implemented by v1 of this
	// package.
	currentSuite byte = 0x01
)

const suiteName = "CPACE-RISTR255-SHA512"

// Initiator is a single-use initiator state returned by Start.
type Initiator struct {
	state *initiatorState
}

type initiatorState = singleUseState[*initiatorCore]

// Responder is a single-use responder state returned by Respond.
type Responder struct {
	state *responderState
}

type responderState = singleUseState[*responderCore]

// Session is an explicitly confirmed CPace session. Copies of a Session share
// the same close state and secret key material.
type Session struct {
	state        *sessionState
	transcriptID []byte
	peerAD       []byte
	peerID       []byte
}

type sessionState struct {
	mu     sync.Mutex
	closed bool
	isk    []byte
}

// Start creates initiator state and message A.
func Start(input Input) (*Initiator, []byte, error) {
	return startWithRandom(input, rand.Reader)
}

func startWithRandom(input Input, random io.Reader) (*Initiator, []byte, error) {
	nc, err := normalizeStartInput(input)
	if err != nil {
		return nil, nil, err
	}
	defer nc.wipe()

	core, ya, err := newInitiatorCore(nc, random)
	if err != nil {
		return nil, nil, err
	}
	return &Initiator{state: newSingleUseState(core, "uninitialized initiator")}, encodeMessageA(nc.sid, ya, nc.ad), nil
}

// Respond consumes message A, creates responder state, and returns message B.
// Message B includes the responder's explicit key-confirmation tag. A nil
// error from Respond does not authenticate the initiator; authentication is
// established only by successful Finish calls.
func Respond(input Input, messageA []byte) (*Responder, []byte, error) {
	return respondWithRandom(input, messageA, rand.Reader)
}

func respondWithRandom(input Input, messageA []byte, random io.Reader) (*Responder, []byte, error) {
	nc, err := normalizeRespondInput(input)
	if err != nil {
		return nil, nil, err
	}
	defer nc.wipe()
	a, err := decodeMessageA(messageA)
	if err != nil {
		return nil, nil, err
	}
	if !bytes.Equal(a.sid, nc.sid) {
		return nil, nil, fmt.Errorf("%w: session id mismatch", ErrMessage)
	}
	core, yb, tagB, err := newResponderCore(nc, a.ya, a.ada, random)
	if err != nil {
		return nil, nil, err
	}
	return &Responder{state: newSingleUseState(core, "uninitialized responder")}, encodeMessageB(yb, nc.ad, tagB), nil
}

// Finish consumes message B, verifies the responder confirmation tag, and
// returns message C plus an authenticated session. The initiator state is
// consumed even when message parsing or confirmation fails.
func (i *Initiator) Finish(messageB []byte) ([]byte, *Session, error) {
	core, err := i.finishCore()
	if err != nil {
		return nil, nil, err
	}
	defer core.clear()
	b, err := decodeMessageB(messageB)
	if err != nil {
		return nil, nil, err
	}
	tagA, sess, err := core.finish(b.yb, b.adb, b.tag)
	if err != nil {
		return nil, nil, err
	}
	return encodeMessageC(tagA), sess, nil
}

// Close releases the persistent secret material held by the initiator state
// when an exchange is abandoned before Finish. Close is idempotent and
// nil-safe; calling Close on a nil *Initiator returns nil. Copies of an
// Initiator share the same terminal state, so closing one copy closes them all.
func (i *Initiator) Close() error {
	if i == nil {
		return nil
	}
	if i.state == nil {
		return fmt.Errorf("%w: uninitialized initiator", ErrInvalidInput)
	}
	return i.state.closeCore()
}

// Finish consumes message C, verifies the initiator confirmation tag, and
// returns an authenticated session. The responder state is consumed even when
// message parsing or confirmation fails.
func (r *Responder) Finish(messageC []byte) (*Session, error) {
	core, err := r.finishCore()
	if err != nil {
		return nil, err
	}
	defer core.clear()
	c, err := decodeMessageC(messageC)
	if err != nil {
		return nil, err
	}
	return core.finish(c.tag)
}

// Close releases the persistent secret material held by the responder state
// when an exchange is abandoned before Finish. Close is idempotent and
// nil-safe; calling Close on a nil *Responder returns nil. Copies of a
// Responder share the same terminal state, so closing one copy closes them all.
func (r *Responder) Close() error {
	if r == nil {
		return nil
	}
	if r.state == nil {
		return fmt.Errorf("%w: uninitialized responder", ErrInvalidInput)
	}
	return r.state.closeCore()
}

func (i *Initiator) finishCore() (*initiatorCore, error) {
	if i == nil || i.state == nil {
		return nil, fmt.Errorf("%w: uninitialized initiator", ErrInvalidInput)
	}
	return i.state.claimFinish()
}

func (r *Responder) finishCore() (*responderCore, error) {
	if r == nil || r.state == nil {
		return nil, fmt.Errorf("%w: uninitialized responder", ErrInvalidInput)
	}
	return r.state.claimFinish()
}

func newSession(isk, transcriptID, peerAD, peerID []byte) *Session {
	return &Session{
		state:        &sessionState{isk: clone(isk)},
		transcriptID: clone(transcriptID),
		peerAD:       clone(peerAD),
		peerID:       clone(peerID),
	}
}

func clone(in []byte) []byte {
	if in == nil {
		return nil
	}
	out := make([]byte, len(in))
	copy(out, in)
	return out
}
