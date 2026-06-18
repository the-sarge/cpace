package cpace

import "fmt"

func prependLen(data []byte) []byte {
	return appendLengthValue(nil, data)
}

func lvCat(args ...[]byte) []byte {
	var total int
	for _, arg := range args {
		total += lengthValueLen(len(arg))
	}
	out := make([]byte, 0, total)
	for _, arg := range args {
		out = appendLengthValue(out, arg)
	}
	return out
}

func appendLengthValue(dst, data []byte) []byte {
	dst = appendLEB128(dst, uint64(len(data)))
	return append(dst, data...)
}

func lengthValueLen(n int) int {
	return leb128LenInt(n) + n
}

func appendLEB128(dst []byte, n uint64) []byte {
	for {
		b := byte(n & 0x7f)
		n >>= 7
		if n != 0 {
			b |= 0x80
		}
		dst = append(dst, b)
		if n == 0 {
			return dst
		}
	}
}

func readLEB128(buf []byte, off, maxBytes int) (int, int, error) {
	var n int
	for i := range maxBytes {
		if off >= len(buf) {
			return 0, off, fmt.Errorf("%w: truncated LEB128", ErrMessage)
		}
		b := buf[off]
		off++
		n |= int(b&0x7f) << (7 * i)
		if b&0x80 == 0 {
			if i > 0 && n < 1<<(7*i) {
				return 0, off, fmt.Errorf("%w: non-canonical LEB128", ErrMessage)
			}
			return n, off, nil
		}
	}
	return 0, off, fmt.Errorf("%w: malformed LEB128", ErrMessage)
}

func leb128LenInt(n int) int {
	out := 1
	for n >= 0x80 {
		n >>= 7
		out++
	}
	return out
}
