package cpace

import "fmt"

const (
	wireFormatV1 byte = 0xc1
	wireSuite    byte = byte(SuiteCPaceRistretto255SHA512)
	roleA        byte = 0x01
	roleB        byte = 0x02
	roleC        byte = 0x03

	maxPasswordLength       = 4 << 10
	maxIDLength             = 4 << 10
	maxContextLength        = 1 << 10
	maxSessionIDLength      = 1 << 10
	maxAssociatedDataLength = 64 << 10

	// maxLEB128BytesForField is a uniform length-prefix ceiling that must cover
	// maxAssociatedDataLength, the largest package field cap. The caps are
	// per-field, not aggregate message limits.
	maxLEB128BytesForField = 3
)

const _ = uint((1 << (7 * maxLEB128BytesForField)) - 1 - maxAssociatedDataLength)

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

func encodeMessageA(sid, ya, ada []byte) []byte {
	out := []byte{wireFormatV1, wireSuite, roleA}
	out = append(out, prependLen(sid)...)
	out = append(out, prependLen(ya)...)
	out = append(out, prependLen(ada)...)
	return out
}

func encodeMessageB(yb, adb, tag []byte) []byte {
	out := []byte{wireFormatV1, wireSuite, roleB}
	out = append(out, prependLen(yb)...)
	out = append(out, prependLen(adb)...)
	out = append(out, prependLen(tag)...)
	return out
}

func encodeMessageC(tag []byte) []byte {
	out := []byte{wireFormatV1, wireSuite, roleC}
	out = append(out, prependLen(tag)...)
	return out
}

func decodeMessageA(in []byte) (messageA, error) {
	r, err := newMessageReader(in, roleA)
	if err != nil {
		return messageA{}, err
	}
	sid, err := r.readField(maxSessionIDLength, "message A session id")
	if err != nil {
		return messageA{}, err
	}
	ya, err := r.readExactField(pointSize, "message A point")
	if err != nil {
		return messageA{}, err
	}
	ada, err := r.readField(maxAssociatedDataLength, "message A associated data")
	if err != nil {
		return messageA{}, err
	}
	if err := r.done(); err != nil {
		return messageA{}, err
	}
	return messageA{sid: sid, ya: ya, ada: ada}, nil
}

func decodeMessageB(in []byte) (messageB, error) {
	r, err := newMessageReader(in, roleB)
	if err != nil {
		return messageB{}, err
	}
	yb, err := r.readExactField(pointSize, "message B point")
	if err != nil {
		return messageB{}, err
	}
	adb, err := r.readField(maxAssociatedDataLength, "message B associated data")
	if err != nil {
		return messageB{}, err
	}
	tag, err := r.readExactField(tagSize, "message B tag")
	if err != nil {
		return messageB{}, err
	}
	if err := r.done(); err != nil {
		return messageB{}, err
	}
	return messageB{yb: yb, adb: adb, tag: tag}, nil
}

func decodeMessageC(in []byte) (messageC, error) {
	r, err := newMessageReader(in, roleC)
	if err != nil {
		return messageC{}, err
	}
	tag, err := r.readExactField(tagSize, "message C tag")
	if err != nil {
		return messageC{}, err
	}
	if err := r.done(); err != nil {
		return messageC{}, err
	}
	return messageC{tag: tag}, nil
}

type messageReader struct {
	buf []byte
	off int
}

func newMessageReader(in []byte, wantRole byte) (*messageReader, error) {
	if len(in) < 3 {
		return nil, fmt.Errorf("%w: truncated header", ErrMessage)
	}
	if in[0] != wireFormatV1 {
		return nil, fmt.Errorf("%w: wrong wire format version", ErrMessage)
	}
	if in[1] != wireSuite {
		return nil, fmt.Errorf("%w: wrong suite", ErrMessage)
	}
	if in[2] != wantRole {
		return nil, fmt.Errorf("%w: wrong role", ErrMessage)
	}
	return &messageReader{buf: in, off: 3}, nil
}

func (r *messageReader) readField(maxLen int, name string) ([]byte, error) {
	n, err := r.readLEB128()
	if err != nil {
		return nil, err
	}
	if n > maxLen {
		return nil, fmt.Errorf("%w: %s field too large", ErrMessage, name)
	}
	if len(r.buf)-r.off < n {
		return nil, fmt.Errorf("%w: truncated %s field", ErrMessage, name)
	}
	out := clone(r.buf[r.off : r.off+n])
	r.off += n
	return out, nil
}

func (r *messageReader) readExactField(wantLen int, name string) ([]byte, error) {
	n, err := r.readLEB128()
	if err != nil {
		return nil, err
	}
	if n != wantLen {
		return nil, fmt.Errorf("%w: %s length", ErrMessage, name)
	}
	if len(r.buf)-r.off < wantLen {
		return nil, fmt.Errorf("%w: truncated %s field", ErrMessage, name)
	}
	out := clone(r.buf[r.off : r.off+wantLen])
	r.off += wantLen
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
