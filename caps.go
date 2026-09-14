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

func newCappedPackageCapField(name string, maxLen int) packageCapField {
	return packageCapField{name: name, length: maxLen}
}

func newExactPackageCapField(name string, wantLen int) packageCapField {
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

var (
	passwordCap            = newCappedPackageCapField("password", maxPasswordLength)
	selfIDCap              = newCappedPackageCapField("self id", maxIDLength)
	peerIDCap              = newCappedPackageCapField("peer id", maxIDLength)
	contextCap             = newCappedPackageCapField("context", maxContextLength)
	sessionIDCap           = newCappedPackageCapField("session id", maxSessionIDLength)
	localAssociatedDataCap = newCappedPackageCapField("local associated data", maxAssociatedDataLength)

	messageASessionIDCap      = newCappedPackageCapField("message A session id", maxSessionIDLength)
	messageAPointCap          = newExactPackageCapField("message A point", pointSize)
	messageAAssociatedDataCap = newCappedPackageCapField("message A associated data", maxAssociatedDataLength)
	messageBPointCap          = newExactPackageCapField("message B point", pointSize)
	messageBAssociatedDataCap = newCappedPackageCapField("message B associated data", maxAssociatedDataLength)
	messageBTagCap            = newExactPackageCapField("message B tag", tagSize)
	messageCTagCap            = newExactPackageCapField("message C tag", tagSize)
)
