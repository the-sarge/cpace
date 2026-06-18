package cpace

type irTranscript struct {
	transcript []byte
	ya         []byte
	ada        []byte
	yb         []byte
	adb        []byte
}

func newIRTranscript(ya, ada, yb, adb []byte) irTranscript {
	ya = clone(ya)
	ada = clone(ada)
	yb = clone(yb)
	adb = clone(adb)
	transcript := lvCat(ya, ada)
	transcript = append(transcript, lvCat(yb, adb)...)
	return irTranscript{
		transcript: transcript,
		ya:         ya,
		ada:        ada,
		yb:         yb,
		adb:        adb,
	}
}

func (t irTranscript) bytes() []byte {
	return clone(t.transcript)
}

func (t irTranscript) deriveISK(sid, k []byte) []byte {
	return deriveISK(sid, k, t.transcript)
}

func (t irTranscript) initiatorConfirmationTag(isk, sid []byte) []byte {
	return initiatorRoleConfirmationTag(isk, sid, t.ya, t.ada)
}

func (t irTranscript) responderConfirmationTag(isk, sid []byte) []byte {
	return responderRoleConfirmationTag(isk, sid, t.yb, t.adb)
}

func initiatorRoleConfirmationTag(isk, sid, ya, ada []byte) []byte {
	return confirmationTag(isk, sid, ya, ada)
}

func responderRoleConfirmationTag(isk, sid, yb, adb []byte) []byte {
	return confirmationTag(isk, sid, yb, adb)
}
