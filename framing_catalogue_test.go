package cpace

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
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

func TestMessageFramingDecodeReturnsOwnedFields(t *testing.T) {
	msg := encodeMessageA([]byte("sid"), bytes.Repeat([]byte{0x42}, pointSize), []byte("ADa"))
	got, err := decodeMessageA(msg)
	if err != nil {
		t.Fatal(err)
	}

	msg[len(msg)-1] ^= 0xff
	if !bytes.Equal(got.ada, []byte("ADa")) {
		t.Fatalf("decoded associated data aliases message buffer: %q", got.ada)
	}
	msg[messageHeaderSize+1] ^= 0xff
	if !bytes.Equal(got.sid, []byte("sid")) {
		t.Fatalf("decoded session id aliases message buffer: %q", got.sid)
	}
}

func TestMessageFramingCatalogueRejectsMalformed(t *testing.T) {
	for _, target := range messageFramingTargets() {
		t.Run(target.name, func(t *testing.T) {
			for _, tc := range messageFramingMalformedCases(target) {
				t.Run(tc.name, func(t *testing.T) {
					assertMessageFramingError(t, target.decode(tc.msg), tc.wantErrContains)
				})
			}
		})
	}
}

func TestMessageFramingCataloguePinsAggregateSizePrecedence(t *testing.T) {
	for _, target := range messageFramingTargets() {
		t.Run(target.name, func(t *testing.T) {
			for _, tc := range messageFramingAggregateCases(target) {
				t.Run(tc.name, func(t *testing.T) {
					err := target.decode(tc.msg)
					assertMessageFramingError(t, err, tc.wantErrContains)
					if tc.wantErrContains != "message too large" && strings.Contains(err.Error(), "message too large") {
						t.Fatalf("decode oversized %s err=%v reached size check before %q", target.name, err, tc.wantErrContains)
					}
				})
			}
		})
	}
}

func TestMessageFramingCatalogueAcceptsMaxFields(t *testing.T) {
	for _, tc := range messageFramingMaxFieldCases() {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.msg) >= maxMessageLength {
				t.Fatalf("max-size message len=%d exceeds aggregate cap %d", len(tc.msg), maxMessageLength)
			}
			if err := decodeMessageFromCatalogue(tc.msg); err != nil {
				t.Fatalf("decode max fields: %v", err)
			}
		})
	}
}

func TestMessageFramingCatalogueOwnsFieldLengthAcceptance(t *testing.T) {
	for _, spec := range messageFramingCatalogue() {
		t.Run(spec.name, func(t *testing.T) {
			maxFields := maxMessageFieldsForCatalogue(spec)
			if !spec.acceptsFieldLengths(maxFields...) {
				t.Fatal("rejected max-size fields")
			}
			if spec.acceptsFieldLengths(maxFields[:len(maxFields)-1]...) {
				t.Fatal("accepted too few fields")
			}
			if spec.acceptsFieldLengths(append(maxFields, nil)...) {
				t.Fatal("accepted wrong field count")
			}
			for i, field := range spec.fields {
				if field.exact && field.length > 0 {
					invalidLength := field.length - 1
					fields := messageFieldsForCatalogueWithLength(spec, i, invalidLength)
					if spec.acceptsFieldLengths(fields...) {
						t.Fatalf("accepted invalid %s length %d", field.name, invalidLength)
					}
				}

				invalidLength := field.length + 1
				fields := messageFieldsForCatalogueWithLength(spec, i, invalidLength)
				if spec.acceptsFieldLengths(fields...) {
					t.Fatalf("accepted invalid %s length %d", field.name, invalidLength)
				}
			}
		})
	}
}

func TestMessageFramingCatalogueRejectsFieldLimits(t *testing.T) {
	for _, tc := range messageFramingFieldLimitCases() {
		t.Run(tc.name, func(t *testing.T) {
			assertMessageFramingError(t, decodeMessageFromCatalogue(tc.msg), tc.wantErrContains)
		})
	}
}

