package cpace

import (
	"crypto/hmac"
	"crypto/rand"
	"io"

	"github.com/gtank/ristretto255"
)

type initiatorCore struct {
	scalar *ristretto255.Scalar
	sid    []byte
	ya     []byte
	ada    []byte
	peerID []byte
}

type responderCore struct {
	isk        []byte
	transcript []byte
	sid        []byte
	ya         []byte
	ada        []byte
	peerID     []byte
}

func newInitiatorCore(nc normalizedConfig, random io.Reader) (*initiatorCore, []byte, error) {
	if random == nil {
		random = rand.Reader
	}
	g := calculateGenerator(nc.password, nc.ci, nc.sid)
	defer clearElement(g)
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
	k, err := scalarMultVFY(c.scalar, peerYb)
	defer clearBytes(k)
	if err != nil {
		return nil, nil, wrapPeerShareError(err, "responder")
	}
	tr := transcriptIR(c.ya, c.ada, peerYb, peerAdb)
	isk := deriveISK(c.sid, k, tr)
	defer clearBytes(isk)
	expectedB := confirmationTag(isk, c.sid, peerYb, peerAdb)
	if !hmac.Equal(expectedB, peerTag) {
		return nil, nil, ErrConfirmationFailed
	}
	tagA := confirmationTag(isk, c.sid, c.ya, c.ada)
	return tagA, newSession(isk, tr, peerAdb, c.peerID), nil
}

func newResponderCore(nc normalizedConfig, peerYa, peerAda []byte, random io.Reader) (*responderCore, []byte, []byte, error) {
	if random == nil {
		random = rand.Reader
	}
	if _, err := decodePublicShare(peerYa); err != nil {
		return nil, nil, nil, wrapPeerShareError(err, "initiator")
	}
	g := calculateGenerator(nc.password, nc.ci, nc.sid)
	defer clearElement(g)
	clearBytes(nc.password)
	nc.password = nil
	y, err := sampleScalar(random)
	if err != nil {
		return nil, nil, nil, err
	}
	defer clearScalar(y)
	yb := scalarMult(y, g)
	k, err := scalarMultVFY(y, peerYa)
	defer clearBytes(k)
	if err != nil {
		return nil, nil, nil, wrapPeerShareError(err, "initiator")
	}
	tr := transcriptIR(peerYa, peerAda, yb, nc.ad)
	isk := deriveISK(nc.sid, k, tr)
	tagB := confirmationTag(isk, nc.sid, yb, nc.ad)
	return &responderCore{
		isk:        isk,
		transcript: tr,
		sid:        clone(nc.sid),
		ya:         clone(peerYa),
		ada:        clone(peerAda),
		peerID:     clone(nc.initiatorID),
	}, yb, tagB, nil
}

func (c *responderCore) finish(peerTagC []byte) (*Session, error) {
	expectedA := confirmationTag(c.isk, c.sid, c.ya, c.ada)
	if !hmac.Equal(expectedA, peerTagC) {
		return nil, ErrConfirmationFailed
	}
	return newSession(c.isk, c.transcript, c.ada, c.peerID), nil
}
