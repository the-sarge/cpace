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

func (f packageCapField) validateInputLength(n int) error {
	if n > f.length {
		return fmt.Errorf("%w: %s too large", ErrInvalidInput, f.name)
	}
	return nil
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
		selfIDCap,
		peerIDCap,
		contextCap,
		sessionIDCap,
		localAssociatedDataCap,
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
	passwordCap            = cappedPackageCapField("password", maxPasswordLength)
	selfIDCap              = cappedPackageCapField("self id", maxIDLength)
	peerIDCap              = cappedPackageCapField("peer id", maxIDLength)
	contextCap             = cappedPackageCapField("context", maxContextLength)
	sessionIDCap           = cappedPackageCapField("session id", maxSessionIDLength)
	localAssociatedDataCap = cappedPackageCapField("local associated data", maxAssociatedDataLength)

	messageASessionIDCap      = cappedPackageCapField("message A session id", maxSessionIDLength)
	messageAPointCap          = exactPackageCapField("message A point", pointSize)
	messageAAssociatedDataCap = cappedPackageCapField("message A associated data", maxAssociatedDataLength)
	messageBPointCap          = exactPackageCapField("message B point", pointSize)
	messageBAssociatedDataCap = cappedPackageCapField("message B associated data", maxAssociatedDataLength)
	messageBTagCap            = exactPackageCapField("message B tag", tagSize)
	messageCTagCap            = exactPackageCapField("message C tag", tagSize)
)
