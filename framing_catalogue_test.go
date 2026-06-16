package cpace

import (
	"bytes"
	"fmt"
)

type messageFramingTarget struct {
	name   string
	role   byte
	valid  func() []byte
	decode func([]byte) error
}

type messageFramingCase struct {
	name            string
	msg             []byte
	wantErrContains string
}

func messageFramingTargets() []messageFramingTarget {
	return []messageFramingTarget{
		{
			name:  "A",
			role:  roleA,
			valid: validMessageAForCatalogue,
			decode: func(msg []byte) error {
				_, err := decodeMessageA(msg)
				return err
			},
		},
		{
			name:  "B",
			role:  roleB,
			valid: validMessageBForCatalogue,
			decode: func(msg []byte) error {
				_, err := decodeMessageB(msg)
				return err
			},
		},
		{
			name:  "C",
			role:  roleC,
			valid: validMessageCForCatalogue,
			decode: func(msg []byte) error {
				_, err := decodeMessageC(msg)
				return err
			},
		},
	}
}

func validMessageAForCatalogue() []byte {
	return encodeMessageA([]byte("sid"), bytes.Repeat([]byte{0x42}, messageAPointCap.length), []byte("ADa"))
}

func validMessageBForCatalogue() []byte {
	return encodeMessageB(bytes.Repeat([]byte{0x42}, messageBPointCap.length), []byte("ADb"), bytes.Repeat([]byte{0x99}, messageBTagCap.length))
}

func validMessageCForCatalogue() []byte {
	return encodeMessageC(bytes.Repeat([]byte{0x99}, messageCTagCap.length))
}

func messageFramingMalformedCases(target messageFramingTarget) []messageFramingCase {
	valid := target.valid()
	cases := []messageFramingCase{
		{"truncated field", truncatedMessage(valid), messageFramingFinalFieldTruncation(target.role)},
		{"trailing data", append(clone(valid), 0), "trailing data"},
		{"wrong format", withMessageHeader(valid, 0, wireSuite, target.role), "wrong wire format version"},
		{"wrong suite", withMessageHeader(valid, wireFormatV1, 0, target.role), "wrong suite"},
		{"wrong role", withMessageHeader(valid, wireFormatV1, wireSuite, otherMessageRole(target.role)), "wrong role"},
		{"swapped format suite", withMessageHeader(valid, wireSuite, wireFormatV1, target.role), "wrong wire format version"},
	}
	cases = append(cases, messageFramingLEB128Cases(target.role)...)
	return cases
}

func messageFramingLEB128Cases(role byte) []messageFramingCase {
	return []messageFramingCase{
		{"missing length prefix", messageHeader(role), "truncated LEB128"},
		{"truncated LEB128 continuation", append(messageHeader(role), 0x80), "truncated LEB128"},
		{"non-canonical zero LEB128", append(messageHeader(role), 0x80, 0x00), "non-canonical LEB128"},
		{"non-canonical value LEB128", append(append(messageHeader(role), 0xc0, 0x00), bytes.Repeat([]byte{0x99}, messageCTagCap.length)...), "non-canonical LEB128"},
		{"malformed LEB128", append(messageHeader(role), 0x80, 0x80, 0x80, 0x80, 0x00), "malformed LEB128"},
	}
}

func messageFramingFinalFieldTruncation(role byte) string {
	switch role {
	case roleA:
		return "truncated message A associated data field"
	case roleB:
		return "truncated message B tag field"
	case roleC:
		return "truncated message C tag field"
	default:
		return "truncated"
	}
}

func messageFramingAggregateCases(target messageFramingTarget) []messageFramingCase {
	tooLarge := append(messageHeader(target.role), bytes.Repeat([]byte{0x80}, maxMessageLength)...)
	return []messageFramingCase{
		{"too large", tooLarge, "message too large"},
		{"wrong suite before size", withMessageHeader(tooLarge, wireFormatV1, 0, target.role), "wrong suite"},
		{"wrong role before size", withMessageHeader(tooLarge, wireFormatV1, wireSuite, otherMessageRole(target.role)), "wrong role"},
	}
}

func messageFramingMaxFieldCases() []messageFramingCase {
	aPoint := bytes.Repeat([]byte{0x42}, messageAPointCap.length)
	bPoint := bytes.Repeat([]byte{0x42}, messageBPointCap.length)
	bTag := bytes.Repeat([]byte{0x99}, messageBTagCap.length)
	cTag := bytes.Repeat([]byte{0x99}, messageCTagCap.length)
	return []messageFramingCase{
		{"A max fields", encodeMessageA(bytes.Repeat([]byte{0x11}, messageASessionIDCap.length), aPoint, bytes.Repeat([]byte{0x22}, messageAAssociatedDataCap.length)), ""},
		{"B max fields", encodeMessageB(bPoint, bytes.Repeat([]byte{0x33}, messageBAssociatedDataCap.length), bTag), ""},
		{"C exact tag", encodeMessageC(cTag), ""},
	}
}

