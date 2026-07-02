package cpace

import "testing"

var (
	benchmarkBytesSink       []byte
	benchmarkInitSessionSink *Session
	benchmarkRespSessionSink *Session
	benchmarkMessageASink    messageA
	benchmarkMessageBSink    messageB
	benchmarkMessageCSink    messageC
)

func BenchmarkRoundTrip(b *testing.B) {
	initCfg, respCfg := defaultExchangeInputs()
	initRand := repeatingRand(0x11)
	respRand := repeatingRand(0x22)
	b.ReportAllocs()
	for b.Loop() {
		initiator, msgA, err := startWithRandom(initCfg, initRand)
		if err != nil {
			b.Fatal(err)
		}
		responder, msgB, err := respondWithRandom(respCfg, msgA, respRand)
		if err != nil {
			b.Fatal(err)
		}
		msgC, initSession, err := initiator.Finish(msgB)
		if err != nil {
			b.Fatal(err)
		}
		respSession, err := responder.Finish(msgC)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkInitSessionSink = initSession
		benchmarkRespSessionSink = respSession
	}
}

func BenchmarkStart(b *testing.B) {
	initCfg, _ := defaultExchangeInputs()
	initRand := repeatingRand(0x11)
	b.ReportAllocs()
	for b.Loop() {
		initiator, msgA, err := startWithRandom(initCfg, initRand)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkBytesSink = msgA
		if initiator == nil {
			b.Fatal("nil initiator")
		}
	}
}

func BenchmarkRespond(b *testing.B) {
	initCfg, respCfg := defaultExchangeInputs()
	_, msgA, err := startWithRandom(initCfg, repeatingRand(0x11))
	if err != nil {
		b.Fatal(err)
	}
	respRand := repeatingRand(0x22)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		responder, msgB, err := respondWithRandom(respCfg, msgA, respRand)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkBytesSink = msgB
		if responder == nil {
			b.Fatal("nil responder")
		}
	}
}

func BenchmarkInitiatorFinish(b *testing.B) {
	initCfg, respCfg := defaultExchangeInputs()
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		initiator, msgA, err := startWithRandom(initCfg, repeatingRand(0x11))
		if err != nil {
			b.Fatal(err)
		}
		_, msgB, err := respondWithRandom(respCfg, msgA, repeatingRand(0x22))
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		msgC, session, err := initiator.Finish(msgB)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkBytesSink = msgC
		benchmarkInitSessionSink = session
	}
}

func BenchmarkResponderFinish(b *testing.B) {
	initCfg, respCfg := defaultExchangeInputs()
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		initiator, msgA, err := startWithRandom(initCfg, repeatingRand(0x11))
		if err != nil {
			b.Fatal(err)
		}
		responder, msgB, err := respondWithRandom(respCfg, msgA, repeatingRand(0x22))
		if err != nil {
			b.Fatal(err)
		}
		msgC, _, err := initiator.Finish(msgB)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		session, err := responder.Finish(msgC)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkRespSessionSink = session
	}
}

func BenchmarkSessionExport(b *testing.B) {
	initCfg, respCfg := defaultExchangeInputs()
	initiator, msgA, err := startWithRandom(initCfg, repeatingRand(0x11))
	if err != nil {
		b.Fatal(err)
	}
	responder, msgB, err := respondWithRandom(respCfg, msgA, repeatingRand(0x22))
	if err != nil {
		b.Fatal(err)
	}
	msgC, session, err := initiator.Finish(msgB)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := responder.Finish(msgC); err != nil {
		b.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		size int
	}{
		{name: "32", size: 32},
		{name: "64", size: 64},
		{name: "1024", size: 1024},
	} {
		b.Run(tc.name, func(b *testing.B) {
			label := []byte("application key")
			context := []byte("benchmark context")
			b.ReportAllocs()
			for b.Loop() {
				out, err := session.Export(label, context, tc.size)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkBytesSink = out
			}
		})
	}
}

func BenchmarkEncodeMessage(b *testing.B) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(b, initInput, respInput)
	msgC, _ := exchange.finishInitiator()
	a, err := decodeMessageA(exchange.msgA)
	if err != nil {
		b.Fatal(err)
	}
	msgBDecoded, err := decodeMessageB(exchange.msgB)
	if err != nil {
		b.Fatal(err)
	}
	c, err := decodeMessageC(msgC)
	if err != nil {
		b.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{
			name: "A",
			fn: func() {
				benchmarkBytesSink = encodeMessageA(a.sid, a.ya, a.ada)
			},
		},
		{
			name: "B",
			fn: func() {
				benchmarkBytesSink = encodeMessageB(msgBDecoded.yb, msgBDecoded.adb, msgBDecoded.tag)
			},
		},
		{
			name: "C",
			fn: func() {
				benchmarkBytesSink = encodeMessageC(c.tag)
			},
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				tc.fn()
			}
		})
	}
}

func BenchmarkDecodeMessage(b *testing.B) {
	initInput, respInput := defaultExchangeInputs()
	exchange := newExchange(b, initInput, respInput)
	msgC, _ := exchange.finishInitiator()
	for _, tc := range []struct {
		name string
		size int
		fn   func()
	}{
		{
			name: "A",
			size: len(exchange.msgA),
			fn: func() {
				var err error
				benchmarkMessageASink, err = decodeMessageA(exchange.msgA)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			name: "B",
			size: len(exchange.msgB),
			fn: func() {
				var err error
				benchmarkMessageBSink, err = decodeMessageB(exchange.msgB)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			name: "C",
			size: len(msgC),
			fn: func() {
				var err error
				benchmarkMessageCSink, err = decodeMessageC(msgC)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(tc.size))
			b.ReportAllocs()
			for b.Loop() {
				tc.fn()
			}
		})
	}
}
