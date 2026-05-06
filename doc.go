// Package cpace implements an auditable draft-irtf-cfrg-cpace-21
// CPACE-RISTR255-SHA512 initiator-responder flow.
//
// This module is an Internet-Draft implementation. It is not independently
// audited and must not be treated as production-ready cryptographic software.
//
// The public API intentionally exposes only an initiator-responder flow with
// mandatory explicit key confirmation. A session is returned only after both
// sides have confirmed possession of the same intermediate session key.
// Respond success alone is not authentication.
//
// Scalar randomness always comes from crypto/rand.Reader; the package does not
// accept caller-supplied randomness through the public API.
//
// Both parties must use the same role orientation: InitiatorID identifies the
// Start side, and ResponderID identifies the Respond side. Applications that
// negotiate PAKE versions, suites, or protocol modes outside this package must
// provide their own downgrade protection for that outer negotiation.
package cpace
