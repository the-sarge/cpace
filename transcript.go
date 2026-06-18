package cpace

import "crypto/sha512"

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

func (t irTranscript) transcriptID() []byte {
	return transcriptID(t.transcript)
}

func (t irTranscript) deriveISK(sid, k []byte) []byte {
	return deriveISK(sid, k, t.transcript)
}

// initiatorAD returns a copy of the initiator's associated data bound into the
// transcript. The responder uses it to populate the confirmed Session's peer
// associated data without retaining a decomposed field of its own.
func (t irTranscript) initiatorAD() []byte {
	return clone(t.ada)
}

// clear zeroes then nils the transcript's public byte fields. The responder
// calls it from responderCore.clear so the stored transcript is wiped
// alongside the ISK as hygiene (ADR-0001); the transcript holds no secret of
// its own. Safe to call more than once and on a nil receiver.
func (t *irTranscript) clear() {
	if t == nil {
		return
	}
	clearBytes(t.transcript)
	clearBytes(t.ya)
	clearBytes(t.ada)
	clearBytes(t.yb)
	clearBytes(t.adb)
	t.transcript = nil
	t.ya = nil
	t.ada = nil
	t.yb = nil
	t.adb = nil
}

func (t irTranscript) initiatorConfirmationTag(isk, sid []byte) []byte {
	return confirmationTag(isk, sid, t.ya, t.ada)
}

func (t irTranscript) responderConfirmationTag(isk, sid []byte) []byte {
	return confirmationTag(isk, sid, t.yb, t.adb)
}

func transcriptID(transcript []byte) []byte {
	h := sha512.New()
	_, _ = h.Write([]byte("CPaceSidOutput"))
	_, _ = h.Write(transcript)
	return h.Sum(nil)
}
