package dns

import (
	"context"
	"net"
	"testing"

	"github.com/miekg/dns"
)

type MockDNSResponseWriter struct {
	msgs []*dns.Msg
}

func (w *MockDNSResponseWriter) LocalAddr() net.Addr {
	panic("not implemented")
	return &net.IPAddr{}
}

func (w *MockDNSResponseWriter) RemoteAddr() net.Addr {
	panic("not implemented")
	return &net.IPAddr{}
}

func (w *MockDNSResponseWriter) WriteMsg(m *dns.Msg) error {
	w.msgs = append(w.msgs, m)
	return nil
}

func (w *MockDNSResponseWriter) Write([]byte) (int, error) {
	panic("not implemented")
	return 0, nil
}

func (w *MockDNSResponseWriter) Close() error {
	return nil
}

func (w *MockDNSResponseWriter) TsigStatus() error {
	panic("not implemented")
	return nil
}

func (w *MockDNSResponseWriter) TsigTimersOnly(bool) {
	panic("not implemented")
}

func (w *MockDNSResponseWriter) Hijack() {
	panic("not implemented")
}

var acceptedTypes = map[uint16]bool{
	dns.TypeTXT: true,
	dns.TypeMX:  true,
	dns.TypeSPF: true,
}

func TestWrongRequestType(t *testing.T) {
	w := &MockDNSResponseWriter{}
	m := new(dns.Msg)
	ctx := context.Background()
	hc := HealthCheckPlugin{}
	for qt := range dns.TypeToString {
		// Skip the accepted types
		if _, ok := acceptedTypes[qt]; ok {
			continue
		}
		m.SetQuestion("example.com.", qt)
		response, _ := hc.ServeDNS(ctx, w, m)
		if response != dns.RcodeServerFailure {
			t.Fatalf("Unexpected response value: %d", response)
		}
	}
}
