package cpace

import "errors"

var (
	// ErrInvalidInput reports invalid local configuration or parameters.
	ErrInvalidInput = errors.New("cpace: invalid input")

	// ErrMessage reports malformed or unexpected wire messages.
	ErrMessage = errors.New("cpace: invalid message")

	// ErrStateUsed reports an attempt to reuse a single-use protocol state.
	ErrStateUsed = errors.New("cpace: state already used")

	// ErrAbort reports a draft abort condition such as an invalid point or
	// neutral-element Diffie-Hellman result.
	ErrAbort = errors.New("cpace: protocol abort")

	// ErrConfirmationFailed reports failed explicit key confirmation.
	ErrConfirmationFailed = errors.New("cpace: key confirmation failed")
)
