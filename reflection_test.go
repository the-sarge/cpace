package cpace

import (
	"bytes"
	"errors"
	"testing"
)

func reflectionADCases() []struct {
	name string
	a, b []byte
} {
	return []struct {
		name string
		a, b []byte
	}{
		{name: "empty"},
		{name: "equal_nonempty", a: []byte("same"), b: []byte("same")},
		{name: "different", a: []byte("first"), b: []byte("second")},
	}
}

// Forced equality is a test-only condition: the attacker cannot select the
// responder's fresh scalar through Respond. Both inputs retain distinct party
// identities and a nonempty shared SessionID.
func mustReflectionExchange(t *testing.T, ada, adb []byte, equalShares bool) *exchangeFixture {
	t.Helper()
	iInput, rInput := testInitiatorInput(), testResponderInput()
	iInput.LocalAssociatedData, rInput.LocalAssociatedData = ada, adb
	start, respond := Start, Respond
	if equalShares {
		start = startTestInitiator
		respond = func(input Input, msgA []byte) (*Responder, []byte, error) {
			return respondWithRandom(input, msgA, repeatingRand(0x11))
		}
	}
	i, msgA, err := start(iInput)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = i.Close() })
	r, msgB, err := respond(rInput, msgA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	a, err := decodeMessageA(msgA)
	if err != nil {
		t.Fatal(err)
	}
	b, err := decodeMessageB(msgB)
	if err != nil {
		t.Fatal(err)
	}
	if got := bytes.Equal(a.ya, b.yb); got != equalShares {
		t.Fatalf("share equality got %v want %v", got, equalShares)
	}
	return &exchangeFixture{tb: t, initiator: i, responder: r, msgA: msgA, msgB: msgB}
}

func TestReflectionResponderRejectsOwnTag(t *testing.T) {
	for _, ad := range reflectionADCases() {
		for _, mode := range []struct {
			name        string
			equalShares bool
		}{{name: "production_randomness"}, {name: "forced_equal_shares", equalShares: true}} {
			t.Run(ad.name+"/"+mode.name, func(t *testing.T) {
				x := mustReflectionExchange(t, ad.a, ad.b, mode.equalShares)
				b, err := decodeMessageB(x.msgB)
				if err != nil {
					t.Fatal(err)
				}
				// Copy only the authentic wire tag. The initiator has never
				// called Finish, and the attacker reads no secret material.
				msgC := encodeMessageC(b.tag)
				secrets := snapshotResponderSecrets(t, x.responder)
				session, err := x.responder.Finish(msgC)
				if session != nil || !errors.Is(err, ErrConfirmationFailed) {
					t.Fatalf("reflected C got session=%v err=%v want nil ErrConfirmationFailed", session != nil, err)
				}
				secrets.assertCleared()
				if session, err := x.responder.Finish(msgC); session != nil || !errors.Is(err, ErrStateUsed) {
					t.Fatalf("second Finish got session=%v err=%v want nil ErrStateUsed", session != nil, err)
				}
			})
		}
	}
}

func TestReflectionInitiatorRejectsReflectedShareAndAuthenticTag(t *testing.T) {
	for _, ad := range reflectionADCases() {
		for _, mode := range []struct {
			name        string
			equalShares bool
		}{{name: "production_randomness"}, {name: "forced_equal_shares", equalShares: true}} {
			t.Run(ad.name+"/"+mode.name, func(t *testing.T) {
				x := mustReflectionExchange(t, ad.a, ad.b, mode.equalShares)
				a, err := decodeMessageA(x.msgA)
				if err != nil {
					t.Fatal(err)
				}
				b, err := decodeMessageB(x.msgB)
				if err != nil {
					t.Fatal(err)
				}
				// Reframe A's share and AD as B and copy an authentic B tag.
				// For forced equal shares and equal AD this is the genuine B.
				reflectedB := encodeMessageB(a.ya, a.ada, b.tag)
				assertReflectionInitiatorRejected(t, x.initiator, reflectedB)
			})
		}
	}
}