func assertMessageFramingError(t *testing.T, err error, wantErrContains string) {
	t.Helper()
	if !errors.Is(err, ErrMessage) {
		t.Fatalf("decode err=%v want ErrMessage", err)
	}
	if wantErrContains != "" && !strings.Contains(err.Error(), wantErrContains) {
		t.Fatalf("decode err=%q missing %q", err, wantErrContains)
	}
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
		fields[i] = fmt.Appendf(nil, "field-%d", i)
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

func messageFieldsForCatalogueWithLength(spec messageSpec, target, length int) [][]byte {
	fields := make([][]byte, 0, len(spec.fields))
	for i, field := range spec.fields {
		n := field.length
		if i == target {
			n = length
		}
		fields = append(fields, bytes.Repeat([]byte{byte(0x30 + i)}, n))
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
	seeds = append(seeds, messageWithDecodedField(messageASpec, validA, 0, []byte("other sid")))
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
	if pointIndex, ok := spec.exactFieldIndex(pointSize); ok {
		seeds = append(seeds, messageWithDecodedField(spec, valid, pointIndex, identityEncoding))
		if invalidY != nil {
			seeds = append(seeds, messageWithDecodedField(spec, valid, pointIndex, invalidY))
		}
	}
	if tagIndex, ok := spec.exactFieldIndex(tagSize); ok {
		seeds = append(seeds, messageWithDecodedField(spec, valid, tagIndex, bytes.Repeat([]byte{0x99}, tagSize-1)))
		seeds = append(seeds, withMessageTamperedLastByte(valid))
	}
	seeds = append(seeds, clone(crossRole))
	return seeds
}

func TestMessageAProtocolFuzzSeedsPreserveValidFields(t *testing.T) {
	_, _, _, msgA, msgB, _ := makeFuzzExchange(t)
	baseA, err := decodeMessageA(msgA)
	if err != nil {
		t.Fatal(err)
	}
	invalid := fuzzDraftInvalidVector(t)
	var decoded messageAProtocolFuzzSeedCounts
	for _, seed := range messageAProtocolFuzzSeeds(msgA, msgB, invalid.InvalidY1) {
		got, err := decodeMessageA(seed)
		if err != nil {
			continue
		}
		category := classifyMessageAProtocolFuzzSeed(got, baseA, invalid.InvalidY1)
		decoded.add(category)
		if category == messageAProtocolFuzzSeedUnclassified {
			t.Fatalf("unclassified Message A protocol fuzz seed decoded as sid=%x ya=%x ada=%x", got.sid, got.ya, got.ada)
		}
	}
	if want := (messageAProtocolFuzzSeedCounts{valid: 1, identityPoint: 1, invalidPoint: 1, otherSessionID: 1}); decoded != want {
		t.Fatalf("decoded Message A protocol fuzz seed categories=%+v want %+v", decoded, want)
	}
}

func TestMessageAProtocolFuzzSeedPreservationRejectsUnclassifiedDecode(t *testing.T) {
	_, _, _, msgA, _, _ := makeFuzzExchange(t)
	baseA, err := decodeMessageA(msgA)
	if err != nil {
		t.Fatal(err)
	}
	unclassified := encodeMessageA(baseA.sid, baseA.ya, []byte("unexpected ADa"))
	got, err := decodeMessageA(unclassified)
	if err != nil {
		t.Fatal(err)
	}
	if classifyMessageAProtocolFuzzSeed(got, baseA, fuzzDraftInvalidVector(t).InvalidY1) != messageAProtocolFuzzSeedUnclassified {
		t.Fatal("unclassified successful decode passed preservation assertion")
	}
}

type messageAProtocolFuzzSeedCategory byte

const (
	messageAProtocolFuzzSeedUnclassified messageAProtocolFuzzSeedCategory = iota
	messageAProtocolFuzzSeedValid
	messageAProtocolFuzzSeedIdentityPoint
	messageAProtocolFuzzSeedInvalidPoint
	messageAProtocolFuzzSeedOtherSessionID
)

type messageAProtocolFuzzSeedCounts struct {
	valid          int
	identityPoint  int
	invalidPoint   int
	otherSessionID int
}

func (c *messageAProtocolFuzzSeedCounts) add(category messageAProtocolFuzzSeedCategory) {
	switch category {
	case messageAProtocolFuzzSeedUnclassified:
	case messageAProtocolFuzzSeedValid:
		c.valid++
	case messageAProtocolFuzzSeedIdentityPoint:
		c.identityPoint++
	case messageAProtocolFuzzSeedInvalidPoint:
		c.invalidPoint++
	case messageAProtocolFuzzSeedOtherSessionID:
		c.otherSessionID++
	}
}

func classifyMessageAProtocolFuzzSeed(got, base messageA, invalidY []byte) messageAProtocolFuzzSeedCategory {
	switch {
	case bytes.Equal(got.sid, base.sid) && bytes.Equal(got.ya, base.ya) && bytes.Equal(got.ada, base.ada):
		return messageAProtocolFuzzSeedValid
	case bytes.Equal(got.ya, identityEncoding) && bytes.Equal(got.sid, base.sid) && bytes.Equal(got.ada, base.ada):
		return messageAProtocolFuzzSeedIdentityPoint
	case bytes.Equal(got.ya, invalidY) && bytes.Equal(got.sid, base.sid) && bytes.Equal(got.ada, base.ada):
		return messageAProtocolFuzzSeedInvalidPoint
	case bytes.Equal(got.sid, []byte("other sid")):
		if bytes.Equal(got.ya, base.ya) && bytes.Equal(got.ada, base.ada) {
			return messageAProtocolFuzzSeedOtherSessionID
		}
		return messageAProtocolFuzzSeedUnclassified
	default:
		return messageAProtocolFuzzSeedUnclassified
	}
}

func TestMessageBFuzzSeedsPreserveValidFields(t *testing.T) {
	_, _, _, _, msgB, msgC := makeFuzzExchange(t)
	baseB, err := decodeMessageB(msgB)
	if err != nil {
		t.Fatal(err)
	}
	tamperedB, err := decodeMessageB(withMessageTamperedLastByte(msgB))
	if err != nil {
		t.Fatal(err)
	}
	invalid := fuzzDraftInvalidVector(t)
	var decoded messageBFuzzSeedCounts
	for _, seed := range messageBFuzzSeeds(msgB, msgC, invalid.InvalidY1) {
		got, err := decodeMessageB(seed)
		if err != nil {
			continue
		}
		category := classifyMessageBFuzzSeed(got, baseB, invalid.InvalidY1, tamperedB.tag)
		decoded.add(category)
		if category == messageBFuzzSeedUnclassified {
			t.Fatalf("unclassified Message B fuzz seed decoded as yb=%x adb=%x tag=%x", got.yb, got.adb, got.tag)
		}
	}
	if want := (messageBFuzzSeedCounts{valid: 1, identityPoint: 1, invalidPoint: 1, tamperedTag: 1}); decoded != want {
		t.Fatalf("decoded Message B fuzz seed categories=%+v want %+v", decoded, want)
	}
}

func TestMessageBFuzzSeedPreservationRejectsUnclassifiedDecode(t *testing.T) {
	_, _, _, _, msgB, _ := makeFuzzExchange(t)
	baseB, err := decodeMessageB(msgB)
	if err != nil {
		t.Fatal(err)
	}
	tamperedB, err := decodeMessageB(withMessageTamperedLastByte(msgB))
	if err != nil {
		t.Fatal(err)
	}
	unclassified := encodeMessageB(baseB.yb, []byte("unexpected ADb"), baseB.tag)
	got, err := decodeMessageB(unclassified)
	if err != nil {
		t.Fatal(err)
	}
	if classifyMessageBFuzzSeed(got, baseB, fuzzDraftInvalidVector(t).InvalidY1, tamperedB.tag) != messageBFuzzSeedUnclassified {
		t.Fatal("unclassified successful Message B decode passed preservation assertion")
	}
}

type messageBFuzzSeedCategory byte

const (
	messageBFuzzSeedUnclassified messageBFuzzSeedCategory = iota
	messageBFuzzSeedValid
	messageBFuzzSeedIdentityPoint
	messageBFuzzSeedInvalidPoint
	messageBFuzzSeedTamperedTag
)

type messageBFuzzSeedCounts struct {
	valid         int
	identityPoint int
	invalidPoint  int
	tamperedTag   int
}

func (c *messageBFuzzSeedCounts) add(category messageBFuzzSeedCategory) {
	switch category {
	case messageBFuzzSeedUnclassified:
	case messageBFuzzSeedValid:
		c.valid++
	case messageBFuzzSeedIdentityPoint:
		c.identityPoint++
	case messageBFuzzSeedInvalidPoint:
		c.invalidPoint++
	case messageBFuzzSeedTamperedTag:
		c.tamperedTag++
	}
}

func classifyMessageBFuzzSeed(got, base messageB, invalidY, tamperedTag []byte) messageBFuzzSeedCategory {
	switch {
	case bytes.Equal(got.yb, base.yb) && bytes.Equal(got.adb, base.adb) && bytes.Equal(got.tag, base.tag):
		return messageBFuzzSeedValid
	case bytes.Equal(got.yb, identityEncoding) && bytes.Equal(got.adb, base.adb) && bytes.Equal(got.tag, base.tag):
		return messageBFuzzSeedIdentityPoint
	case bytes.Equal(got.yb, invalidY) && bytes.Equal(got.adb, base.adb) && bytes.Equal(got.tag, base.tag):
		return messageBFuzzSeedInvalidPoint
	case bytes.Equal(got.tag, tamperedTag):
		if bytes.Equal(got.yb, base.yb) && bytes.Equal(got.adb, base.adb) {
			return messageBFuzzSeedTamperedTag
		}
		return messageBFuzzSeedUnclassified
	default:
		return messageBFuzzSeedUnclassified
	}
}

func TestExactMessageFieldIndexRejectsAmbiguousLengths(t *testing.T) {
	spec := messageSpec{
		name: "ambiguous",
		role: roleA,
		fields: []messageFieldSpec{
			{name: "first exact", length: pointSize, exact: true},
			{name: "second exact", length: pointSize, exact: true},
		},
	}
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("messageSpec.exactFieldIndex accepted ambiguous exact field lengths")
		}
		if !strings.Contains(fmt.Sprint(got), "ambiguous exact 32-byte field") {
			t.Fatalf("panic=%v want ambiguous exact field diagnostic", got)
		}
	}()
	_, _ = spec.exactFieldIndex(pointSize)
}

