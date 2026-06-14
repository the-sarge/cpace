package cpace

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
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

// Config contains the local inputs for one CPace role.
//
// Password, InitiatorID, and ResponderID must be non-empty. Context and
// AssociatedData may be empty. Both parties must use the same role orientation:
// InitiatorID is the party that called Start, and ResponderID is the party that
// called Respond. SessionID must be a fresh, non-secret, parties-agree-on value
// for every session. Empty SessionID values are rejected by default because they
// weaken replay and transcript separation properties. Set AllowEmptySessionID
// only for draft-21 compatibility tests or profiles that have deliberately
// accepted the weaker empty-sid behavior. Scalar randomness always comes from
// crypto/rand.Reader. The AssociatedData field is ADa for Start and ADb for
// Respond. Field lengths are capped at 4 KiB for Password and IDs, 1 KiB for
// Context and SessionID, and 64 KiB for AssociatedData. Inputs exceeding these
// caps are rejected before copying; accepted byte slices are copied by Start
// and Respond before use.
type Config struct {
	Password            []byte
	InitiatorID         []byte
	ResponderID         []byte
	Context             []byte
	SessionID           []byte
	AssociatedData      []byte
	AllowEmptySessionID bool
}

// Initiator is a single-use initiator state returned by Start.
type Initiator struct {
	mu   sync.Mutex
	used bool

	// core is never reassigned or nil'd after construction: clear() zeroes
	// and nils the core's fields, not this pointer. The unsynchronized
	// core == nil check in Finish and the ErrStateUsed-after-use semantics
	// both rely on this.
	core *initiatorCore
}

// Responder is a single-use responder state returned by Respond.
type Responder struct {
	mu   sync.Mutex
	used bool

	// core is never reassigned or nil'd after construction — same invariant
	// as Initiator.core.
	core *responderCore
}

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

type normalizedConfig struct {
	password    []byte
	initiatorID []byte
	responderID []byte
	ci          []byte
	sid         []byte
	ad          []byte
}

// wipe performs best-effort zeroization of every byte slice owned by the
// normalized config. Called via defer in startWithRandom and respondWithRandom
// so that all cloned input bytes are cleared on every exit path — including
// core-constructor error returns and panics. Idempotent against fields whose
// backing arrays were already zeroed behind the core seam (the password is
// eagerly cleared inside the core constructors).
func (nc *normalizedConfig) wipe() {
	clearBytes(nc.password)
	clearBytes(nc.initiatorID)
	clearBytes(nc.responderID)
	clearBytes(nc.ci)
	clearBytes(nc.sid)
	clearBytes(nc.ad)
}

// Start creates initiator state and message A.
func Start(cfg Config) (*Initiator, []byte, error) {
	return startWithRandom(cfg, rand.Reader)
}

func startWithRandom(cfg Config, random io.Reader) (*Initiator, []byte, error) {
	nc, err := normalizeConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	defer nc.wipe()

	core, ya, err := newInitiatorCore(nc, random)
	if err != nil {
		return nil, nil, err
	}
	return &Initiator{core: core}, encodeMessageA(nc.sid, ya, nc.ad), nil
}

// Respond consumes message A, creates responder state, and returns message B.
// Message B includes the responder's explicit key-confirmation tag. A nil
// error from Respond does not authenticate the initiator; authentication is
// established only by successful Finish calls.
func Respond(cfg Config, messageA []byte) (*Responder, []byte, error) {
	return respondWithRandom(cfg, messageA, rand.Reader)
}

func respondWithRandom(cfg Config, messageA []byte, random io.Reader) (*Responder, []byte, error) {
	nc, err := normalizeConfig(cfg)
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
	return &Responder{core: core}, encodeMessageB(yb, nc.ad, tagB), nil
}

