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
	isk        []byte       // persistent secret — owned by clear()
	transcript irTranscript // public wire data — zeroed alongside isk as hygiene
	sid        []byte
	peerID     []byte
}

func newInitiatorCore(ni normalizedInput, random io.Reader) (*initiatorCore, []byte, error) {
	if random == nil {
		random = rand.Reader
	}
	g := calculateGenerator(ni.password, ni.ci, ni.sid)
	defer clearElement(g)
	// Early-clear the password so its residency is bounded by the generator
	// derivation, not the full constructor lifetime. ni is a by-value copy:
	// clearBytes zeroes the shared backing array, and the shell's deferred
	// ni.wipe() re-covers the field on every exit path, including this
	// constructor's error returns and panics.
	clearBytes(ni.password)
	ni.password = nil
	y, err := sampleScalar(random)
	if err != nil {
		return nil, nil, err
	}
	ya := scalarMult(y, g)
	return &initiatorCore{
		scalar: y,
		sid:    clone(ni.sid),
		ya:     clone(ya),
		ada:    clone(ni.ad),
		peerID: clone(ni.responderID),
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
	// Reject equal confirmation tags before releasing our tag or a Session.
	// CI role binding does not distinguish identical MAC inputs (#308).
	if hmac.Equal(tagA, peerTag) {
		return nil, nil, ErrConfirmationFailed
	}
	return tagA, newSession(isk, tr.transcriptID(), peerAdb, c.peerID), nil
}

func newResponderCore(ni normalizedInput, peerYa, peerAda []byte, random io.Reader) (*responderCore, []byte, []byte, error) {
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
	g := calculateGenerator(ni.password, ni.ci, ni.sid)
	defer clearElement(g)
	// Early-clear the password as in newInitiatorCore; the shell's deferred
	// ni.wipe() re-covers the field on exit.
	clearBytes(ni.password)
	ni.password = nil
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
	tr := newIRTranscript(peerYa, peerAda, yb, ni.ad)
	isk := tr.deriveISK(ni.sid, k)
	tagB := tr.responderConfirmationTag(isk, ni.sid)
	return &responderCore{
		isk:        isk,
		transcript: tr,
		sid:        clone(ni.sid),
		peerID:     clone(ni.initiatorID),
	}, yb, tagB, nil
}

func (c *responderCore) finish(peerTagC []byte) (*Session, error) {
	expectedA := c.transcript.initiatorConfirmationTag(c.isk, c.sid)
	if !hmac.Equal(expectedA, peerTagC) {
		return nil, ErrConfirmationFailed
	}
	// A correct tag is not peer confirmation if it is also our own tag.
	// Equal shares and AD otherwise let B's tag be reflected as C (#308).
	tagB := c.transcript.responderConfirmationTag(c.isk, c.sid)
	if hmac.Equal(tagB, peerTagC) {
		return nil, ErrConfirmationFailed
	}
	return newSession(c.isk, c.transcript.transcriptID(), c.transcript.initiatorAD(), c.peerID), nil
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

// clear zeroes then nils the persistent ISK and wipes the stored transcript
// bytes as hygiene; a second call finds nil and is a safe no-op. Safe on a nil
// receiver.
func (c *responderCore) clear() {
	if c == nil {
		return
	}
	clearBytes(c.isk)
	c.isk = nil
	c.transcript.clear()
}