func TestMessageFuzzSeedsRejectsAmbiguousExactFieldLengths(t *testing.T) {
	spec := messageSpec{
		name: "ambiguous",
		role: roleA,
		fields: []messageFieldSpec{
			{name: "first exact", length: pointSize, exact: true},
			{name: "second exact", length: pointSize, exact: true},
		},
	}
	valid := spec.encode(bytes.Repeat([]byte{0x11}, pointSize), bytes.Repeat([]byte{0x22}, pointSize))
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("messageFuzzSeeds accepted ambiguous exact field lengths")
		}
		if !strings.Contains(fmt.Sprint(got), "ambiguous exact 32-byte field") {
			t.Fatalf("panic=%v want ambiguous exact field diagnostic", got)
		}
	}()
	_ = messageFuzzSeeds(spec, valid, withMessageRole(valid, otherMessageRole(spec.role)), nil)
}

func TestMessageFuzzSeedsSkipsAbsentExactFieldLengths(t *testing.T) {
	spec := messageSpec{
		name: "no exact fields",
		role: roleA,
		fields: []messageFieldSpec{
			{name: "associated data", length: maxAssociatedDataLength},
		},
	}
	valid := spec.encode([]byte("AD"))
	crossRole := append(messageHeader(otherMessageRole(spec.role)), 0x01, 0x02)
	seeds := messageFuzzSeeds(spec, valid, crossRole, nil)
	want := [][]byte{
		clone(valid),
		truncatedMessage(valid),
		withMessageRole(valid, otherMessageRole(spec.role)),
		append(messageHeader(spec.role), 0x80, 0x00),
		clone(crossRole),
	}
	if len(seeds) != len(want) {
		t.Fatalf("messageFuzzSeeds returned %d seeds, want %d", len(seeds), len(want))
	}
	for i := range want {
		if !bytes.Equal(seeds[i], want[i]) {
			t.Fatalf("messageFuzzSeeds seed %d=%x want %x", i, seeds[i], want[i])
		}
	}
}

