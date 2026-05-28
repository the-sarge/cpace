package cpace

import (
	"crypto/hkdf"
	"crypto/sha512"
	"fmt"
)

const maxHKDFOutput = 255 * 64

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
// confirmed by the completed exchange. The value is copied from Config; it is
// not parsed from peer-controlled wire data. The returned slice is a copy and
// remains available after Close.
func (s *Session) PeerID() []byte {
	if s == nil {
		return nil
	}
	return clone(s.peerID)
}

// Close releases the secret key material held by the Session. Close is
// idempotent. It performs best-effort in-memory key cleanup, but Go does not
// provide guaranteed secure memory erasure and the runtime or compiler may make
// additional copies. Non-secret metadata such as TranscriptID,
// PeerAssociatedData, and PeerID remains available after Close.
func (s *Session) Close() error {
	if s == nil || s.state == nil {
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
