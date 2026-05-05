package cpace

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"io"

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
	zpadLen := sInBytes - len(prependLen(prs)) - len(prependLen(dsi)) - 1
	if zpadLen < 0 {
		zpadLen = 0
	}
	return lvCat(dsi, prs, make([]byte, zpadLen), ci, sid)
}

func calculateGenerator(prs, ci, sid []byte) *ristretto255.Element {
	genStr := generatorString([]byte(dsiRistretto255), prs, ci, sid, sha512BlockSize)
	hash := sha512.Sum512(genStr)
	g, err := ristretto255.NewIdentityElement().SetUniformBytes(hash[:])
	if err != nil {
		panic("cpace: SHA-512 output rejected by Ristretto255 SetUniformBytes")
	}
	return g
}

func sampleScalar(r io.Reader) (*ristretto255.Scalar, error) {
	var b [scalarSize]byte
	for attempts := 0; attempts < maxScalarTries; attempts++ {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, fmt.Errorf("%w: scalar randomness: %w", ErrRandomness, err)
		}
		b[31] &= 0x0f
		s, err := ristretto255.NewScalar().SetCanonicalBytes(b[:])
		if err != nil {
			return nil, fmt.Errorf("%w: scalar sampling rejected masked bytes: %w", ErrRandomness, err)
		}
		if s.Equal(ristretto255.NewScalar().Zero()) == 1 {
			continue
		}
		return s, nil
	}
	return nil, fmt.Errorf("%w: scalar randomness produced only zero scalars", ErrRandomness)
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

func scalarMultVFY(s *ristretto255.Scalar, encoded []byte) ([]byte, bool) {
	// Defensive for internal callers; public message decoders enforce pointSize.
	if len(encoded) != pointSize {
		return clone(identityEncoding), false
	}
	p, err := ristretto255.NewIdentityElement().SetCanonicalBytes(encoded)
	if err != nil {
		return clone(identityEncoding), false
	}
	out := ristretto255.NewIdentityElement().ScalarMult(s, p).Bytes()
	if hmac.Equal(out, identityEncoding) {
		return clone(identityEncoding), false
	}
	return out, true
}

func deriveISK(sid, k, transcript []byte) []byte {
	material := lvCat([]byte(dsiISK), sid, k)
	material = append(material, transcript...)
	sum := sha512.Sum512(material)
	return sum[:]
}

func confirmationTag(isk, sid, y, ad []byte) []byte {
	keyInput := append([]byte("CPaceMac"), sid...)
	// Raw concatenation follows draft-21. This is unambiguous in this suite
	// because ISK is fixed at SHA-512's 64-byte output length.
	keyInput = append(keyInput, isk...)
	macKey := sha512.Sum512(keyInput)
	m := hmac.New(sha512.New, macKey[:])
	_, _ = m.Write(lvCat(y, ad))
	return m.Sum(nil)
}
