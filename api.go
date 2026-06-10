package cpace

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"
	"sync"

	"github.com/gtank/ristretto255"
)

const (
	// DraftVersion identifies the CPace Internet-Draft revision implemented by
	// this package.
	DraftVersion = "draft-irtf-cfrg-cpace-21"

	// SuiteCPaceRistretto255SHA512 is the only suite implemented by v1 of this
	// package.
	SuiteCPaceRistretto255SHA512 Suite = 0x01
)

const suiteName = "CPACE-RISTR255-SHA512"

// Suite identifies a CPace ciphersuite in this package's wire framing.
type Suite byte

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

	scalar *ristretto255.Scalar
	sid    []byte
	ya     []byte
	ada    []byte
	peerID []byte
}

// Responder is a single-use responder state returned by Respond.
type Responder struct {
	mu   sync.Mutex
	used bool

	sid        []byte
	ya         []byte
	ada        []byte
	yb         []byte
	adb        []byte
	isk        []byte
	transcript []byte
	peerID     []byte
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
// normalized config. Called via defer at the top of Start/Respond so that all
// cloned input bytes are cleared on every exit path, including failure paths
// and panics. Idempotent against fields that have already been cleared and
// set to nil earlier in the function.
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
	if random == nil {
		random = rand.Reader
	}
	nc, err := normalizeConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	defer nc.wipe()

	g := calculateGenerator(nc.password, nc.ci, nc.sid)
	defer clearElement(g)
	// Early-clear the password so its residency is bounded by the generator
	// derivation, not the full function lifetime. nc.wipe() will still cover
	// the (now-nil) field on exit alongside the other normalized fields.
	clearBytes(nc.password)
	nc.password = nil
	y, err := sampleScalar(random)
	if err != nil {
		return nil, nil, err
	}
	ya := scalarMult(y, g)
	msg := encodeMessageA(nc.sid, ya, nc.ad)

	st := &Initiator{
		scalar: y,
		sid:    clone(nc.sid),
		ya:     clone(ya),
		ada:    clone(nc.ad),
		peerID: clone(nc.responderID),
	}
	return st, msg, nil
}

// Respond consumes message A, creates responder state, and returns message B.
// Message B includes the responder's explicit key-confirmation tag. A nil
// error from Respond does not authenticate the initiator; authentication is
// established only by successful Finish calls.
func Respond(cfg Config, messageA []byte) (*Responder, []byte, error) {
	return respondWithRandom(cfg, messageA, rand.Reader)
}

func respondWithRandom(cfg Config, messageA []byte, random io.Reader) (*Responder, []byte, error) {
	if random == nil {
		random = rand.Reader
	}
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
	if _, ok := decodePublicShare(a.ya); !ok {
		return nil, nil, fmt.Errorf("%w: invalid initiator share", ErrAbort)
	}

	g := calculateGenerator(nc.password, nc.ci, nc.sid)
	defer clearElement(g)
	// Early-clear the password as in startWithRandom; nc.wipe() handles the
	// remaining normalized fields on exit.
	clearBytes(nc.password)
	nc.password = nil
	y, err := sampleScalar(random)
	if err != nil {
		return nil, nil, err
	}
	defer clearScalar(y)
	yb := scalarMult(y, g)
	k, ok := scalarMultVFY(y, a.ya)
	defer clearBytes(k)
	if !ok {
		return nil, nil, fmt.Errorf("%w: invalid initiator share", ErrAbort)
	}
	tr := transcriptIR(a.ya, a.ada, yb, nc.ad)
	isk := deriveISK(nc.sid, k, tr)
	tagB := confirmationTag(isk, nc.sid, yb, nc.ad)
	msg := encodeMessageB(yb, nc.ad, tagB)

	st := &Responder{
		sid:        clone(nc.sid),
		ya:         clone(a.ya),
		ada:        clone(a.ada),
		yb:         clone(yb),
		adb:        clone(nc.ad),
		isk:        isk,
		transcript: tr,
		peerID:     clone(nc.initiatorID),
	}
	return st, msg, nil
}

// Finish consumes message B, verifies the responder confirmation tag, and
// returns message C plus an authenticated session. The initiator state is
// consumed even when message parsing or confirmation fails.
func (i *Initiator) Finish(messageB []byte) ([]byte, *Session, error) {
	if i == nil {
		return nil, nil, fmt.Errorf("%w: nil initiator", ErrInvalidInput)
	}
	if err := i.consume(); err != nil {
		return nil, nil, err
	}
	scalar := i.scalar
	defer func() {
		clearScalar(scalar)
		i.scalar = nil
	}()
	b, err := decodeMessageB(messageB)
	if err != nil {
		return nil, nil, err
	}
	k, ok := scalarMultVFY(scalar, b.yb)
	defer clearBytes(k)
	if !ok {
		return nil, nil, fmt.Errorf("%w: invalid responder share", ErrAbort)
	}
	tr := transcriptIR(i.ya, i.ada, b.yb, b.adb)
	isk := deriveISK(i.sid, k, tr)
	// Defer ISK wipe so every exit path — including future early returns and
	// panics between here and the success branch — clears the secret. Mirrors
	// the deferred wipe in Responder.Finish. newSession clones isk into the
	// returned Session, so this wipe does not affect the caller's Session.
	defer clearBytes(isk)
	expectedB := confirmationTag(isk, i.sid, b.yb, b.adb)
	if !hmac.Equal(expectedB, b.tag) {
		return nil, nil, ErrConfirmationFailed
	}
	tagA := confirmationTag(isk, i.sid, i.ya, i.ada)
	msgC := encodeMessageC(tagA)
	sess := newSession(isk, tr, b.adb, i.peerID)
	return msgC, sess, nil
}

// Finish consumes message C, verifies the initiator confirmation tag, and
// returns an authenticated session. The responder state is consumed even when
// message parsing or confirmation fails.
func (r *Responder) Finish(messageC []byte) (*Session, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: nil responder", ErrInvalidInput)
	}
	if err := r.consume(); err != nil {
		return nil, err
	}
	defer func() {
		clearBytes(r.isk)
		clearBytes(r.transcript)
		r.isk = nil
		r.transcript = nil
	}()
	c, err := decodeMessageC(messageC)
	if err != nil {
		return nil, err
	}
	expectedA := confirmationTag(r.isk, r.sid, r.ya, r.ada)
	if !hmac.Equal(expectedA, c.tag) {
		return nil, ErrConfirmationFailed
	}
	sess := newSession(r.isk, r.transcript, r.ada, r.peerID)
	return sess, nil
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
	if len(cfg.Password) > maxPasswordLength {
		return normalizedConfig{}, fmt.Errorf("%w: password too large", ErrInvalidInput)
	}
	if len(cfg.InitiatorID) > maxIDLength {
		return normalizedConfig{}, fmt.Errorf("%w: initiator id too large", ErrInvalidInput)
	}
	if len(cfg.ResponderID) > maxIDLength {
		return normalizedConfig{}, fmt.Errorf("%w: responder id too large", ErrInvalidInput)
	}
	if len(cfg.Context) > maxContextLength {
		return normalizedConfig{}, fmt.Errorf("%w: context too large", ErrInvalidInput)
	}
	if len(cfg.SessionID) > maxSessionIDLength {
		return normalizedConfig{}, fmt.Errorf("%w: session id too large", ErrInvalidInput)
	}
	if len(cfg.AssociatedData) > maxAssociatedDataLength {
		return normalizedConfig{}, fmt.Errorf("%w: associated data too large", ErrInvalidInput)
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
