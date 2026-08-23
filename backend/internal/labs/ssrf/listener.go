package ssrf

import (
	"context"
	"net"
	"net/http/httptest"
	"time"
)

// newLoopbackListener binds an ephemeral port on 127.0.0.1 only.
func newLoopbackListener() net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("ssrf lab: bind loopback listener: " + err.Error())
	}
	return ln
}

func portOf(ln net.Listener) int {
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		panic("ssrf lab: listener is not TCP")
	}
	return addr.Port
}

var _ = httptest.NewServer // keep import surface stable for future test reuse

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
