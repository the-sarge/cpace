package cpace

import (
	"crypto/hmac"
	"crypto/rand"
	"io"

	"github.com/gtank/ristretto255"
)

type initiatorCore struct {
	scalar *ristretto255.Scalar // persistent secret — owned by clear()
	sid    []byte
	ya     []byte
	ada    []byte
	peerID []byte
}

type responderCore struct {
	isk        []byte // persistent secret — owned by clear()
	transcript []byte // public wire data — zeroed alongside isk as hygiene
	sid        []byte
	ya         []byte
	ada        []byte
	peerID     []byte
}

func newInitiatorCore(nc normalizedInput, random io.Reader) (*initiatorCore, []byte, error) {
	if random == nil {
		random = rand.Reader
	}
	g := calculateGenerator(nc.password, nc.ci, nc.sid)
	defer clearElement(g)
	// Early-clear the password so its residency is bounded by the generator
	// derivation, not the full constructor lifetime. nc is a by-value copy:
	// clearBytes zeroes the shared backing array, and the shell's deferred
	// nc.wipe() re-covers the field on every exit path, including this
	// constructor's error returns and panics.
	clearBytes(nc.password)
	nc.password = nil
	y, err := sampleScalar(random)
	if err != nil {
		return nil, nil, err
	}
	ya := scalarMult(y, g)
	return &initiatorCore{
		scalar: y,
		sid:    clone(nc.sid),
		ya:     clone(ya),
		ada:    clone(nc.ad),
		peerID: clone(nc.responderID),
	}, ya, nil
}

func (c *initiatorCore) finish(peerYb, peerAdb, peerTag []byte) ([]byte, *Session, error) {
	k, err := responderPeerShare.sharedSecret(c.scalar, peerYb)
	defer clearBytes(k)
	if err != nil {
		return nil, nil, err
	}
	tr := newIRTranscript(c.ya, c.ada, peerYb, peerAdb)
	isk := tr.deriveISK(c.sid, k)
	// Scratch — the initiator's finish-local ISK. The deferred wipe covers
	// the tag computations below, including panic paths; newSession clones
	// isk, so the returned Session is unaffected.
	defer clearBytes(isk)
	expectedB := tr.responderConfirmationTag(isk, c.sid)
	if !hmac.Equal(expectedB, peerTag) {
		return nil, nil, ErrConfirmationFailed
	}
	tagA := tr.initiatorConfirmationTag(isk, c.sid)
	return tagA, newSession(isk, tr.bytes(), peerAdb, c.peerID), nil
}

func newResponderCore(nc normalizedInput, peerYa, peerAda []byte, random io.Reader) (*responderCore, []byte, []byte, error) {
	if random == nil {
		random = rand.Reader
	}
	// Validate Ya FIRST before generator derivation and scalar sampling;
	// the ordering is pinned by
	// TestResponderPrevalidatesInvalidInitiatorShareBeforeRandomness.
	peerYaElement, err := initiatorPeerShare.decode(peerYa)
	if err != nil {
		return nil, nil, nil, err
	}
	g := calculateGenerator(nc.password, nc.ci, nc.sid)
	defer clearElement(g)
	// Early-clear the password as in newInitiatorCore; the shell's deferred
	// nc.wipe() re-covers the field on exit.
	clearBytes(nc.password)
	nc.password = nil
	y, err := sampleScalar(random)
	if err != nil {
		return nil, nil, nil, err
	}
	defer clearScalar(y)
	yb := scalarMult(y, g)
	k, err := initiatorPeerShare.sharedSecretElement(y, peerYaElement)
	defer clearBytes(k)
	if err != nil {
		return nil, nil, nil, err
	}
	tr := newIRTranscript(peerYa, peerAda, yb, nc.ad)
	isk := tr.deriveISK(nc.sid, k)
	tagB := tr.responderConfirmationTag(isk, nc.sid)
	return &responderCore{
		isk:        isk,
		transcript: tr.bytes(),
		sid:        clone(nc.sid),
		ya:         clone(peerYa),
		ada:        clone(peerAda),
		peerID:     clone(nc.initiatorID),
	}, yb, tagB, nil
}

func (c *responderCore) finish(peerTagC []byte) (*Session, error) {
	expectedA := initiatorRoleConfirmationTag(c.isk, c.sid, c.ya, c.ada)
	if !hmac.Equal(expectedA, peerTagC) {
		return nil, ErrConfirmationFailed
	}
	return newSession(c.isk, c.transcript, c.ada, c.peerID), nil
}

// clear zeroes then nils each persistent-secret field; a second call finds
// nil and is a safe no-op. Safe on a nil receiver.
func (c *initiatorCore) clear() {
	if c == nil {
		return
	}
	clearScalar(c.scalar)
	c.scalar = nil
}

// clear zeroes then nils each persistent-secret field; a second call finds
// nil and is a safe no-op. Safe on a nil receiver.
func (c *responderCore) clear() {
	if c == nil {
		return
	}
	clearBytes(c.isk)
	clearBytes(c.transcript)
	c.isk = nil
	c.transcript = nil
}
