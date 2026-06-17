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
	targets := make([]messageFramingTarget, 0, len(messageFramingCatalogue()))
	for _, spec := range messageFramingCatalogue() {
		spec := spec
		targets = append(targets, messageFramingTarget{
			name:  spec.name,
			role:  spec.role,
			valid: func() []byte { return validMessageForCatalogue(spec) },
			decode: func(msg []byte) error {
				_, err := spec.decode(msg)
				return err
			},
		})
	}
	return targets
}

func validMessageForCatalogue(spec messageSpec) []byte {
	return spec.encode(validMessageFieldsForCatalogue(spec)...)
}

func validMessageFieldsForCatalogue(spec messageSpec) [][]byte {
	fields := make([][]byte, len(spec.fields))
	for i, field := range spec.fields {
		if field.exact {
			fields[i] = bytes.Repeat([]byte{byte(0x40 + i)}, field.length)
			continue
		}
		fields[i] = []byte(fmt.Sprintf("field-%d", i))
	}
	return fields
}

func maxMessageFieldsForCatalogue(spec messageSpec) [][]byte {
	fields := make([][]byte, len(spec.fields))
	for i, field := range spec.fields {
		fields[i] = bytes.Repeat([]byte{byte(0x30 + i)}, field.length)
	}
	return fields
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
		{"non-canonical value LEB128", append(append(messageHeader(role), 0xc0, 0x00), exactMessageFieldBytes(messageCSpec, tagSize, 0x99, 0)...), "non-canonical LEB128"},
		{"malformed LEB128", append(messageHeader(role), 0x80, 0x80, 0x80, 0x80, 0x00), "malformed LEB128"},
	}
}

func messageFramingFinalFieldTruncation(role byte) string {
	for _, spec := range messageFramingCatalogue() {
		if spec.role != role || len(spec.fields) == 0 {
			continue
		}
		return "truncated " + spec.fields[len(spec.fields)-1].name + " field"
	}
	return "truncated"
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
	cases := make([]messageFramingCase, 0, len(messageFramingCatalogue()))
	for _, spec := range messageFramingCatalogue() {
		cases = append(cases, messageFramingCase{
			name: spec.name + " max fields",
			msg:  spec.encode(maxMessageFieldsForCatalogue(spec)...),
		})
	}
	return cases
}

func messageFramingFieldLimitCases() []messageFramingCase {
	var cases []messageFramingCase
	for _, spec := range messageFramingCatalogue() {
		for i, field := range spec.fields {
			if field.exact {
				shortFields := validMessageFieldsForCatalogue(spec)
				shortFields[i] = bytes.Repeat([]byte{byte(0x50 + i)}, field.length-1)
				cases = append(cases, messageFramingCase{
					name:            fmt.Sprintf("%s %s short", spec.name, field.name),
					msg:             spec.encode(shortFields...),
					wantErrContains: field.name + " length",
				})

				longFields := validMessageFieldsForCatalogue(spec)
				longFields[i] = bytes.Repeat([]byte{byte(0x60 + i)}, field.length+1)
				cases = append(cases, messageFramingCase{
					name:            fmt.Sprintf("%s %s long", spec.name, field.name),
					msg:             spec.encode(longFields...),
					wantErrContains: field.name + " length",
				})
				continue
			}

			oversizedFields := validMessageFieldsForCatalogue(spec)
			oversizedFields[i] = bytes.Repeat([]byte{byte(0x70 + i)}, field.length+1)
			cases = append(cases, messageFramingCase{
				name:            fmt.Sprintf("%s %s oversized", spec.name, field.name),
				msg:             spec.encode(oversizedFields...),
				wantErrContains: field.name + " field too large",
			})
			cases = append(cases, messageFramingCase{
				name:            fmt.Sprintf("%s %s over-declared absent bytes", spec.name, field.name),
				msg:             overDeclaredMessageField(spec, i),
				wantErrContains: field.name + " field too large",
			})
		}
	}
	return cases
}

func overDeclaredMessageField(spec messageSpec, fieldIndex int) []byte {
	msg := messageHeader(spec.role)
	fields := validMessageFieldsForCatalogue(spec)
	for i := range fieldIndex {
		msg = appendLengthValue(msg, fields[i])
	}
	fieldLength := spec.fields[fieldIndex].length
	if fieldLength < 0 {
		panic("cpace test: message field length must be non-negative")
	}
	declaredLength := uint64(fieldLength)
	return appendLEB128(msg, declaredLength+1)
}