func assertReflectionInitiatorRejected(t *testing.T, i *Initiator, msgB []byte) {
	t.Helper()
	secrets := snapshotInitiatorSecrets(t, i)
	msgC, session, err := i.Finish(msgB)
	if msgC != nil || session != nil || !errors.Is(err, ErrConfirmationFailed) {
		t.Fatalf("reflected B got C=%x session=%v err=%v want nil nil ErrConfirmationFailed", msgC, session != nil, err)
	}
	secrets.assertCleared()
	if msgC, session, err := i.Finish(msgB); msgC != nil || session != nil || !errors.Is(err, ErrStateUsed) {
		t.Fatalf("second Finish got C=%x session=%v err=%v want nil nil ErrStateUsed", msgC, session != nil, err)
	}
}

func TestReflectionInitiatorWithoutTagOracle(t *testing.T) {
	for _, ad := range reflectionADCases() {
		t.Run(ad.name, func(t *testing.T) {
			input := testInitiatorInput()
			input.LocalAssociatedData = ad.a
			i, msgA, err := Start(input)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = i.Close() })
			a, err := decodeMessageA(msgA)
			if err != nil {
				t.Fatal(err)
			}
			// A has no tag to reflect. A network attacker must guess B's
			// mandatory tag before the initiator will release C.
			msgB := encodeMessageB(a.ya, ad.b, make([]byte, tagSize))
			assertReflectionInitiatorRejected(t, i, msgB)
		})
	}
}

func TestReflectionHonestExchangeControls(t *testing.T) {
	for _, ad := range reflectionADCases() {
		t.Run(ad.name, func(t *testing.T) {
			x := mustReflectionExchange(t, ad.a, ad.b, false)
			iSession, rSession := x.complete()
			defer iSession.Close()
			defer rSession.Close()
			assertReflectionSessionKeysMatch(t, iSession, rSession)
		})
	}
	// Rejecting equal tags must not become a blanket equal-share ban:
	// differing AD produces distinct tags even with forced equal shares.
	t.Run("forced_equal_shares_different_ad", func(t *testing.T) {
		x := mustReflectionExchange(t, []byte("first"), []byte("second"), true)
		msgC, iSession := x.finishInitiator()
		defer iSession.Close()
		b, err := decodeMessageB(x.msgB)
		if err != nil {
			t.Fatal(err)
		}
		c, err := decodeMessageC(msgC)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(b.tag, c.tag) {
			t.Fatal("confirmation tags got equal want distinct with differing AD")
		}
		rSession := x.finishResponder(msgC)
		defer rSession.Close()
		assertReflectionSessionKeysMatch(t, iSession, rSession)
	})
}

func assertReflectionSessionKeysMatch(t *testing.T, iSession, rSession *Session) {
	t.Helper()
	iKey, err := iSession.Export([]byte("reflection-control"), nil, 32)
	if err != nil {
		t.Fatal(err)
	}
	defer clearBytes(iKey)
	rKey, err := rSession.Export([]byte("reflection-control"), nil, 32)
	if err != nil {
		t.Fatal(err)
	}
	defer clearBytes(rKey)
	if !bytes.Equal(iKey, rKey) {
		t.Fatal("exported keys got unequal want equal")
	}
}

func TestReflectionRejectsTagsFromAnotherExchange(t *testing.T) {
	for _, ad := range reflectionADCases() {
		t.Run(ad.name, func(t *testing.T) {
			// Deliberately reuse all caller inputs, including SessionID.
			// Fresh shares still bind each tag to its own exchange.
			first := mustReflectionExchange(t, ad.a, ad.b, false)
			second := mustReflectionExchange(t, ad.a, ad.b, false)
			firstB, err := decodeMessageB(first.msgB)
			if err != nil {
				t.Fatal(err)
			}
			secondB, err := decodeMessageB(second.msgB)
			if err != nil {
				t.Fatal(err)
			}
			msgC, iSession := first.finishInitiator()
			defer iSession.Close()
			rSession := first.finishResponder(msgC)
			defer rSession.Close()
			assertReflectionInitiatorRejected(t, second.initiator, encodeMessageB(secondB.yb, secondB.adb, firstB.tag))
			secrets := snapshotResponderSecrets(t, second.responder)
			if session, err := second.responder.Finish(msgC); session != nil || !errors.Is(err, ErrConfirmationFailed) {
				t.Fatalf("replayed C got session=%v err=%v want nil ErrConfirmationFailed", session != nil, err)
			}
			secrets.assertCleared()
			if session, err := second.responder.Finish(msgC); session != nil || !errors.Is(err, ErrStateUsed) {
				t.Fatalf("second Finish got session=%v err=%v want nil ErrStateUsed", session != nil, err)
			}
		})
	}
}
