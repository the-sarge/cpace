package cpace

import (
	"bytes"
	"testing"
)

func TestPackageOwnedCapPolicyPinsShippedValues(t *testing.T) {
	want := []struct {
		name       string
		wantName   string
		wantLength int
		wantExact  bool
	}{
		{"password", "password", 4 << 10, false},
		{"initiator id", "initiator id", 4 << 10, false},
		{"responder id", "responder id", 4 << 10, false},
		{"context", "context", 1 << 10, false},
		{"session id", "session id", 1 << 10, false},
		{"associated data", "associated data", 64 << 10, false},
		{"message A session id", "message A session id", 1 << 10, false},
		{"message A point", "message A point", pointSize, true},
		{"message A associated data", "message A associated data", 64 << 10, false},
		{"message B point", "message B point", pointSize, true},
		{"message B associated data", "message B associated data", 64 << 10, false},
		{"message B tag", "message B tag", tagSize, true},
		{"message C tag", "message C tag", tagSize, true},
	}
	got := shippedPackageCapPolicy()
	if len(got) != len(want) {
		t.Fatalf("shipped cap policy length=%d want %d", len(got), len(want))
	}
	for i, tc := range want {
		t.Run(tc.name, func(t *testing.T) {
			field := got[i]
			if field.name != tc.wantName {
				t.Fatalf("name=%q want %q", field.name, tc.wantName)
			}
			if field.length != tc.wantLength {
				t.Fatalf("length=%d want %d", field.length, tc.wantLength)
			}
			if field.exact != tc.wantExact {
				t.Fatalf("exact=%t want %t", field.exact, tc.wantExact)
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

func TestPackageOwnedCapPolicyAcceptsConfigCopies(t *testing.T) {
	cfg := testConfig()
	cfg.AssociatedData = []byte("AD")
	accepted, err := acceptConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer accepted.wipe()

	for _, field := range [][]byte{
		cfg.Password,
		cfg.InitiatorID,
		cfg.ResponderID,
		cfg.Context,
		cfg.SessionID,
		cfg.AssociatedData,
	} {
		for i := range field {
			field[i] ^= 0xff
		}
	}

	if !bytes.Equal(accepted.password, []byte("password")) {
		t.Fatalf("accepted password aliases caller input: %q", accepted.password)
	}
	if !bytes.Equal(accepted.initiatorID, []byte("initiator")) {
		t.Fatalf("accepted initiator ID aliases caller input: %q", accepted.initiatorID)
	}
	if !bytes.Equal(accepted.responderID, []byte("responder")) {
		t.Fatalf("accepted responder ID aliases caller input: %q", accepted.responderID)
	}
	if !bytes.Equal(accepted.context, []byte("context")) {
		t.Fatalf("accepted context aliases caller input: %q", accepted.context)
	}
	if !bytes.Equal(accepted.sid, []byte("sid")) {
		t.Fatalf("accepted session ID aliases caller input: %q", accepted.sid)
	}
	if !bytes.Equal(accepted.ad, []byte("AD")) {
		t.Fatalf("accepted associated data aliases caller input: %q", accepted.ad)
	}
}

func TestPackageOwnedCapPolicyRejectsConfigBeforeCopying(t *testing.T) {
	cfg := testConfig()
	cfg.AssociatedData = bytes.Repeat([]byte{0x42}, associatedDataCap.length+1)
	originalPassword := clone(cfg.Password)

	accepted, err := acceptConfig(cfg)
	if err == nil {
		accepted.wipe()
		t.Fatal("acceptConfig succeeded for oversized associated data")
	}
	if !bytes.Equal(cfg.Password, originalPassword) {
		t.Fatal("acceptConfig mutated caller input on a later cap failure")
	}
}