func messageAFuzzSeeds(validA, crossRoleB, invalidY []byte) [][]byte {
	return messageFuzzSeeds(messageASpec, validA, crossRoleB, invalidY)
}

func messageAProtocolFuzzSeeds(validA, crossRoleB, invalidY []byte) [][]byte {
	seeds := messageAFuzzSeeds(validA, crossRoleB, invalidY)
	seeds = append(seeds, messageWithCatalogueField(messageASpec, 0, []byte("other sid")))
	return seeds
}

func messageBFuzzSeeds(validB, crossRoleC, invalidY []byte) [][]byte {
	return messageFuzzSeeds(messageBSpec, validB, crossRoleC, invalidY)
}

func messageCFuzzSeeds(validC, crossRoleA []byte) [][]byte {
	return messageFuzzSeeds(messageCSpec, validC, crossRoleA, nil)
}

func messageFuzzSeeds(spec messageSpec, valid, crossRole, invalidY []byte) [][]byte {
	seeds := [][]byte{
		clone(valid),
		truncatedMessage(valid),
		withMessageRole(valid, otherMessageRole(spec.role)),
		append(messageHeader(spec.role), 0x80, 0x00),
	}
	if pointIndex, ok := exactMessageFieldIndex(spec, pointSize); ok {
		seeds = append(seeds, messageWithCatalogueField(spec, pointIndex, identityEncoding))
		if invalidY != nil {
			seeds = append(seeds, messageWithCatalogueField(spec, pointIndex, invalidY))
		}
	}
	if tagIndex, ok := exactMessageFieldIndex(spec, tagSize); ok {
		seeds = append(seeds, messageWithCatalogueField(spec, tagIndex, bytes.Repeat([]byte{0x99}, tagSize-1)))
		seeds = append(seeds, withMessageTamperedLastByte(valid))
	}
	seeds = append(seeds, clone(crossRole))
	return seeds
}

func exactMessageFieldIndex(spec messageSpec, length int) (int, bool) {
	for i, field := range spec.fields {
		if field.exact && field.length == length {
			return i, true
		}
	}
	return 0, false
}

func exactMessageFieldBytes(spec messageSpec, length int, fill byte, delta int) []byte {
	i, ok := exactMessageFieldIndex(spec, length)
	if !ok {
		panic("cpace test: exact message field missing from catalogue")
	}
	n := spec.fields[i].length + delta
	if n < 0 {
		n = 0
	}
	return bytes.Repeat([]byte{fill}, n)
}

func messageWithCatalogueField(spec messageSpec, fieldIndex int, value []byte) []byte {
	fields := validMessageFieldsForCatalogue(spec)
	fields[fieldIndex] = value
	return spec.encode(fields...)
}

func messageFieldsAcceptedBySpec(spec messageSpec, fields ...[]byte) bool {
	if len(fields) != len(spec.fields) {
		return false
	}
	remainingSpecs := spec.fields
	for _, got := range fields {
		if len(remainingSpecs) == 0 {
			return false
		}
		field := remainingSpecs[0]
		remainingSpecs = remainingSpecs[1:]
		gotLength := len(got)
		if field.exact {
			if gotLength != field.length {
				return false
			}
			continue
		}
		if gotLength > field.length {
			return false
		}
	}
	return true
}

func messageHeader(role byte) []byte {
	return []byte{wireFormatV1, wireSuite, role}
}

func decodeMessageFromCatalogue(msg []byte) error {
	if len(msg) < messageHeaderSize {
		return fmt.Errorf("cpace test: message framing catalogue case has short header")
	}
	for _, spec := range messageFramingCatalogue() {
		if msg[2] != spec.role {
			continue
		}
		_, err := spec.decode(msg)
		return err
	}
	return fmt.Errorf("cpace test: message framing catalogue case has unexpected role %#x", msg[2])
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
	for _, spec := range messageFramingCatalogue() {
		if spec.role != role {
			return spec.role
		}
	}
	return 0xff
}
