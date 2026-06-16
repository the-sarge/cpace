package cpace

import "bytes"

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
