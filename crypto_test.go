package cpace

import (
	"fmt"

	"github.com/gtank/ristretto255"
)

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
