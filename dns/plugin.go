package dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/gophish/healthcheck/config"
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

func (hc HealthCheckPlugin) generateSPFTemplate(message *db.Message) string {
	response := "v=spf1 "
	switch message.MessageConfiguration.SPF {
	case db.Pass:
		response += fmt.Sprintf("%s -all", config.Config.EmailHostname)
		// Return a valid SPF record
	case db.SoftFail:
		// Return an invalid SPF record with the softfail directive set
		response += "~all"
	case db.HardFail:
		// Return an invalid SPF record with the hardfail directive set
		response += "-all"
	case db.Neutral:
		response += "?all"
	}
	return response
}

func (hc HealthCheckPlugin) generateDKIMTemplate(message *db.Message) string {
	// TODO
	switch message.MessageConfiguration.DKIM {
	case db.Pass:
		// Return a valid DKIM public key
	case db.HardFail:
		// Return an invalid DKIM public key
	}
	return "DKIM"
}

func (hc HealthCheckPlugin) generateDMARCTemplate(message *db.Message) string {
	// The DMARC policy pass/fail is determined by the SPF/DKIM configuration.
	// However, we can still set the DMARC result to none, quarantine (softfail)
	// or reject (hardfail).
	response := "v=DMARC1;"
	switch message.MessageConfiguration.DMARC {
	case db.Neutral:
		// Return the DMARC policy set to none
		response += " p=none; sp=none; adkim=r; aspf=r;"
	case db.Quarantine:
		// Return the DMARC policy set to quarantine
		response += " p=quarantine; sp=quarantine; adkim=r; aspf=r;"
	case db.Reject:
		// Return the DMARC policy set to reject
		response += " p=reject; sp=reject; adkim=r; aspf=r;"
	}
	response += " pct=100;"
	return response
}

func (hc HealthCheckPlugin) processDMARCRecord(state request.Request, messageID string) ([]dns.RR, error) {
	rrs := []dns.RR{}
	message, err := db.GetMessage(messageID)
	if err != nil {
		return rrs, err
	}
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: state.QClass()}
	rr.Txt = []string{hc.generateDMARCTemplate(message)}
	rrs = append(rrs, rr)
	return rrs, nil
}

func (hc HealthCheckPlugin) processDKIMRecord(state request.Request, messageID string) ([]dns.RR, error) {
	rrs := []dns.RR{}
	message, err := db.GetMessage(messageID)
	if err != nil {
		return rrs, err
	}
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: state.QClass()}
	rr.Txt = []string{hc.generateDKIMTemplate(message)}
	rrs = append(rrs, rr)
	return rrs, nil
}

func (hc HealthCheckPlugin) processSPFRecord(state request.Request) ([]dns.RR, error) {
	rrs := []dns.RR{}
	messageID := strings.Split(state.QName(), ".")[0]
	message, err := db.GetMessage(messageID)
	if err != nil {
		return rrs, err
	}
	rr := new(dns.SPF)
	rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeSPF, Class: state.QClass()}
	rr.Txt = []string{hc.generateSPFTemplate(message)}
	rrs = append(rrs, rr)
	return rrs, nil
}

func (hc HealthCheckPlugin) processTXTRecord(state request.Request) ([]dns.RR, error) {
	rrs := []dns.RR{}
	var messageID string
	parts := strings.Split(state.QName(), ".")
	switch parts[0] {
	case config.DMARCPrefix:
		messageID = parts[1]
		return hc.processDMARCRecord(state, messageID)
	case config.DKIMPrefix:
		messageID = parts[1]
		return hc.processDKIMRecord(state, messageID)
	}
	messageID = parts[0]
	message, err := db.GetMessage(messageID)
	if err != nil {
		return rrs, err
	}
	// Process the SPF (as a TXT record) response
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: state.QClass()}
	rr.Txt = []string{hc.generateSPFTemplate(message)}
	rrs = append(rrs, rr)
	return rrs, nil
}

func (hc HealthCheckPlugin) processMXRecord(state request.Request) ([]dns.RR, error) {
	rrs := []dns.RR{}
	messageID := strings.Split(state.QName(), ".")[0]
	message, err := db.GetMessage(messageID)
	if err != nil {
		return rrs, err
	}
	rr := new(dns.MX)
	rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeMX, Class: state.QClass()}
	rr.Preference = 10
	fmt.Println(message.MessageConfiguration)
	switch message.MessageConfiguration.MX {
	case db.None:
		return rrs, nil
	case db.HardFail:
		rr.Mx = dns.Fqdn(fmt.Sprintf("invalid.%s", config.Config.EmailHostname))
	case db.Pass:
		rr.Mx = dns.Fqdn(config.Config.EmailHostname)
	}
	fmt.Println(rr.Mx)
	rrs = append(rrs, rr)
	return rrs, nil
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

	var err error

	switch state.QType() {
	case dns.TypeTXT:
		a.Answer, err = hc.processTXTRecord(state)
		if err != nil {
			return plugin.NextOrFailure(hc.Name(), hc.Next, ctx, w, r)
		}
	case dns.TypeMX:
		a.Answer, err = hc.processMXRecord(state)
		if err != nil {
			return plugin.NextOrFailure(hc.Name(), hc.Next, ctx, w, r)
		}
	// This is really only supported for odd legacy issues. Per RFC 7208, SPF
	// records must be TXT records
	case dns.TypeSPF:
		a.Answer, err = hc.processSPFRecord(state)
		if err != nil {
			return plugin.NextOrFailure(hc.Name(), hc.Next, ctx, w, r)
		}
	}

	state.SizeAndDo(a)
	w.WriteMsg(a)

	return dns.RcodeSuccess, nil
}
