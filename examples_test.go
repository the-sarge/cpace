package cpace_test

import (
	"bytes"
	"fmt"

	"github.com/the-sarge/cpace"
)

func Example() {
	common := cpace.Config{
		Password:    []byte("correct horse battery staple"),
		InitiatorID: []byte("client@example"),
		ResponderID: []byte("server@example"),
		Context:     []byte("example protocol v1"),
		SessionID:   []byte("session-1234"),
	}

	initCfg := common
	initCfg.AssociatedData = []byte("client hello")
	initiator, msgA, err := cpace.Start(initCfg)
	if err != nil {
		panic(err)
	}

	respCfg := common
	respCfg.AssociatedData = []byte("server hello")
	responder, msgB, err := cpace.Respond(respCfg, msgA)
	if err != nil {
		panic(err)
	}

	msgC, initSession, err := initiator.Finish(msgB)
	if err != nil {
		panic(err)
	}
	respSession, err := responder.Finish(msgC)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := initSession.Close(); err != nil {
			panic(err)
		}
	}()
	defer func() {
		if err := respSession.Close(); err != nil {
			panic(err)
		}
	}()

	initKey, _ := initSession.Export([]byte("application key"), nil, 32)
	respKey, _ := respSession.Export([]byte("application key"), nil, 32)
	fmt.Println(bytes.Equal(initKey, respKey))
	fmt.Println(bytes.Equal(initSession.TranscriptID(), respSession.TranscriptID()))
	// Output:
	// true
	// true
}
