package cpace

import (
	"crypto/hkdf"
	"crypto/sha512"
	"fmt"
	"sync"
)

const maxHKDFOutput = 255 * 64

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

// TranscriptID returns the draft CPaceSidOutput value for the confirmed
// initiator-responder CPace transcript. It is not a complete channel binding
// for any outer version, suite, or application-protocol negotiation.
// TranscriptID remains available after Close.
func (s *Session) TranscriptID() []byte {
	if s == nil {
		return nil
	}
	return clone(s.transcriptID)
}

// PeerAssociatedData returns the peer associated data that was bound into the
// confirmed exchange. The returned slice is a copy and remains available after
// Close.
func (s *Session) PeerAssociatedData() []byte {
	if s == nil {
		return nil
	}
	return clone(s.peerAD)
}

// PeerID returns the caller-configured peer identity that was bound into CI and
// confirmed by the completed exchange. The value is copied from Input; it is
// not parsed from peer-controlled wire data. The returned slice is a copy and
// remains available after Close.
func (s *Session) PeerID() []byte {
	if s == nil {
		return nil
	}
	return clone(s.peerID)
}

// Close releases the secret key material held by the Session. Close is
// idempotent and nil-safe; calling Close on a nil *Session returns nil. It
// performs best-effort in-memory key cleanup, but Go does not provide
// guaranteed secure memory erasure and the runtime or compiler may make
// additional copies. Non-secret metadata such as TranscriptID,
// PeerAssociatedData, and PeerID remains available after Close.
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	if s.state == nil {
		return fmt.Errorf("%w: nil session", ErrInvalidInput)
	}
	st := s.state
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.closed {
		return nil
	}
	clearBytes(st.isk)
	st.isk = nil
	st.closed = true
	return nil
}

// Export derives deterministic application key material from the confirmed ISK
// using HKDF-SHA512. The label and context are prefix-free encoded into HKDF
// info. Export output is not fresh randomness or a randomness pool; use
// separate, domain-specific labels and contexts for each application purpose.
// length must be in the range [0, 16320] (255 * 64, the HKDF-SHA512 maximum).
// A length of zero returns a result with length 0; callers must not distinguish
// nil from empty output for this case. Negative values and values exceeding the
// maximum are rejected with a wrapped ErrInvalidInput.
func (s *Session) Export(label, context []byte, length int) ([]byte, error) {
	if s == nil || s.state == nil {
		return nil, fmt.Errorf("%w: nil session", ErrInvalidInput)
	}
	st := s.state
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.closed {
		return nil, ErrSessionClosed
	}
	// Negative lengths must be rejected here: crypto/hkdf.Key panics on
	// negative length, so do not delegate this validation to the standard
	// library. See ADR-0005.
	if length < 0 || length > maxHKDFOutput {
		return nil, fmt.Errorf("%w: invalid export length", ErrInvalidInput)
	}
	info := lvCat([]byte("CPaceExport"), label, context)
	out, err := hkdf.Key(sha512.New, st.isk, nil, string(info), length)
	if err != nil {
		return nil, fmt.Errorf("%w: export failed: %w", ErrInvalidInput, err)
	}
	return out, nil
}

func newSession(isk, transcriptID, peerAD, peerID []byte) *Session {
	return &Session{
		state:        &sessionState{isk: clone(isk)},
		transcriptID: clone(transcriptID),
		peerAD:       clone(peerAD),
		peerID:       clone(peerID),
	}
}
