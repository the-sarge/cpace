package cpace

import "fmt"

const (
	wireFormatV1 byte = 0xc1
	wireSuite    byte = currentSuite
	roleA        byte = 0x01
	roleB        byte = 0x02
	roleC        byte = 0x03

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
	name   string
	role   byte
	fields []messageFieldSpec
}

type messageFieldSpec = packageCapField

var (
	messageASpec = messageSpec{
		name: "A",
		role: roleA,
		fields: []messageFieldSpec{
			messageASessionIDCap,
			messageAPointCap,
			messageAAssociatedDataCap,
		},
	}
	messageBSpec = messageSpec{
		name: "B",
		role: roleB,
		fields: []messageFieldSpec{
			messageBPointCap,
			messageBAssociatedDataCap,
			messageBTagCap,
		},
	}
	messageCSpec = messageSpec{
		name: "C",
		role: roleC,
		fields: []messageFieldSpec{
			messageCTagCap,
		},
	}
)

func messageFramingCatalogue() []messageSpec {
	return []messageSpec{messageASpec, messageBSpec, messageCSpec}
}

func encodeMessageA(sid, ya, ada []byte) []byte {
	return messageASpec.encode(sid, ya, ada)
}

func encodeMessageB(yb, adb, tag []byte) []byte {
	return messageBSpec.encode(yb, adb, tag)
}

func encodeMessageC(tag []byte) []byte {
	return messageCSpec.encode(tag)
}

func (spec messageSpec) encode(fields ...[]byte) []byte {
	if len(fields) != len(spec.fields) {
		panic("cpace: internal message framing field count mismatch")
	}
	capacity := messageHeaderSize
	for _, field := range fields {
		capacity += lengthValueLen(len(field))
	}
	out := make([]byte, 0, capacity)
	out = append(out, wireFormatV1, wireSuite, spec.role)
	for _, field := range fields {
		out = appendLengthValue(out, field)
	}
	return out
}

func decodeMessageA(in []byte) (messageA, error) {
	fields, err := messageASpec.decode(in)
	if err != nil {
		return messageA{}, err
	}
	return messageA{sid: fields[0], ya: fields[1], ada: fields[2]}, nil
}

func decodeMessageB(in []byte) (messageB, error) {
	fields, err := messageBSpec.decode(in)
	if err != nil {
		return messageB{}, err
	}
	return messageB{yb: fields[0], adb: fields[1], tag: fields[2]}, nil
}

func decodeMessageC(in []byte) (messageC, error) {
	fields, err := messageCSpec.decode(in)
	if err != nil {
		return messageC{}, err
	}
	return messageC{tag: fields[0]}, nil
}

func (spec messageSpec) decode(in []byte) ([][]byte, error) {
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
	out, next, err := spec.acceptMessageField(r.buf, r.off, n)
	if err != nil {
		return nil, err
	}
	r.off = next
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
