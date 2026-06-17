package cpace

import (
	"bytes"
	"slices"
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
		{"self id", "self id", 4 << 10, false},
		{"peer id", "peer id", 4 << 10, false},
		{"context", "context", 1 << 10, false},
		{"session id", "session id", 1 << 10, false},
		{"local associated data", "local associated data", 64 << 10, false},
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
	capPolicy := map[string]packageCapField{}
	for _, field := range shippedPackageCapPolicy() {
		capPolicy[field.name] = field
	}
	want := []struct {
		name   string
		fields []messageFieldSpec
	}{
		{"A", []messageFieldSpec{messageASessionIDCap, messageAPointCap, messageAAssociatedDataCap}},
		{"B", []messageFieldSpec{messageBPointCap, messageBAssociatedDataCap, messageBTagCap}},
		{"C", []messageFieldSpec{messageCTagCap}},
	}
	got := messageFramingCatalogue()
	if len(got) != len(want) {
		t.Fatalf("message framing catalogue length=%d want %d", len(got), len(want))
	}
	for i, tc := range want {
		spec := got[i]
		t.Run(tc.name, func(t *testing.T) {
			if spec.name != tc.name {
				t.Fatalf("name=%q want %q", spec.name, tc.name)
			}
			if !slices.Equal(spec.fields, tc.fields) {
				t.Fatalf("fields=%#v want %#v", spec.fields, tc.fields)
			}
			for _, field := range spec.fields {
				policyField, ok := capPolicy[field.name]
				if !ok {
					t.Fatalf("message field %q is missing from shipped cap policy", field.name)
				}
				if field != policyField {
					t.Fatalf("message field=%#v want cap policy field %#v", field, policyField)
				}
			}
		})
	}
}

func TestPackageOwnedCapPolicyAcceptsInputCopies(t *testing.T) {
	cfg := testInitiatorInput()
	cfg.LocalAssociatedData = []byte("AD")
	accepted, err := acceptInput(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer accepted.wipe()

	for _, field := range [][]byte{
		cfg.Password,
		cfg.SelfID,
		cfg.PeerID,
		cfg.Context,
		cfg.SessionID,
		cfg.LocalAssociatedData,
	} {
		for i := range field {
			field[i] ^= 0xff
		}
	}

	if !bytes.Equal(accepted.password, []byte("password")) {
		t.Fatalf("accepted password aliases caller input: %q", accepted.password)
	}
	if !bytes.Equal(accepted.selfID, []byte("initiator")) {
		t.Fatalf("accepted self ID aliases caller input: %q", accepted.selfID)
	}
	if !bytes.Equal(accepted.peerID, []byte("responder")) {
		t.Fatalf("accepted peer ID aliases caller input: %q", accepted.peerID)
	}
	if !bytes.Equal(accepted.context, []byte("context")) {
		t.Fatalf("accepted context aliases caller input: %q", accepted.context)
	}
	if !bytes.Equal(accepted.sid, []byte("sid")) {
		t.Fatalf("accepted session ID aliases caller input: %q", accepted.sid)
	}
	if !bytes.Equal(accepted.localAD, []byte("AD")) {
		t.Fatalf("accepted local associated data aliases caller input: %q", accepted.localAD)
	}
}

func TestPackageOwnedCapPolicyRejectsInputBeforeCopying(t *testing.T) {
	cfg := testInitiatorInput()
	cfg.LocalAssociatedData = bytes.Repeat([]byte{0x42}, localAssociatedDataCap.length+1)
	originalPassword := clone(cfg.Password)

	accepted, err := acceptInput(cfg)
	if err == nil {
		accepted.wipe()
		t.Fatal("acceptInput succeeded for oversized local associated data")
	}
	if !bytes.Equal(cfg.Password, originalPassword) {
		t.Fatal("acceptInput mutated caller input on a later cap failure")
	}
}