// Finish consumes message B, verifies the responder confirmation tag, and
// returns message C plus an authenticated session. The initiator state is
// consumed even when message parsing or confirmation fails.
func (i *Initiator) Finish(messageB []byte) ([]byte, *Session, error) {
	if i == nil || i.core == nil {
		return nil, nil, fmt.Errorf("%w: uninitialized initiator", ErrInvalidInput)
	}
	if err := i.consume(); err != nil {
		return nil, nil, err
	}
	defer i.core.clear()
	b, err := decodeMessageB(messageB)
	if err != nil {
		return nil, nil, err
	}
	tagA, sess, err := i.core.finish(b.yb, b.adb, b.tag)
	if err != nil {
		return nil, nil, err
	}
	return encodeMessageC(tagA), sess, nil
}

// Finish consumes message C, verifies the initiator confirmation tag, and
// returns an authenticated session. The responder state is consumed even when
// message parsing or confirmation fails.
func (r *Responder) Finish(messageC []byte) (*Session, error) {
	if r == nil || r.core == nil {
		return nil, fmt.Errorf("%w: uninitialized responder", ErrInvalidInput)
	}
	if err := r.consume(); err != nil {
		return nil, err
	}
	defer r.core.clear()
	c, err := decodeMessageC(messageC)
	if err != nil {
		return nil, err
	}
	return r.core.finish(c.tag)
}

func (i *Initiator) consume() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.used {
		return ErrStateUsed
	}
	i.used = true
	return nil
}

func (r *Responder) consume() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.used {
		return ErrStateUsed
	}
	r.used = true
	return nil
}

func normalizeConfig(cfg Config) (normalizedConfig, error) {
	if len(cfg.Password) == 0 {
		return normalizedConfig{}, fmt.Errorf("%w: empty password", ErrInvalidInput)
	}
	if len(cfg.InitiatorID) == 0 {
		return normalizedConfig{}, fmt.Errorf("%w: empty initiator id", ErrInvalidInput)
	}
	if len(cfg.ResponderID) == 0 {
		return normalizedConfig{}, fmt.Errorf("%w: empty responder id", ErrInvalidInput)
	}
	if len(cfg.SessionID) == 0 && !cfg.AllowEmptySessionID {
		return normalizedConfig{}, fmt.Errorf("%w: %w", ErrInvalidInput, ErrEmptySessionID)
	}
	if len(cfg.Password) > passwordCap.length {
		return normalizedConfig{}, fmt.Errorf("%w: %s too large", ErrInvalidInput, passwordCap.name)
	}
	if len(cfg.InitiatorID) > initiatorIDCap.length {
		return normalizedConfig{}, fmt.Errorf("%w: %s too large", ErrInvalidInput, initiatorIDCap.name)
	}
	if len(cfg.ResponderID) > responderIDCap.length {
		return normalizedConfig{}, fmt.Errorf("%w: %s too large", ErrInvalidInput, responderIDCap.name)
	}
	if len(cfg.Context) > contextCap.length {
		return normalizedConfig{}, fmt.Errorf("%w: %s too large", ErrInvalidInput, contextCap.name)
	}
	if len(cfg.SessionID) > sessionIDCap.length {
		return normalizedConfig{}, fmt.Errorf("%w: %s too large", ErrInvalidInput, sessionIDCap.name)
	}
	if len(cfg.AssociatedData) > associatedDataCap.length {
		return normalizedConfig{}, fmt.Errorf("%w: %s too large", ErrInvalidInput, associatedDataCap.name)
	}
	return normalizedConfig{
		password:    clone(cfg.Password),
		initiatorID: clone(cfg.InitiatorID),
		responderID: clone(cfg.ResponderID),
		ci:          buildCI(cfg.InitiatorID, cfg.ResponderID, cfg.Context),
		sid:         clone(cfg.SessionID),
		ad:          clone(cfg.AssociatedData),
	}, nil
}

func buildCI(initiatorID, responderID, context []byte) []byte {
	return lvCat(
		[]byte("CPace-Go-CI"),
		[]byte(DraftVersion),
		[]byte(suiteName),
		[]byte("initiator"),
		initiatorID,
		[]byte("responder"),
		responderID,
		[]byte("context"),
		context,
	)
}

func newSession(isk, transcript, peerAD, peerID []byte) *Session {
	sidOut := sha512.Sum512(append([]byte("CPaceSidOutput"), transcript...))
	return &Session{
		state:        &sessionState{isk: clone(isk)},
		transcriptID: sidOut[:],
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
