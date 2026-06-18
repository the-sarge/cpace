package cpace

import (
	"bytes"
	"testing"
)

func TestCallerInputHandoffTransfersOwnershipToNormalizedInput(t *testing.T) {
	cfg := testInitiatorInput()
	cfg.LocalAssociatedData = []byte("AD")
	caller, err := acceptInput(cfg)
	if err != nil {
		t.Fatal(err)
	}
	contextStorage := caller.context

	normalized := caller.handoff(initiatorInputRole)
	caller.wipe()
	defer normalized.wipe()

	if !bytes.Equal(normalized.password, []byte("password")) {
		t.Fatalf("normalized password got %q", normalized.password)
	}
	if !bytes.Equal(normalized.initiatorID, []byte("initiator")) {
		t.Fatalf("normalized initiator ID got %q", normalized.initiatorID)
	}
	if !bytes.Equal(normalized.responderID, []byte("responder")) {
		t.Fatalf("normalized responder ID got %q", normalized.responderID)
	}
	if !bytes.Equal(normalized.sid, []byte("sid")) {
		t.Fatalf("normalized session ID got %q", normalized.sid)
	}
	if !bytes.Equal(normalized.ad, []byte("AD")) {
		t.Fatalf("normalized associated data got %q", normalized.ad)
	}
	wantCI := buildCI([]byte("initiator"), []byte("responder"), []byte("context"))
	if !bytes.Equal(normalized.ci, wantCI) {
		t.Fatalf("normalized CI got %x want %x", normalized.ci, wantCI)
	}
	if !allZero(contextStorage) {
		t.Fatalf("caller context storage was not wiped: %q", contextStorage)
	}
	if caller.password != nil || caller.selfID != nil || caller.peerID != nil || caller.context != nil || caller.sid != nil || caller.localAD != nil {
		t.Fatalf("caller input retained handed-off references: %#v", caller)
	}
}

func TestCallerInputHandoffMapsResponderRole(t *testing.T) {
	caller, err := acceptInput(testResponderInput())
	if err != nil {
		t.Fatal(err)
	}
	normalized := caller.handoff(responderInputRole)
	defer normalized.wipe()

	if !bytes.Equal(normalized.initiatorID, []byte("initiator")) {
		t.Fatalf("normalized initiator ID got %q", normalized.initiatorID)
	}
	if !bytes.Equal(normalized.responderID, []byte("responder")) {
		t.Fatalf("normalized responder ID got %q", normalized.responderID)
	}
}