func (spec messageSpec) exactFieldIndex(length int) (int, bool) {
	found := -1
	foundName := ""
	for i, field := range spec.fields {
		if field.exact && field.length == length {
			if found >= 0 {
				panic(fmt.Sprintf("cpace test: %s has ambiguous exact %d-byte field: %s and %s", spec.name, length, foundName, field.name))
			}
			found = i
			foundName = field.name
		}
	}
	if found < 0 {
		return 0, false
	}
	return found, true
}

func exactMessageFieldBytes(spec messageSpec, length int, fill byte, delta int) []byte {
	i, ok := spec.exactFieldIndex(length)
	if !ok {
		panic("cpace test: exact message field missing from catalogue")
	}
	n := max(spec.fields[i].length+delta, 0)
	return bytes.Repeat([]byte{fill}, n)
}

func messageWithDecodedField(spec messageSpec, msg []byte, fieldIndex int, value []byte) []byte {
	fields, err := spec.decode(msg)
	if err != nil {
		panic(fmt.Sprintf("cpace test: valid %s message failed to decode: %v", spec.name, err))
	}
	fields[fieldIndex] = value
	return spec.encode(fields...)
}

func (spec messageSpec) acceptsFieldLengths(fields ...[]byte) bool {
	if len(fields) != len(spec.fields) {
		return false
	}
	remainingSpecs := spec.fields
	for _, got := range fields {
		field := remainingSpecs[0]
		remainingSpecs = remainingSpecs[1:]
		if err := field.validateMessageLength(len(got)); err != nil {
			return false
		}
	}
	return true
}

func messageFieldsMatchFramingShape(spec messageSpec, fields ...[]byte) bool {
	if len(fields) != len(spec.fields) {
		return false
	}
	// Keep this independent from acceptsFieldLengths and validateMessageLength so round-trip fuzzers can detect decoder acceptance drift.
	remainingSpecs := spec.fields
	for _, got := range fields {
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
