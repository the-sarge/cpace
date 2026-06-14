package cpace

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"io"
	"runtime"

	"github.com/gtank/ristretto255"
)

const (
	dsiRistretto255 = "CPaceRistretto255"
	dsiISK          = "CPaceRistretto255_ISK"
	sha512BlockSize = 128
	scalarSize      = 32
	pointSize       = 32
	tagSize         = 64
	maxScalarTries  = 128
)

var identityEncoding = make([]byte, pointSize)

func generatorString(dsi, prs, ci, sid []byte, sInBytes int) []byte {
	// The trailing subtraction accounts for the length byte of the zero-padding
	// field. For this draft-21 suite, ZPAD is shorter than 128 bytes, so its
	// LEB128 length prefix is exactly one byte.
	rawZPADLen := sInBytes - len(prependLen(prs)) - len(prependLen(dsi)) - 1
	zpadLen := max(rawZPADLen, 0)
	return lvCat(dsi, prs, make([]byte, zpadLen), ci, sid)
}

func calculateGenerator(prs, ci, sid []byte) *ristretto255.Element {
	genStr := generatorString([]byte(dsiRistretto255), prs, ci, sid, sha512BlockSize)
	hash := sha512.Sum512(genStr)
	clearBytes(genStr)
	g, err := ristretto255.NewIdentityElement().SetUniformBytes(hash[:])
	clearBytes(hash[:])
	if err != nil {
		panic("cpace: SHA-512 output rejected by Ristretto255 SetUniformBytes")
	}
	return g
}

//go:noinline
func clearElement(e *ristretto255.Element) {
	if e == nil {
		return
	}
	e.Set(ristretto255.NewIdentityElement())
	runtime.KeepAlive(e)
}

func sampleScalar(r io.Reader) (*ristretto255.Scalar, error) {
	var b [scalarSize]byte
	defer clearBytes(b[:])
	for range maxScalarTries {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, fmt.Errorf("%w: scalar randomness: %w", ErrRandomness, err)
		}
		b[31] &= 0x0f
		s, err := ristretto255.NewScalar().SetCanonicalBytes(b[:])
		if err != nil {
			// After masking the top four bits the value is below 2^252, but the
			// Ristretto255 scalar order L is only slightly above 2^252, so
			// SetCanonicalBytes can reject in the (~2^-125) window [L, 2^252).
			// Treat that as an unusable sample and retry rather than aborting.
			continue
		}
		if s.Equal(ristretto255.NewScalar().Zero()) == 1 {
			continue
		}
		return s, nil
	}
	return nil, fmt.Errorf("%w: scalar randomness produced only usable-rejection samples", ErrRandomness)
}

//go:noinline
func clearScalar(s *ristretto255.Scalar) {
	if s == nil {
		return
	}
	s.Zero()
	runtime.KeepAlive(s)
}

func scalarFromCanonical(b []byte) (*ristretto255.Scalar, error) {
	if len(b) != scalarSize {
		return nil, fmt.Errorf("%w: scalar length", ErrInvalidInput)
	}
	s, err := ristretto255.NewScalar().SetCanonicalBytes(b)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid scalar", ErrInvalidInput)
	}
	return s, nil
}

func scalarMult(s *ristretto255.Scalar, p *ristretto255.Element) []byte {
	return ristretto255.NewIdentityElement().ScalarMult(s, p).Bytes()
}

func deriveISK(sid, k, transcript []byte) []byte {
	// lvCat fixes the DSI, sid, and K boundaries. The remaining raw transcript
	// is injective for the public initiator-responder flow because transcriptIR
	// has a fixed sequence of length-value fields.
	prefix := lvCat([]byte(dsiISK), sid, k)
	material := make([]byte, len(prefix)+len(transcript))
	copy(material, prefix)
	copy(material[len(prefix):], transcript)
	clearBytes(prefix)
	sum := sha512.Sum512(material)
	clearBytes(material)
	out := make([]byte, sha512.Size)
	copy(out, sum[:])
	clearBytes(sum[:])
	return out
}

func confirmationTag(isk, sid, y, ad []byte) []byte {
	keyInput := append([]byte("CPaceMac"), sid...)
	// Raw concatenation follows draft-21 §10.4. It is collision-free because
	// ISK is always exactly 64 bytes (SHA-512 output), so the boundary between
	// sid and ISK is recoverable as the last 64 bytes. Two different
	// (sid, ISK) pairs collide only if len(sid_1) = len(sid_2) and
	// sid_1 || ISK_1 = sid_2 || ISK_2 — i.e. sid_1 = sid_2 and ISK_1 = ISK_2,
	// which is the same session.
	keyInput = append(keyInput, isk...)
	macKey := sha512.Sum512(keyInput)
	clearBytes(keyInput)
	m := hmac.New(sha512.New, macKey[:])
	clearBytes(macKey[:])
	_, _ = m.Write(lvCat(y, ad))
	return m.Sum(nil)
}

//go:noinline
func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
	runtime.KeepAlive(b)
}
