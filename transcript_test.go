package cpace

func (t *irTranscript) bytes() []byte {
	return clone(t.transcript)
}