func messageFramingFieldLimitCases() []messageFramingCase {
	aPoint := bytes.Repeat([]byte{0x42}, messageAPointCap.length)
	bPoint := bytes.Repeat([]byte{0x42}, messageBPointCap.length)
	bTag := bytes.Repeat([]byte{0x99}, messageBTagCap.length)
	overDeclaredBAd := append(messageHeader(roleB), prependLen(bPoint)...)
	overDeclaredBAd = append(overDeclaredBAd, appendLEB128(nil, maxAssociatedDataLength+1)...)
	return []messageFramingCase{
		{"A session id oversized", encodeMessageA(bytes.Repeat([]byte{0x11}, messageASessionIDCap.length+1), aPoint, nil), "message A session id field too large"},
		{"A point short", encodeMessageA([]byte("sid"), bytes.Repeat([]byte{0x42}, messageAPointCap.length-1), nil), "message A point length"},
		{"A point long", encodeMessageA([]byte("sid"), bytes.Repeat([]byte{0x42}, messageAPointCap.length+1), nil), "message A point length"},
		{"A associated data oversized", encodeMessageA([]byte("sid"), aPoint, bytes.Repeat([]byte{0x22}, messageAAssociatedDataCap.length+1)), "message A associated data field too large"},
		{"B point short", encodeMessageB(bytes.Repeat([]byte{0x42}, messageBPointCap.length-1), nil, bTag), "message B point length"},
		{"B point long", encodeMessageB(bytes.Repeat([]byte{0x42}, messageBPointCap.length+1), nil, bTag), "message B point length"},
		{"B associated data oversized", encodeMessageB(bPoint, bytes.Repeat([]byte{0x33}, messageBAssociatedDataCap.length+1), bTag), "message B associated data field too large"},
		{"B associated data over-declared absent bytes", overDeclaredBAd, "message B associated data field too large"},
		{"B tag short", encodeMessageB(bPoint, nil, bytes.Repeat([]byte{0x99}, messageBTagCap.length-1)), "message B tag length"},
		{"B tag long", encodeMessageB(bPoint, nil, bytes.Repeat([]byte{0x99}, messageBTagCap.length+1)), "message B tag length"},
		{"C tag short", encodeMessageC(bytes.Repeat([]byte{0x99}, messageCTagCap.length-1)), "message C tag length"},
		{"C tag long", encodeMessageC(bytes.Repeat([]byte{0x99}, messageCTagCap.length+1)), "message C tag length"},
	}
}

func messageAFuzzSeeds(validA, crossRoleB, invalidY []byte) [][]byte {
	return [][]byte{
		clone(validA),
		truncatedMessage(validA),
		withMessageRole(validA, roleB),
		append(messageHeader(roleA), 0x80, 0x00),
		encodeMessageA([]byte("sid"), identityEncoding, nil),
		encodeMessageA([]byte("sid"), invalidY, nil),
		clone(crossRoleB),
	}
}

func messageAProtocolFuzzSeeds(validA, crossRoleB, invalidY []byte) [][]byte {
	seeds := messageAFuzzSeeds(validA, crossRoleB, invalidY)
	seeds = append(seeds, encodeMessageA([]byte("other sid"), bytes.Repeat([]byte{0x42}, messageAPointCap.length), nil))
	return seeds
}

func messageBFuzzSeeds(validB, crossRoleC, invalidY []byte) [][]byte {
	return [][]byte{
		clone(validB),
		truncatedMessage(validB),
		withMessageRole(validB, roleA),
		append(messageHeader(roleB), 0x80, 0x00),
		encodeMessageB(identityEncoding, nil, bytes.Repeat([]byte{0x99}, messageBTagCap.length)),
		encodeMessageB(invalidY, nil, bytes.Repeat([]byte{0x99}, messageBTagCap.length)),
		withMessageTamperedLastByte(validB),
		clone(crossRoleC),
	}
}

func messageCFuzzSeeds(validC, crossRoleA []byte) [][]byte {
	return [][]byte{
		clone(validC),
		truncatedMessage(validC),
		withMessageRole(validC, roleA),
		append(messageHeader(roleC), 0x80, 0x00),
		encodeMessageC(bytes.Repeat([]byte{0x99}, messageCTagCap.length-1)),
		withMessageTamperedLastByte(validC),
		clone(crossRoleA),
	}
}

func messageHeader(role byte) []byte {
	return []byte{wireFormatV1, wireSuite, role}
}

func decodeMessageFromCatalogue(msg []byte) error {
	if len(msg) < messageHeaderSize {
		return fmt.Errorf("cpace test: message framing catalogue case has short header")
	}
	switch msg[2] {
	case roleA:
		_, err := decodeMessageA(msg)
		return err
	case roleB:
		_, err := decodeMessageB(msg)
		return err
	case roleC:
		_, err := decodeMessageC(msg)
		return err
	default:
		return fmt.Errorf("cpace test: message framing catalogue case has unexpected role %#x", msg[2])
	}
}

func withMessageHeader(msg []byte, format, suite, role byte) []byte {
	out := clone(msg)
	if len(out) > 0 {
		out[0] = format
	}
	if len(out) > 1 {
		out[1] = suite
	}
	if len(out) > 2 {
		out[2] = role
	}
	return out
}

func withMessageRole(msg []byte, role byte) []byte {
	out := clone(msg)
	if len(out) > 2 {
		out[2] = role
	}
	return out
}

func withMessageTamperedLastByte(msg []byte) []byte {
	out := clone(msg)
	if len(out) > 0 {
		out[len(out)-1] ^= 0x01
	}
	return out
}

func truncatedMessage(msg []byte) []byte {
	if len(msg) == 0 {
		return nil
	}
	return clone(msg[:len(msg)-1])
}

func otherMessageRole(role byte) byte {
	if role == roleA {
		return roleB
	}
	return roleA
}
