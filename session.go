package cpace

import (
	"crypto/hkdf"
	"crypto/sha512"
	"fmt"
)

const maxHKDFOutput = 255 * 64

// TranscriptID returns the draft CPaceSidOutput value for the confirmed
// initiator-responder transcript.
func (s *Session) TranscriptID() []byte {
	if s == nil {
		return nil
	}
	return clone(s.transcriptID)
}

// Export derives application key material from the confirmed ISK using
// HKDF-SHA512. The label and context are prefix-free encoded into HKDF info.
func (s *Session) Export(label, context []byte, length int) ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: nil session", ErrInvalidInput)
	}
	if length < 0 || length > maxHKDFOutput {
		return nil, fmt.Errorf("%w: invalid export length", ErrInvalidInput)
	}
	info := lvCat([]byte("CPaceExport"), label, context)
	out, err := hkdf.Key(sha512.New, s.isk, nil, string(info), length)
	if err != nil {
		return nil, fmt.Errorf("%w: export failed: %v", ErrInvalidInput, err)
	}
	return out, nil
}
