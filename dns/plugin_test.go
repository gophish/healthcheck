package dns

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/coredns/coredns/request"
	"github.com/gophish/healthcheck/config"
	"github.com/gophish/healthcheck/db"
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

func setupConfig(t *testing.T) {
	config.Config.DBName = "sqlite3"
	config.Config.DBPath = ":memory:"
	config.Config.MigrationsPath = "../db/sqlite3/migrations/"
	config.Config.EmailHostname = "example.com"
	err := db.Setup()
	if err != nil {
		t.Fatalf("Failed setting up the database: %s", err.Error())
	}
}

func createMessage() *db.Message {
	return &db.Message{
		Recipient:  "test@example.com",
		MailServer: "localhost",
	}
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

func TestGenerateSPFTemplate(t *testing.T) {
	setupConfig(t)
	testSuite := map[string]string{
		db.Pass:     fmt.Sprintf("v=spf1 %s -all", config.Config.EmailHostname),
		db.SoftFail: "v=spf1 ~all",
		db.HardFail: "v=spf1 -all",
		db.Neutral:  "v=spf1 ?all",
	}
	hc := HealthCheckPlugin{}
	m := &db.Message{}
	for valid, expected := range testSuite {
		m.MessageConfiguration.SPF = valid
		got := hc.generateSPFTemplate(m)
		if got != expected {
			t.Fatalf("Unexpected SPF %s response. Got %s Expected %s", valid, got, expected)
		}
	}
}

func TestGenerateDMARCTemplate(t *testing.T) {
	testSuite := map[string]string{
		db.Neutral:    "v=DMARC1; p=none; sp=none; adkim=r; aspf=r; pct=100;",
		db.Quarantine: "v=DMARC1; p=quarantine; sp=quarantine; adkim=r; aspf=r; pct=100;",
		db.Reject:     "v=DMARC1; p=reject; sp=reject; adkim=r; aspf=r; pct=100;",
	}
	hc := HealthCheckPlugin{}
	m := &db.Message{}
	for valid, expected := range testSuite {
		m.MessageConfiguration.DMARC = valid
		got := hc.generateDMARCTemplate(m)
		if got != expected {
			t.Fatalf("Unexpected DMARC %s response.\nGot %s\nExpected %s", valid, got, expected)
		}
	}
}

func TestProcessMX(t *testing.T) {
	setupConfig(t)
	testSuite := map[string]string{
		db.HardFail: dns.Fqdn(fmt.Sprintf("invalid.%s", config.Config.EmailHostname)),
		db.Pass:     dns.Fqdn(config.Config.EmailHostname),
	}
	hc := HealthCheckPlugin{}
	r := new(dns.Msg)
	w := &MockDNSResponseWriter{}
	state := request.Request{W: w, Req: r}
	for valid, expected := range testSuite {
		m := createMessage()
		m.MessageConfiguration.MX = valid
		err := db.PostMessage(m)
		if err != nil {
			t.Fatalf("Unexpected error when creating message: %v", err)
		}
		r.SetQuestion(dns.Fqdn(fmt.Sprintf("%s.%s", m.MessageID, config.Config.EmailHostname)), dns.TypeMX)
		response, err := hc.processMXRecord(state)
		if err != nil {
			t.Fatalf("Unexpected error when generating DNS response for %s: %v", valid, err)
		}
		if len(response) == 0 {
			t.Fatalf("No response provided for MX query with %s configuration", valid)
		}
		got := response[0].(*dns.MX)
		if got.Mx != expected {
			t.Fatalf("Unexpected MX %s response.\nGot %s\nExpected %s", valid, got.String(), expected)
		}
	}
}
