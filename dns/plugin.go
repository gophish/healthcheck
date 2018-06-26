package dns

import (
	"context"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/gophish/healthcheck/db"
	"github.com/miekg/dns"
)

// HealthCheckPluginName is the name of the health check plugin needed
// to implement the Handler interface
const HealthCheckPluginName = "healthcheck"

// HealthCheckPlugin is a CoreDNS plugin that emulate various email
// authentication states.
type HealthCheckPlugin struct {
	Next plugin.Handler
}

// Name implements the Handler interface.
func (hc HealthCheckPlugin) Name() string {
	return HealthCheckPluginName
}

func (hc HealthCheckPlugin) processSPFRecord(state request.Request, m *db.Message) []dns.RR {
	rr := new(dns.SPF)
	rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: state.QClass()}
	rr.Txt = []string{"SPF Response"}

	switch m.MessageConfiguration.SPF {
	case db.Pass:
		// Return a valid SPF record
	case db.SoftFail:
		// Return an invalid SPF record with the softfail directive set
	case db.HardFail:
		// Return an invalid SPF record with the hardfail directive set
	}
	return []dns.RR{rr}
}

func (hc HealthCheckPlugin) processTXTRecord(state request.Request, m *db.Message) []dns.RR {
	rrs := []dns.RR{}
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: state.QClass()}
	rr.Txt = []string{"It works!"}

	switch m.MessageConfiguration.DKIM {
	case db.Pass:
		// Return a valid DKIM public key
	case db.HardFail:
		// Return an invalid DKIM public key
	}

	spfRR := hc.processSPFRecord(state, m)
	if len(spfRR) > 0 {
		rrs = append(rrs, spfRR...)
	}

	// The DMARC policy pass/fail is determined by the SPF/DKIM configuration.
	// However, we can still set the DMARC result to none, quarantine (softfail)
	// or reject (hardfail).
	switch m.MessageConfiguration.DMARC {
	case db.SoftFail:
		// Return the DMARC policy set to quarantine
	case db.HardFail:
		// Return the DMARC policy set to reject
	}
	rrs = append(rrs, rr)
	return rrs
}

// ServeDNS retrieves the health check configuration for the requested message
// and returns an appropriate response.
func (hc HealthCheckPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	if state.QType() != dns.TypeTXT && state.QType() != dns.TypeSPF && state.QType() != dns.TypeMX {
		return plugin.NextOrFailure(hc.Name(), hc.Next, ctx, w, r)
	}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true
	a.Compress = true

	messageID := strings.Split(state.QName(), ".")[0]
	// Lookup the message
	message, err := db.GetMessage(messageID)
	if err != nil {
		return plugin.NextOrFailure(hc.Name(), hc.Next, ctx, w, r)
	}

	switch state.QType() {
	case dns.TypeTXT:
		a.Answer = hc.processTXTRecord(state, message)

	case dns.TypeSPF:
		a.Answer = hc.processSPFRecord(state, message)

	case dns.TypeMX:
		rr := new(dns.MX)
		rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeMX, Class: state.QClass()}
		rr.Preference = 10
		rr.Mx = "example.com."
		a.Answer = []dns.RR{rr}
	}

	state.SizeAndDo(a)
	w.WriteMsg(a)

	return dns.RcodeSuccess, nil
}
