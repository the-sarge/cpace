package cpace

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
