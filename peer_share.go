package cpace

import (
	"crypto/hmac"
	"errors"
	"fmt"

	"github.com/gtank/ristretto255"
)

type peerShareRole string

const (
	initiatorPeerShare peerShareRole = "initiator"
	responderPeerShare peerShareRole = "responder"
)

func (r peerShareRole) decode(encoded []byte) (*ristretto255.Element, error) {
	p, err := decodePublicShare(encoded)
	if err != nil {
		return nil, r.wrapError(err)
	}
	return p, nil
}

func (r peerShareRole) sharedSecret(s *ristretto255.Scalar, encoded []byte) ([]byte, error) {
	k, err := scalarMultVFY(s, encoded)
	if err != nil {
		return nil, r.wrapError(err)
	}
	return k, nil
}

func (r peerShareRole) sharedSecretElement(s *ristretto255.Scalar, p *ristretto255.Element) ([]byte, error) {
	k, err := scalarMultVFYElement(s, p)
	if err != nil {
		return nil, r.wrapError(err)
	}
	return k, nil
}

// wrapError applies the ADR-0003 call-site sentinel mapping: the two exported
// peer-share sentinels are rewrapped from the plain sentinel, never from the
// helper's already-ErrAbort-wrapped error, with role context added.
// Non-sentinel defensive errors pass through unchanged.
// A new peer-share sentinel added in decodePublicShare must get a case here,
// or it surfaces without role context.
func (r peerShareRole) wrapError(err error) error {
	switch {
	case errors.Is(err, ErrPeerShareEncoding):
		return fmt.Errorf("%w: invalid %s share: %w", ErrAbort, r, ErrPeerShareEncoding)
	case errors.Is(err, ErrPeerShareIdentity):
		return fmt.Errorf("%w: invalid %s share: %w", ErrAbort, r, ErrPeerShareIdentity)
	default:
		return err
	}
}

func scalarMultVFY(s *ristretto255.Scalar, encoded []byte) ([]byte, error) {
	p, err := decodePublicShare(encoded)
	if err != nil {
		return nil, err
	}
	return scalarMultVFYElement(s, p)
}

func scalarMultVFYElement(s *ristretto255.Scalar, p *ristretto255.Element) ([]byte, error) {
	out := ristretto255.NewIdentityElement().ScalarMult(s, p).Bytes()
	if hmac.Equal(out, identityEncoding) {
		// Unreachable in production for prime-order Ristretto255: every
		// scalar sampleScalar can return is non-zero mod the group order, so
		// s*p is non-identity for any decoded non-identity p. Kept as
		// defense-in-depth; tests exercise it with a zero scalar.
		return nil, fmt.Errorf("%w: neutral-element shared secret", ErrAbort)
	}
	return out, nil
}

func decodePublicShare(encoded []byte) (*ristretto255.Element, error) {
	// Defensive for internal callers; public message decoders enforce
	// pointSize, so malformed wire lengths surface as ErrMessage from framing
	// and never reach this branch.
	if len(encoded) != pointSize {
		return nil, fmt.Errorf("%w: invalid peer share length", ErrAbort)
	}
	p, err := ristretto255.NewIdentityElement().SetCanonicalBytes(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAbort, ErrPeerShareEncoding)
	}
	if hmac.Equal(p.Bytes(), identityEncoding) {
		return nil, fmt.Errorf("%w: %w", ErrAbort, ErrPeerShareIdentity)
	}
	return p, nil
}
