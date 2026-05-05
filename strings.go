package cpace

import "bytes"

func prependLen(data []byte) []byte {
	out := encodeLEB128(uint64(len(data)))
	out = append(out, data...)
	return out
}

func lvCat(args ...[]byte) []byte {
	var total int
	for _, arg := range args {
		total += len(encodeLEB128(uint64(len(arg)))) + len(arg)
	}
	out := make([]byte, 0, total)
	for _, arg := range args {
		out = append(out, prependLen(arg)...)
	}
	return out
}

func oCat(a, b []byte) []byte {
	out := []byte("oc")
	if lexicographicallyLarger(a, b) {
		out = append(out, a...)
		out = append(out, b...)
		return out
	}
	out = append(out, b...)
	out = append(out, a...)
	return out
}

func transcriptIR(ya, ada, yb, adb []byte) []byte {
	out := lvCat(ya, ada)
	out = append(out, lvCat(yb, adb)...)
	return out
}

func transcriptOC(ya, ada, yb, adb []byte) []byte {
	return oCat(lvCat(ya, ada), lvCat(yb, adb))
}

func lexicographicallyLarger(a, b []byte) bool {
	return bytes.Compare(a, b) > 0
}

func encodeLEB128(n uint64) []byte {
	out := make([]byte, 0, 10)
	for {
		b := byte(n & 0x7f)
		n >>= 7
		if n != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if n == 0 {
			return out
		}
	}
}
