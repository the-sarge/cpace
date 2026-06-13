package cpace

import "testing"

func TestPackageOwnedCapPolicyPinsShippedValues(t *testing.T) {
	cases := []struct {
		name       string
		field      packageCapField
		wantName   string
		wantLength int
		wantExact  bool
	}{
		{"password", passwordCap, "password", 4 << 10, false},
		{"initiator id", initiatorIDCap, "initiator id", 4 << 10, false},
		{"responder id", responderIDCap, "responder id", 4 << 10, false},
		{"context", contextCap, "context", 1 << 10, false},
		{"session id", sessionIDCap, "session id", 1 << 10, false},
		{"associated data", associatedDataCap, "associated data", 64 << 10, false},
		{"message A session id", messageASessionIDCap, "message A session id", 1 << 10, false},
		{"message A point", messageAPointCap, "message A point", pointSize, true},
		{"message A associated data", messageAAssociatedDataCap, "message A associated data", 64 << 10, false},
		{"message B point", messageBPointCap, "message B point", pointSize, true},
		{"message B associated data", messageBAssociatedDataCap, "message B associated data", 64 << 10, false},
		{"message B tag", messageBTagCap, "message B tag", tagSize, true},
		{"message C tag", messageCTagCap, "message C tag", tagSize, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.field.name != tc.wantName {
				t.Fatalf("name=%q want %q", tc.field.name, tc.wantName)
			}
			if tc.field.length != tc.wantLength {
				t.Fatalf("length=%d want %d", tc.field.length, tc.wantLength)
			}
			if tc.field.exact != tc.wantExact {
				t.Fatalf("exact=%t want %t", tc.field.exact, tc.wantExact)
			}
		})
	}
}

func TestPackageOwnedCapPolicyFeedsMessageFramingSpecs(t *testing.T) {
	cases := []struct {
		name string
		got  packageCapField
		want packageCapField
	}{
		{"message A session id", messageASpec.fields[0], messageASessionIDCap},
		{"message A point", messageASpec.fields[1], messageAPointCap},
		{"message A associated data", messageASpec.fields[2], messageAAssociatedDataCap},
		{"message B point", messageBSpec.fields[0], messageBPointCap},
		{"message B associated data", messageBSpec.fields[1], messageBAssociatedDataCap},
		{"message B tag", messageBSpec.fields[2], messageBTagCap},
		{"message C tag", messageCSpec.fields[0], messageCTagCap},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("message field=%#v want cap policy field %#v", tc.got, tc.want)
			}
		})
	}
}
