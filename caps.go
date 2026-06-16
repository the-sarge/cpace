package cpace

import "fmt"

const (
	maxPasswordLength       = 4 << 10
	maxIDLength             = 4 << 10
	maxContextLength        = 1 << 10
	maxSessionIDLength      = 1 << 10
	maxAssociatedDataLength = 64 << 10
)

type packageCapField struct {
	name   string
	length int
	exact  bool
}

func cappedPackageCapField(name string, maxLen int) packageCapField {
	return packageCapField{name: name, length: maxLen}
}

func exactPackageCapField(name string, wantLen int) packageCapField {
	return packageCapField{name: name, length: wantLen, exact: true}
}

type acceptedConfig struct {
	password    []byte
	initiatorID []byte
	responderID []byte
	context     []byte
	sid         []byte
	ad          []byte
}

func acceptConfig(cfg Config) (acceptedConfig, error) {
	if len(cfg.Password) == 0 {
		return acceptedConfig{}, fmt.Errorf("%w: empty password", ErrInvalidInput)
	}
	if len(cfg.InitiatorID) == 0 {
		return acceptedConfig{}, fmt.Errorf("%w: empty initiator id", ErrInvalidInput)
	}
	if len(cfg.ResponderID) == 0 {
		return acceptedConfig{}, fmt.Errorf("%w: empty responder id", ErrInvalidInput)
	}
	if len(cfg.SessionID) == 0 && !cfg.AllowEmptySessionID {
		return acceptedConfig{}, fmt.Errorf("%w: %w", ErrInvalidInput, ErrEmptySessionID)
	}

	for _, field := range []struct {
		cap   packageCapField
		value []byte
	}{
		{passwordCap, cfg.Password},
		{initiatorIDCap, cfg.InitiatorID},
		{responderIDCap, cfg.ResponderID},
		{contextCap, cfg.Context},
		{sessionIDCap, cfg.SessionID},
		{associatedDataCap, cfg.AssociatedData},
	} {
		if err := field.cap.validateConfigLength(len(field.value)); err != nil {
			return acceptedConfig{}, err
		}
	}

	return acceptedConfig{
		password:    clone(cfg.Password),
		initiatorID: clone(cfg.InitiatorID),
		responderID: clone(cfg.ResponderID),
		context:     clone(cfg.Context),
		sid:         clone(cfg.SessionID),
		ad:          clone(cfg.AssociatedData),
	}, nil
}

func (c *acceptedConfig) wipe() {
	if c == nil {
		return
	}
	clearBytes(c.password)
	clearBytes(c.initiatorID)
	clearBytes(c.responderID)
	clearBytes(c.context)
	clearBytes(c.sid)
	clearBytes(c.ad)
}

func (f packageCapField) validateConfigLength(n int) error {
	if n > f.length {
		return fmt.Errorf("%w: %s too large", ErrInvalidInput, f.name)
	}
	return nil
}

func (f packageCapField) acceptMessageField(buf []byte, off, n int) ([]byte, int, error) {
	if err := f.validateMessageLength(n); err != nil {
		return nil, off, err
	}
	if len(buf)-off < n {
		return nil, off, fmt.Errorf("%w: truncated %s field", ErrMessage, f.name)
	}
	out := clone(buf[off : off+n])
	return out, off + n, nil
}

func (f packageCapField) validateMessageLength(n int) error {
	if f.exact {
		if n != f.length {
			return fmt.Errorf("%w: %s length", ErrMessage, f.name)
		}
	} else if n > f.length {
		return fmt.Errorf("%w: %s field too large", ErrMessage, f.name)
	}
	return nil
}

func shippedPackageCapPolicy() []packageCapField {
	return []packageCapField{
		passwordCap,
		initiatorIDCap,
		responderIDCap,
		contextCap,
		sessionIDCap,
		associatedDataCap,
		messageASessionIDCap,
		messageAPointCap,
		messageAAssociatedDataCap,
		messageBPointCap,
		messageBAssociatedDataCap,
		messageBTagCap,
		messageCTagCap,
	}
}

var (
	passwordCap       = cappedPackageCapField("password", maxPasswordLength)
	initiatorIDCap    = cappedPackageCapField("initiator id", maxIDLength)
	responderIDCap    = cappedPackageCapField("responder id", maxIDLength)
	contextCap        = cappedPackageCapField("context", maxContextLength)
	sessionIDCap      = cappedPackageCapField("session id", maxSessionIDLength)
	associatedDataCap = cappedPackageCapField("associated data", maxAssociatedDataLength)

	messageASessionIDCap      = cappedPackageCapField("message A session id", maxSessionIDLength)
	messageAPointCap          = exactPackageCapField("message A point", pointSize)
	messageAAssociatedDataCap = cappedPackageCapField("message A associated data", maxAssociatedDataLength)
	messageBPointCap          = exactPackageCapField("message B point", pointSize)
	messageBAssociatedDataCap = cappedPackageCapField("message B associated data", maxAssociatedDataLength)
	messageBTagCap            = exactPackageCapField("message B tag", tagSize)
	messageCTagCap            = exactPackageCapField("message C tag", tagSize)
)
