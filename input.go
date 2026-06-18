package cpace

import "fmt"

// Input contains the local inputs for one CPace role.
//
// Password, SelfID, and PeerID must be non-empty. Context and
// LocalAssociatedData may be empty. Password, Context, and SessionID are shared
// session values both parties supply identically. SelfID, PeerID, and
// LocalAssociatedData are role-local values: Start treats SelfID as the
// initiator identity and PeerID as the responder identity; Respond treats
// SelfID as the responder identity and PeerID as the initiator identity.
// SessionID must be a fresh, non-secret, parties-agree-on value for every
// session. Empty SessionID values are rejected by default because they weaken
// replay and transcript separation properties. Set AllowEmptySessionID only for
// draft-21 compatibility tests or profiles that have deliberately accepted the
// weaker empty-sid behavior. Scalar randomness always comes from
// crypto/rand.Reader. Field lengths are capped at 4 KiB for Password and IDs,
// 1 KiB for Context and SessionID, and 64 KiB for LocalAssociatedData. Inputs
// exceeding these caps are rejected before copying; accepted byte slices are
// copied by Start and Respond before use.
type Input struct {
	Password            []byte
	SelfID              []byte
	PeerID              []byte
	Context             []byte
	SessionID           []byte
	LocalAssociatedData []byte
	AllowEmptySessionID bool
}

type callerInput struct {
	password []byte
	selfID   []byte
	peerID   []byte
	context  []byte
	sid      []byte
	localAD  []byte
}

type callerInputRequiredField struct {
	name  string
	value []byte
}

type callerInputCappedField struct {
	cap   packageCapField
	value []byte
}

func acceptInput(input Input) (callerInput, error) {
	if err := validateRequiredCallerInputFields(input); err != nil {
		return callerInput{}, err
	}
	if len(input.SessionID) == 0 && !input.AllowEmptySessionID {
		return callerInput{}, fmt.Errorf("%w: %w", ErrInvalidInput, ErrEmptySessionID)
	}
	if err := validateCallerInputCapFields(input); err != nil {
		return callerInput{}, err
	}

	return callerInput{
		password: clone(input.Password),
		selfID:   clone(input.SelfID),
		peerID:   clone(input.PeerID),
		context:  clone(input.Context),
		sid:      clone(input.SessionID),
		localAD:  clone(input.LocalAssociatedData),
	}, nil
}

func validateRequiredCallerInputFields(input Input) error {
	for _, field := range [...]callerInputRequiredField{
		{name: "password", value: input.Password},
		{name: "self id", value: input.SelfID},
		{name: "peer id", value: input.PeerID},
	} {
		if len(field.value) == 0 {
			return fmt.Errorf("%w: empty %s", ErrInvalidInput, field.name)
		}
	}
	return nil
}

func validateCallerInputCapFields(input Input) error {
	for _, field := range [...]callerInputCappedField{
		{cap: passwordCap, value: input.Password},
		{cap: selfIDCap, value: input.SelfID},
		{cap: peerIDCap, value: input.PeerID},
		{cap: contextCap, value: input.Context},
		{cap: sessionIDCap, value: input.SessionID},
		{cap: localAssociatedDataCap, value: input.LocalAssociatedData},
	} {
		if err := field.cap.validateInputLength(len(field.value)); err != nil {
			return err
		}
	}
	return nil
}

func (c *callerInput) wipe() {
	if c == nil {
		return
	}
	clearBytes(c.password)
	clearBytes(c.selfID)
	clearBytes(c.peerID)
	clearBytes(c.context)
	clearBytes(c.sid)
	clearBytes(c.localAD)
}

// handoff transfers package-owned caller input slices into normalizedInput.
// The callerInput keeps no references to transferred backing arrays, leaving
// normalizedInput.wipe as the next lifetime boundary.
func (c *callerInput) handoff(role callerInputRole) normalizedInput {
	initiatorID, responderID := role.mapTranscriptIDs(*c)
	ci := buildCI(initiatorID, responderID, c.context)
	clearBytes(c.context)
	c.context = nil

	ni := normalizedInput{
		password:    c.password,
		initiatorID: initiatorID,
		responderID: responderID,
		ci:          ci,
		sid:         c.sid,
		ad:          c.localAD,
	}
	c.password = nil
	c.selfID = nil
	c.peerID = nil
	c.sid = nil
	c.localAD = nil
	return ni
}

type normalizedInput struct {
	password    []byte
	initiatorID []byte
	responderID []byte
	ci          []byte
	sid         []byte
	ad          []byte
}

type callerInputRole byte

const (
	initiatorInputRole callerInputRole = iota
	responderInputRole
)

// wipe performs best-effort zeroization of every byte slice owned by the
// normalized input. Called via defer in startWithRandom and respondWithRandom
// so that all cloned input bytes are cleared on every exit path — including
// core-constructor error returns and panics. Idempotent against fields whose
// backing arrays were already zeroed behind the core seam (the password is
// eagerly cleared inside the core constructors).
func (ni *normalizedInput) wipe() {
	clearBytes(ni.password)
	clearBytes(ni.initiatorID)
	clearBytes(ni.responderID)
	clearBytes(ni.ci)
	clearBytes(ni.sid)
	clearBytes(ni.ad)
}

func normalizeStartInput(input Input) (normalizedInput, error) {
	return normalizeInput(input, initiatorInputRole)
}

func normalizeRespondInput(input Input) (normalizedInput, error) {
	return normalizeInput(input, responderInputRole)
}

func normalizeInput(input Input, role callerInputRole) (normalizedInput, error) {
	caller, err := acceptInput(input)
	if err != nil {
		return normalizedInput{}, err
	}
	defer caller.wipe()
	return caller.handoff(role), nil
}

func (r callerInputRole) mapTranscriptIDs(input callerInput) (initiatorID, responderID []byte) {
	switch r {
	case initiatorInputRole:
		return input.selfID, input.peerID
	case responderInputRole:
		return input.peerID, input.selfID
	default:
		panic("cpace: internal caller input role mismatch")
	}
}

func buildCI(initiatorID, responderID, context []byte) []byte {
	return lvCat(
		[]byte("CPace-Go-CI"),
		[]byte(DraftVersion),
		[]byte(suiteName),
		[]byte("initiator"),
		initiatorID,
		[]byte("responder"),
		responderID,
		[]byte("context"),
		context,
	)
}
