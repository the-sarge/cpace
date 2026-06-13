package cpace

import "fmt"

const (
	wireFormatV1 byte = 0xc1
	wireSuite    byte = currentSuite
	roleA        byte = 0x01
	roleB        byte = 0x02
	roleC        byte = 0x03

	maxPasswordLength       = 4 << 10
	maxIDLength             = 4 << 10
	maxContextLength        = 1 << 10
	maxSessionIDLength      = 1 << 10
	maxAssociatedDataLength = 64 << 10

	// maxMessageLength is an aggregate invalid-message cap. It is intentionally
	// above every valid package-owned Message framing shape, so valid wire bytes
	// remain governed by the per-field caps below.
	maxMessageLength = 128 << 10

	// maxLEB128BytesForField is a uniform length-prefix ceiling that must cover
	// maxAssociatedDataLength, the largest package field cap. The caps are
	// per-field, not aggregate message limits.
	maxLEB128BytesForField = 3
)

const _ = uint((1 << (7 * maxLEB128BytesForField)) - 1 - maxAssociatedDataLength)

const (
	messageHeaderSize      = 3
	maxValidMessageALength = messageHeaderSize +
		maxLEB128BytesForField + maxSessionIDLength +
		maxLEB128BytesForField + pointSize +
		maxLEB128BytesForField + maxAssociatedDataLength
	maxValidMessageBLength = messageHeaderSize +
		maxLEB128BytesForField + pointSize +
		maxLEB128BytesForField + maxAssociatedDataLength +
		maxLEB128BytesForField + tagSize
	maxValidMessageCLength = messageHeaderSize +
		maxLEB128BytesForField + tagSize
)

const (
	_ = uint(maxMessageLength - maxValidMessageALength - 1)
	_ = uint(maxMessageLength - maxValidMessageBLength - 1)
	_ = uint(maxMessageLength - maxValidMessageCLength - 1)
)

type messageA struct {
	sid []byte
	ya  []byte
	ada []byte
}

type messageB struct {
	yb  []byte
	adb []byte
	tag []byte
}

type messageC struct {
	tag []byte
}

type messageSpec struct {
	role   byte
	fields []messageFieldSpec
}

type messageFieldSpec struct {
	name   string
	length int
	exact  bool
}

var (
	messageASpec = messageSpec{
		role: roleA,
		fields: []messageFieldSpec{
			cappedField("message A session id", maxSessionIDLength),
			exactField("message A point", pointSize),
			cappedField("message A associated data", maxAssociatedDataLength),
		},
	}
	messageBSpec = messageSpec{
		role: roleB,
		fields: []messageFieldSpec{
			exactField("message B point", pointSize),
			cappedField("message B associated data", maxAssociatedDataLength),
			exactField("message B tag", tagSize),
		},
	}
	messageCSpec = messageSpec{
		role: roleC,
		fields: []messageFieldSpec{
			exactField("message C tag", tagSize),
		},
	}
)

func cappedField(name string, maxLen int) messageFieldSpec {
	return messageFieldSpec{name: name, length: maxLen}
}

func exactField(name string, wantLen int) messageFieldSpec {
	return messageFieldSpec{name: name, length: wantLen, exact: true}
}

func encodeMessageA(sid, ya, ada []byte) []byte {
	return encodeMessage(messageASpec, sid, ya, ada)
}

func encodeMessageB(yb, adb, tag []byte) []byte {
	return encodeMessage(messageBSpec, yb, adb, tag)
}

func encodeMessageC(tag []byte) []byte {
	return encodeMessage(messageCSpec, tag)
}

func encodeMessage(spec messageSpec, fields ...[]byte) []byte {
	if len(fields) != len(spec.fields) {
		panic("cpace: internal message framing field count mismatch")
	}
	out := []byte{wireFormatV1, wireSuite, spec.role}
	for _, field := range fields {
		out = append(out, prependLen(field)...)
	}
	return out
}

func decodeMessageA(in []byte) (messageA, error) {
	fields, err := decodeMessage(messageASpec, in)
	if err != nil {
		return messageA{}, err
	}
	return messageA{sid: fields[0], ya: fields[1], ada: fields[2]}, nil
}

func decodeMessageB(in []byte) (messageB, error) {
	fields, err := decodeMessage(messageBSpec, in)
	if err != nil {
		return messageB{}, err
	}
	return messageB{yb: fields[0], adb: fields[1], tag: fields[2]}, nil
}

func decodeMessageC(in []byte) (messageC, error) {
	fields, err := decodeMessage(messageCSpec, in)
	if err != nil {
		return messageC{}, err
	}
	return messageC{tag: fields[0]}, nil
}

func decodeMessage(spec messageSpec, in []byte) ([][]byte, error) {
	r, err := newMessageReader(in, spec)
	if err != nil {
		return nil, err
	}
	fields := make([][]byte, len(spec.fields))
	for i, field := range spec.fields {
		fields[i], err = r.readField(field)
		if err != nil {
			return nil, err
		}
	}
	if err := r.done(); err != nil {
		return nil, err
	}
	return fields, nil
}

type messageReader struct {
	buf []byte
	off int
}

func newMessageReader(in []byte, spec messageSpec) (*messageReader, error) {
	if len(in) < 3 {
		return nil, fmt.Errorf("%w: truncated header", ErrMessage)
	}
	if in[0] != wireFormatV1 {
		return nil, fmt.Errorf("%w: wrong wire format version", ErrMessage)
	}
	if in[1] != wireSuite {
		return nil, fmt.Errorf("%w: wrong suite", ErrMessage)
	}
	if in[2] != spec.role {
		return nil, fmt.Errorf("%w: wrong role", ErrMessage)
	}
	if len(in) > maxMessageLength {
		return nil, fmt.Errorf("%w: message too large", ErrMessage)
	}
	return &messageReader{buf: in, off: 3}, nil
}

func (r *messageReader) readField(spec messageFieldSpec) ([]byte, error) {
	n, err := r.readLEB128()
	if err != nil {
		return nil, err
	}
	if spec.exact {
		if n != spec.length {
			return nil, fmt.Errorf("%w: %s length", ErrMessage, spec.name)
		}
	} else if n > spec.length {
		return nil, fmt.Errorf("%w: %s field too large", ErrMessage, spec.name)
	}
	if len(r.buf)-r.off < n {
		return nil, fmt.Errorf("%w: truncated %s field", ErrMessage, spec.name)
	}
	out := clone(r.buf[r.off : r.off+n])
	r.off += n
	return out, nil
}

func (r *messageReader) readLEB128() (int, error) {
	var n int
	for i := range int(maxLEB128BytesForField) {
		if r.off >= len(r.buf) {
			return 0, fmt.Errorf("%w: truncated LEB128", ErrMessage)
		}
		b := r.buf[r.off]
		r.off++
		n |= int(b&0x7f) << (7 * i)
		if b&0x80 == 0 {
			if i > 0 && n < 1<<(7*i) {
				return 0, fmt.Errorf("%w: non-canonical LEB128", ErrMessage)
			}
			return n, nil
		}
	}
	return 0, fmt.Errorf("%w: malformed LEB128", ErrMessage)
}

func (r *messageReader) done() error {
	if r.off != len(r.buf) {
		return fmt.Errorf("%w: trailing data", ErrMessage)
	}
	return nil
}
